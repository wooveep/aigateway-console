package portal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	portalshared "higress-portal-backend/schema/shared"
)

const portalSSOProviderTypeOIDC = "oidc"

type PortalSSOClaimMapping struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
}

type PortalSSOConfigRecord struct {
	Enabled                bool                  `json:"enabled"`
	ProviderType           string                `json:"providerType"`
	DisplayName            string                `json:"displayName"`
	IssuerURL              string                `json:"issuerUrl"`
	ClientID               string                `json:"clientId"`
	ClientSecret           string                `json:"clientSecret,omitempty"`
	ClientSecretMasked     string                `json:"clientSecretMasked,omitempty"`
	ClientSecretConfigured bool                  `json:"clientSecretConfigured"`
	Scopes                 []string              `json:"scopes"`
	ClaimMapping           PortalSSOClaimMapping `json:"claimMapping"`
	UpdatedBy              string                `json:"updatedBy,omitempty"`
	UpdatedAt              *time.Time            `json:"updatedAt,omitempty"`
}

type portalSSOConfigStored struct {
	Enabled               bool
	ProviderType          string
	DisplayName           string
	IssuerURL             string
	ClientID              string
	ClientSecretEncrypted string
	ScopesJSON            string
	ClaimMappingJSON      string
	UpdatedBy             string
	UpdatedAt             *time.Time
}

type portalSSODiscoveryDocument struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
}

func (s *Service) GetPortalSSOConfig(ctx context.Context) (*PortalSSOConfigRecord, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}

	stored, err := newPortalStore(db, s.client.Driver()).getPortalSSOConfig(ctx)
	if err != nil {
		return nil, err
	}
	return s.portalSSOConfigFromStored(stored)
}

func (s *Service) SavePortalSSOConfig(ctx context.Context, mutation PortalSSOConfigRecord, updatedBy string) (*PortalSSOConfigRecord, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}

	store := newPortalStore(db, s.client.Driver())
	existing, err := store.getPortalSSOConfig(ctx)
	if err != nil {
		return nil, err
	}

	normalized := normalizePortalSSOConfigRecord(mutation)
	if strings.TrimSpace(normalized.IssuerURL) == "" {
		return nil, fmt.Errorf("issuerUrl cannot be blank")
	}
	if strings.TrimSpace(normalized.ClientID) == "" {
		return nil, fmt.Errorf("clientId cannot be blank")
	}
	if err := validatePortalSSODiscovery(ctx, normalized.IssuerURL); err != nil {
		return nil, err
	}

	encryptedSecret := ""
	if existing != nil {
		encryptedSecret = strings.TrimSpace(existing.ClientSecretEncrypted)
	}
	if secret := strings.TrimSpace(mutation.ClientSecret); secret != "" {
		encryptedSecret, err = portalshared.EncryptPortalSSOSecret(secret)
		if err != nil {
			return nil, fmt.Errorf("encrypt clientSecret failed: %w", err)
		}
	}
	if encryptedSecret == "" {
		return nil, fmt.Errorf("clientSecret cannot be blank")
	}

	scopesJSON, err := json.Marshal(normalized.Scopes)
	if err != nil {
		return nil, fmt.Errorf("marshal scopes failed: %w", err)
	}
	claimMappingJSON, err := json.Marshal(normalized.ClaimMapping)
	if err != nil {
		return nil, fmt.Errorf("marshal claimMapping failed: %w", err)
	}

	if err := store.upsertPortalSSOConfig(ctx, portalSSOConfigStored{
		Enabled:               normalized.Enabled,
		ProviderType:          normalized.ProviderType,
		DisplayName:           normalized.DisplayName,
		IssuerURL:             normalized.IssuerURL,
		ClientID:              normalized.ClientID,
		ClientSecretEncrypted: encryptedSecret,
		ScopesJSON:            string(scopesJSON),
		ClaimMappingJSON:      string(claimMappingJSON),
		UpdatedBy:             firstNonEmpty(strings.TrimSpace(updatedBy), "system"),
	}); err != nil {
		return nil, err
	}
	return s.GetPortalSSOConfig(ctx)
}

