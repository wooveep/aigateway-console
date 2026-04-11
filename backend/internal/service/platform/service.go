package platform

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"gopkg.in/yaml.v3"

	"github.com/alibaba/aigateway-group/aigateway-console/backend/internal/consts"
	"github.com/alibaba/aigateway-group/aigateway-console/backend/internal/model/response"
	grafanaclient "github.com/alibaba/aigateway-group/aigateway-console/backend/utility/clients/grafana"
	k8sclient "github.com/alibaba/aigateway-group/aigateway-console/backend/utility/clients/k8s"
	portaldbclient "github.com/alibaba/aigateway-group/aigateway-console/backend/utility/clients/portaldb"
)

type Service struct {
	startedAt time.Time

	k8sClient     k8sclient.Client
	grafanaClient grafanaclient.Client
	portalClient  portaldbclient.Client

	mu          sync.RWMutex
	userConfigs map[string]any
	adminUser   *response.User
}

func New(k8sClient k8sclient.Client, grafanaClient grafanaclient.Client, portalClient portaldbclient.Client) *Service {
	svc := &Service{
		startedAt:     time.Now(),
		k8sClient:     k8sClient,
		grafanaClient: grafanaClient,
		portalClient:  portalClient,
		userConfigs: map[string]any{
			"system.initialized":             false,
			"route.default.initialized":      false,
			"dashboard.builtin":              grafanaClient.IsBuiltIn(),
			"login.prompt":                   "",
			"index.redirect-target":          "/dashboard",
			"admin.password-change.disabled": false,
			"chat.enabled":                   false,
		},
	}
	return svc
}

func (s *Service) Healthz(ctx context.Context) (*response.HealthzStatus, error) {
	return &response.HealthzStatus{
		Status:        "ok",
		Service:       consts.ServiceName,
		Phase:         consts.CurrentPhase,
		LegacyBackend: consts.LegacyBackendDir,
	}, nil
}

func (s *Service) SystemInfo(ctx context.Context) (*response.SystemInfo, error) {
	return &response.SystemInfo{
		Service:         consts.ServiceName,
		Version:         g.Cfg().MustGet(ctx, "app.version", consts.DefaultBuildVersion).String(),
		Phase:           g.Cfg().MustGet(ctx, "app.phase", consts.CurrentPhase).String(),
		PreferredNaming: g.Cfg().MustGet(ctx, "naming.preferred", consts.PreferredProduct).String(),
		LegacyNaming:    g.Cfg().MustGet(ctx, "naming.legacy", consts.LegacyProduct).String(),
		LegacyBackend:   g.Cfg().MustGet(ctx, "app.legacyBackendDir", consts.LegacyBackendDir).String(),
		BusinessLines:   []string{"platform", "gateway", "portal", "jobs"},
	}, nil
}

func (s *Service) SystemConfig(ctx context.Context) (*response.SystemConfigSnapshot, error) {
	return &response.SystemConfigSnapshot{
		Module:        "github.com/alibaba/aigateway-group/aigateway-console/backend",
		ServerAddress: g.Cfg().MustGet(ctx, "server.address", consts.DefaultServerAddr).String(),
		ExplicitRenameTargets: []string{
			"service names",
			"image names",
			"documentation",
			"config keys",
			"explicit API names",
		},
		ContractDirectories: []string{
			"test/contracts/session",
			"test/contracts/system",
			"test/contracts/consumers",
			"test/contracts/org",
			"test/contracts/routes",
			"test/contracts/mcp",
			"test/contracts/ai-routes",
			"test/contracts/model-assets",
		},
		Properties: s.GetConfigs(ctx),
	}, nil
}

func (s *Service) StartedAt() time.Time {
	return s.startedAt
}

func (s *Service) GetConfigs(ctx context.Context) map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	configs := map[string]any{}
	for key, value := range s.userConfigs {
		configs[key] = value
	}
	return configs
}

func (s *Service) IsSystemInitialized(ctx context.Context) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	initialized, _ := s.userConfigs["system.initialized"].(bool)
	return initialized
}

func (s *Service) InitializeSystem(ctx context.Context, user *response.User, configs map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if initialized, _ := s.userConfigs["system.initialized"].(bool); initialized {
		return errors.New("system already initialized")
	}
	if user == nil || strings.TrimSpace(firstNonEmpty(user.Username, user.Name)) == "" || user.Password == "" {
		return errors.New("incomplete admin user")
	}

	username := firstNonEmpty(user.Username, user.Name)
	s.adminUser = &response.User{
		Name:        username,
		Username:    username,
		Password:    user.Password,
		DisplayName: firstNonEmpty(user.DisplayName, consts.DefaultAdminDisplayName),
		Type:        "admin",
	}
	for key, value := range configs {
		s.userConfigs[key] = value
	}
	s.userConfigs["system.initialized"] = true
	s.userConfigs["route.default.initialized"] = true
	return s.bootstrapDefaultResourcesLocked(ctx)
}

func (s *Service) Login(ctx context.Context, username, password string) (*response.User, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.adminUser == nil {
		return nil, "", errors.New("admin user is not initialized")
	}
	if firstNonEmpty(s.adminUser.Username, s.adminUser.Name) != username || s.adminUser.Password != password {
		return nil, "", errors.New("incorrect username or password")
	}
	token := encodeSessionToken(username, password)
	return cloneUser(s.adminUser), token, nil
}

func (s *Service) ValidateSessionToken(ctx context.Context, token string) (*response.User, error) {
	if token == "" {
		return nil, errors.New("missing session token")
	}
	username, password, err := decodeSessionToken(token)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.adminUser == nil {
		return nil, errors.New("admin user is not initialized")
	}
	if firstNonEmpty(s.adminUser.Username, s.adminUser.Name) != username || s.adminUser.Password != password {
		return nil, errors.New("invalid session token")
	}
	return cloneUser(s.adminUser), nil
}

func (s *Service) ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if disabled, _ := s.userConfigs["admin.password-change.disabled"].(bool); disabled {
		return errors.New("password change is disabled")
	}
	if s.adminUser == nil {
		return errors.New("admin user is not initialized")
	}
	if firstNonEmpty(s.adminUser.Username, s.adminUser.Name) != username {
		return errors.New("unknown username")
	}
	if s.adminUser.Password != oldPassword {
		return errors.New("incorrect old password")
	}
	s.adminUser.Password = newPassword
	return nil
}