func (s *Service) portalSSOConfigFromStored(stored *portalSSOConfigStored) (*PortalSSOConfigRecord, error) {
	cfg := &PortalSSOConfigRecord{
		Enabled:      false,
		ProviderType: portalSSOProviderTypeOIDC,
		DisplayName:  "企业 SSO 登录",
		Scopes:       []string{"openid", "profile", "email"},
		ClaimMapping: defaultPortalSSOClaimMapping(),
	}
	if stored == nil {
		return cfg, nil
	}

	cfg.Enabled = stored.Enabled
	cfg.ProviderType = firstNonEmpty(strings.TrimSpace(stored.ProviderType), portalSSOProviderTypeOIDC)
	cfg.DisplayName = firstNonEmpty(strings.TrimSpace(stored.DisplayName), cfg.DisplayName)
	cfg.IssuerURL = strings.TrimSpace(stored.IssuerURL)
	cfg.ClientID = strings.TrimSpace(stored.ClientID)
	cfg.UpdatedBy = strings.TrimSpace(stored.UpdatedBy)
	cfg.UpdatedAt = stored.UpdatedAt
	if scopes := parsePortalSSOScopes(stored.ScopesJSON); len(scopes) > 0 {
		cfg.Scopes = scopes
	}
	if mapping := parsePortalSSOClaimMapping(stored.ClaimMappingJSON); mapping != (PortalSSOClaimMapping{}) {
		cfg.ClaimMapping = mapping
	}

	secret, err := portalshared.DecryptPortalSSOSecret(stored.ClientSecretEncrypted)
	if err != nil {
		return nil, fmt.Errorf("decrypt clientSecret failed: %w", err)
	}
	cfg.ClientSecretConfigured = strings.TrimSpace(secret) != ""
	cfg.ClientSecretMasked = portalshared.MaskPortalSSOSecret(secret)
	return cfg, nil
}

func validatePortalSSODiscovery(ctx context.Context, issuerURL string) error {
	discoveryURL, err := portalSSODiscoveryURL(issuerURL)
	if err != nil {
		return fmt.Errorf("issuerUrl is invalid: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return fmt.Errorf("build discovery request failed: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("oidc discovery request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("oidc discovery request failed with status %d", resp.StatusCode)
	}

	var discovery portalSSODiscoveryDocument
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return fmt.Errorf("oidc discovery response is invalid: %w", err)
	}
	if !sameOIDCIssuer(discovery.Issuer, issuerURL) {
		return fmt.Errorf("oidc discovery issuer mismatch")
	}
	if strings.TrimSpace(discovery.AuthorizationEndpoint) == "" {
		return fmt.Errorf("oidc discovery authorization_endpoint is missing")
	}
	if strings.TrimSpace(discovery.TokenEndpoint) == "" {
		return fmt.Errorf("oidc discovery token_endpoint is missing")
	}
	if strings.TrimSpace(discovery.JWKSURI) == "" {
		return fmt.Errorf("oidc discovery jwks_uri is missing")
	}
	return nil
}

func portalSSODiscoveryURL(issuerURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(issuerURL))
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("issuer must include scheme and host")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/.well-known/openid-configuration"
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String(), nil
}

func sameOIDCIssuer(actual, expected string) bool {
	return strings.TrimRight(strings.TrimSpace(actual), "/") == strings.TrimRight(strings.TrimSpace(expected), "/")
}

func normalizePortalSSOConfigRecord(input PortalSSOConfigRecord) PortalSSOConfigRecord {
	scopes := make([]string, 0, len(input.Scopes))
	seenScopes := map[string]struct{}{}
	for _, item := range input.Scopes {
		scope := strings.TrimSpace(item)
		if scope == "" {
			continue
		}
		if _, exists := seenScopes[scope]; exists {
			continue
		}
		seenScopes[scope] = struct{}{}
		scopes = append(scopes, scope)
	}
	if len(scopes) == 0 {
		scopes = []string{"openid", "profile", "email"}
	}
	if _, ok := seenScopes["openid"]; !ok {
		scopes = append([]string{"openid"}, scopes...)
	}

	return PortalSSOConfigRecord{
		Enabled:      input.Enabled,
		ProviderType: portalSSOProviderTypeOIDC,
		DisplayName:  firstNonEmpty(strings.TrimSpace(input.DisplayName), "企业 SSO 登录"),
		IssuerURL:    strings.TrimRight(strings.TrimSpace(input.IssuerURL), "/"),
		ClientID:     strings.TrimSpace(input.ClientID),
		Scopes:       scopes,
		ClaimMapping: normalizePortalSSOClaimMapping(input.ClaimMapping),
	}
}

func defaultPortalSSOClaimMapping() PortalSSOClaimMapping {
	return PortalSSOClaimMapping{
		Email:       "email",
		DisplayName: "name",
		Username:    "preferred_username",
	}
}

func normalizePortalSSOClaimMapping(mapping PortalSSOClaimMapping) PortalSSOClaimMapping {
	defaults := defaultPortalSSOClaimMapping()
	return PortalSSOClaimMapping{
		Email:       firstNonEmpty(strings.TrimSpace(mapping.Email), defaults.Email),
		DisplayName: firstNonEmpty(strings.TrimSpace(mapping.DisplayName), defaults.DisplayName),
		Username:    firstNonEmpty(strings.TrimSpace(mapping.Username), defaults.Username),
	}
}

func parsePortalSSOScopes(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(trimmed), &items); err != nil {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		if scope := strings.TrimSpace(item); scope != "" {
			result = append(result, scope)
		}
	}
	return result
}

func parsePortalSSOClaimMapping(raw string) PortalSSOClaimMapping {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return PortalSSOClaimMapping{}
	}
	var mapping PortalSSOClaimMapping
	if err := json.Unmarshal([]byte(trimmed), &mapping); err != nil {
		return PortalSSOClaimMapping{}
	}
	return normalizePortalSSOClaimMapping(mapping)
}