func (s *Service) DashboardInfo(ctx context.Context, dashboardType string) (*response.DashboardInfo, error) {
	info := &response.DashboardInfo{
		BuiltIn: s.grafanaClient.IsBuiltIn(),
		URL:     "",
	}
	if info.BuiltIn {
		info.URL = s.grafanaClient.BaseURL()
		info.UID = dashboardUID(dashboardType)
		return info, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	urlKey := dashboardURLKey(dashboardType)
	if value, ok := s.userConfigs[urlKey].(string); ok {
		info.URL = value
	}
	return info, nil
}

func (s *Service) InitializeDashboard(ctx context.Context, dashboardType string, force bool) (*response.DashboardInfo, error) {
	if !s.grafanaClient.IsBuiltIn() {
		return nil, errors.New("no built-in dashboard is available")
	}
	return s.DashboardInfo(ctx, dashboardType)
}

func (s *Service) SetDashboardURL(ctx context.Context, dashboardType, url string) (*response.DashboardInfo, error) {
	if s.grafanaClient.IsBuiltIn() {
		return nil, errors.New("manual dashboard configuration is disabled")
	}
	s.mu.Lock()
	s.userConfigs[dashboardURLKey(dashboardType)] = url
	s.mu.Unlock()
	return s.DashboardInfo(ctx, dashboardType)
}

func (s *Service) BuildDashboardConfigData(ctx context.Context, dashboardType, datasourceUID string) (string, error) {
	resourceName := strings.ToLower(strings.TrimSpace(dashboardType))
	if resourceName == "" {
		resourceName = "main"
	}
	path := filepath.Join("resource", "dashboard", resourceName+".json")
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(string(content), "${datasource.id}", datasourceUID), nil
}

func (s *Service) NativeDashboard(ctx context.Context, dashboardType string, from, to int64, gateway, namespace string) (*response.NativeDashboardData, error) {
	now := time.Now().UnixMilli()
	if from == 0 {
		from = now - 5*60*1000
	}
	if to == 0 {
		to = now
	}

	data := &response.NativeDashboardData{
		Title:          strings.ToUpper(firstNonEmpty(dashboardType, "MAIN")) + " Dashboard",
		Type:           strings.ToUpper(firstNonEmpty(dashboardType, "MAIN")),
		From:           from,
		To:             to,
		DefaultRangeMS: 5 * 60 * 1000,
		Rows:           []response.NativeDashboardRow{},
	}
	data.Variables.Gateway = response.NativeDashboardVariableState{Value: gateway, Options: []string{gateway}}
	data.Variables.Namespace = response.NativeDashboardVariableState{Value: firstNonEmpty(namespace, "aigateway-system"), Options: []string{"aigateway-system"}}
	return data, nil
}

func (s *Service) GetAIGatewayConfig(ctx context.Context) (string, error) {
	data, err := s.readMigratedConfigMap(ctx)
	if err != nil {
		return "", err
	}
	return k8sclient.RenderConfigMapYAML(consts.DefaultConfigMapName, data)
}

func (s *Service) SetAIGatewayConfig(ctx context.Context, raw string) (string, error) {
	data, err := k8sclient.ParseConfigMapYAML(raw)
	if err != nil {
		return "", err
	}
	migrateLegacyConfigData(data)
	requiredKeys := []string{"aigateway", "mesh", "meshNetworks"}
	for _, key := range requiredKeys {
		if strings.TrimSpace(data[key]) == "" {
			return "", fmt.Errorf("config map data key %s has an empty value", key)
		}
		var node map[string]any
		if err := yaml.Unmarshal([]byte(data[key]), &node); err != nil {
			return "", fmt.Errorf("invalid YAML data for key %s: %w", key, err)
		}
	}
	if data["resourceVersion"] == "" {
		current, err := s.readMigratedConfigMap(ctx)
		if err == nil {
			data["resourceVersion"] = current["resourceVersion"]
		}
	}
	if data["resourceVersion"] == "" {
		data["resourceVersion"] = "1"
	}
	if err := s.k8sClient.UpsertConfigMap(ctx, consts.DefaultConfigMapName, data); err != nil {
		return "", err
	}
	return k8sclient.RenderConfigMapYAML(consts.DefaultConfigMapName, data)
}

func (s *Service) readMigratedConfigMap(ctx context.Context) (map[string]string, error) {
	data, err := s.k8sClient.ReadConfigMap(ctx, consts.DefaultConfigMapName)
	if err != nil {
		return nil, err
	}
	if migrateLegacyConfigData(data) {
		if err := s.k8sClient.UpsertConfigMap(ctx, consts.DefaultConfigMapName, data); err != nil {
			return nil, err
		}
	}
	return data, nil
}

func migrateLegacyConfigData(data map[string]string) bool {
	if data == nil {
		return false
	}
	if strings.TrimSpace(data["aigateway"]) != "" {
		return false
	}
	legacy := strings.TrimSpace(data["higress"])
	if legacy == "" {
		return false
	}
	data["aigateway"] = data["higress"]
	delete(data, "higress")
	return true
}

func (s *Service) bootstrapDefaultResourcesLocked(ctx context.Context) error {
	if err := s.k8sClient.UpsertSecret(ctx, consts.DefaultSecretName, map[string]string{
		"adminUsername":    firstNonEmpty(s.adminUser.Username, s.adminUser.Name),
		"adminPassword":    s.adminUser.Password,
		"adminDisplayName": s.adminUser.DisplayName,
	}); err != nil {
		return err
	}
	if _, err := s.k8sClient.UpsertResource(ctx, "tls-certificates", consts.DefaultTLSCertificateName, map[string]any{
		"name":          consts.DefaultTLSCertificateName,
		"cert":          "placeholder-cert",
		"key":           "placeholder-key",
		"domains":       []string{consts.DefaultTLSCertificateHost},
		"validityStart": time.Now().UTC().Format(time.RFC3339),
		"validityEnd":   time.Now().UTC().Add(365 * 24 * time.Hour).Format(time.RFC3339),
	}); err != nil {
		return err
	}
	if _, err := s.k8sClient.UpsertResource(ctx, "domains", consts.DefaultDomainName, map[string]any{
		"name":           consts.DefaultDomainName,
		"certIdentifier": consts.DefaultTLSCertificateName,
		"enableHttps":    "on",
	}); err != nil {
		return err
	}
	_, err := s.k8sClient.UpsertResource(ctx, "routes", consts.DefaultRouteName, map[string]any{
		"name":    consts.DefaultRouteName,
		"domains": []string{consts.DefaultDomainName},
		"path": map[string]any{
			"matchType":  "EQUAL",
			"matchValue": "/",
		},
		"services": []map[string]any{{
			"name": consts.DefaultConsoleServiceHost,
			"port": consts.DefaultConsoleServicePort,
		}},
		"rewrite": map[string]any{
			"enabled": true,
			"path":    "/landing",
		},
	})
	return err
}

func dashboardUID(dashboardType string) string {
	switch strings.ToUpper(strings.TrimSpace(dashboardType)) {
	case "AI":
		return consts.DefaultDashboardUIDAI
	case "LOG":
		return consts.DefaultDashboardUIDLog
	default:
		return consts.DefaultDashboardUIDMain
	}
}

func dashboardURLKey(dashboardType string) string {
	switch strings.ToUpper(strings.TrimSpace(dashboardType)) {
	case "AI":
		return "dashboard.url.ai"
	case "LOG":
		return "dashboard.url.log"
	default:
		return "dashboard.url"
	}
}

func encodeSessionToken(username, password string) string {
	raw := fmt.Sprintf("%s:%s:%d", username, password, time.Now().UnixMilli())
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func decodeSessionToken(token string) (string, string, error) {
	rawBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", "", err
	}
	parts := strings.SplitN(string(rawBytes), ":", 3)
	if len(parts) < 2 {
		return "", "", errors.New("invalid token format")
	}
	return parts[0], parts[1], nil
}

func cloneUser(user *response.User) *response.User {
	if user == nil {
		return nil
	}
	cp := *user
	cp.Password = ""
	return &cp
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
