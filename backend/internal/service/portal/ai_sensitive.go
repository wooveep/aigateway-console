package portal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type AISensitiveMenuState struct {
	Enabled           bool `json:"enabled"`
	EnabledRouteCount int  `json:"enabledRouteCount"`
}

type AISensitiveDetectRule struct {
	ID          int64      `json:"id,omitempty"`
	Pattern     string     `json:"pattern"`
	MatchType   string     `json:"matchType"`
	Description string     `json:"description,omitempty"`
	Priority    int        `json:"priority,omitempty"`
	Enabled     bool       `json:"enabled"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

type AISensitiveReplaceRule struct {
	ID           int64      `json:"id,omitempty"`
	Pattern      string     `json:"pattern"`
	ReplaceType  string     `json:"replaceType"`
	ReplaceValue string     `json:"replaceValue,omitempty"`
	Restore      bool       `json:"restore,omitempty"`
	Description  string     `json:"description,omitempty"`
	Priority     int        `json:"priority,omitempty"`
	Enabled      bool       `json:"enabled"`
	CreatedAt    *time.Time `json:"createdAt,omitempty"`
	UpdatedAt    *time.Time `json:"updatedAt,omitempty"`
}

type AISensitiveBlockAudit struct {
	ID                int64      `json:"id"`
	RequestID         string     `json:"requestId,omitempty"`
	RouteName         string     `json:"routeName,omitempty"`
	ConsumerName      string     `json:"consumerName,omitempty"`
	DisplayName       string     `json:"displayName,omitempty"`
	BlockedAt         *time.Time `json:"blockedAt,omitempty"`
	BlockedBy         string     `json:"blockedBy,omitempty"`
	RequestPhase      string     `json:"requestPhase,omitempty"`
	BlockedReasonJSON string     `json:"blockedReasonJson,omitempty"`
	MatchType         string     `json:"matchType,omitempty"`
	MatchedRule       string     `json:"matchedRule,omitempty"`
	MatchedExcerpt    string     `json:"matchedExcerpt,omitempty"`
	ProviderID        *int64     `json:"providerId,omitempty"`
	CostUSD           string     `json:"costUsd,omitempty"`
}

type AISensitiveSystemConfig struct {
	SystemDenyEnabled bool       `json:"systemDenyEnabled"`
	DictionaryText    string     `json:"dictionaryText"`
	UpdatedBy         string     `json:"updatedBy,omitempty"`
	UpdatedAt         *time.Time `json:"updatedAt,omitempty"`
}

type AISensitiveStatus struct {
	DBEnabled                 bool       `json:"dbEnabled"`
	DetectRuleCount           int        `json:"detectRuleCount"`
	ReplaceRuleCount          int        `json:"replaceRuleCount"`
	AuditRecordCount          int        `json:"auditRecordCount"`
	SystemDenyEnabled         bool       `json:"systemDenyEnabled"`
	SystemDictionaryWordCount int        `json:"systemDictionaryWordCount"`
	SystemDictionaryUpdatedAt *time.Time `json:"systemDictionaryUpdatedAt,omitempty"`
	ProjectedInstanceCount    int        `json:"projectedInstanceCount"`
	LastReconciledAt          *time.Time `json:"lastReconciledAt,omitempty"`
	LastMigratedAt            *time.Time `json:"lastMigratedAt,omitempty"`
	LastError                 string     `json:"lastError,omitempty"`
}

type AISensitiveAuditQuery struct {
	ConsumerName string
	DisplayName  string
	RouteName    string
	MatchType    string
	StartTime    string
	EndTime      string
	Limit        int
}

func (s *Service) GetAISensitiveMenuState(ctx context.Context) (*AISensitiveMenuState, error) {
	status, err := s.GetAISensitiveStatus(ctx)
	if err != nil {
		return nil, err
	}
	return &AISensitiveMenuState{
		Enabled:           status.DetectRuleCount > 0 || status.ReplaceRuleCount > 0 || status.ProjectedInstanceCount > 0 || status.SystemDenyEnabled,
		EnabledRouteCount: status.ProjectedInstanceCount,
	}, nil
}

func (s *Service) ListAISensitiveDetectRules(ctx context.Context) ([]AISensitiveDetectRule, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id, pattern, match_type, description, priority, enabled, created_at, updated_at
		FROM portal_ai_sensitive_detect_rule
		ORDER BY priority DESC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AISensitiveDetectRule, 0)
	for rows.Next() {
		item, err := scanAISensitiveDetectRule(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) SaveAISensitiveDetectRule(ctx context.Context, rule AISensitiveDetectRule) (*AISensitiveDetectRule, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}
	rule.Pattern = strings.TrimSpace(rule.Pattern)
	rule.MatchType = strings.ToLower(strings.TrimSpace(rule.MatchType))
	if rule.Pattern == "" {
		return nil, errors.New("pattern cannot be blank")
	}
	if rule.MatchType == "" {
		rule.MatchType = "contains"
	}
	if rule.ID == 0 && !rule.Enabled {
		rule.Enabled = true
	}
	enabled := 0
	if rule.Enabled {
		enabled = 1
	}

	if rule.ID > 0 {
		_, err = db.ExecContext(ctx, `
			UPDATE portal_ai_sensitive_detect_rule
			SET pattern = ?, match_type = ?, description = ?, priority = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?`,
			rule.Pattern,
			rule.MatchType,
			trimOrNil(rule.Description),
			rule.Priority,
			enabled,
			rule.ID,
		)
	} else {
		result, execErr := db.ExecContext(ctx, `
			INSERT INTO portal_ai_sensitive_detect_rule (pattern, match_type, description, priority, enabled)
			VALUES (?, ?, ?, ?, ?)`,
			rule.Pattern,
			rule.MatchType,
			trimOrNil(rule.Description),
			rule.Priority,
			enabled,
		)
		err = execErr
		if execErr == nil {
			rule.ID, _ = result.LastInsertId()
		}
	}
	if err != nil {
		return nil, err
	}
	return s.getAISensitiveDetectRule(ctx, rule.ID)
}

func (s *Service) DeleteAISensitiveDetectRule(ctx context.Context, id int64) error {
	db, err := s.db(ctx)
	if err != nil {
		return err
	}
	result, err := db.ExecContext(ctx, `DELETE FROM portal_ai_sensitive_detect_rule WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("detect rule not found: %d", id)
	}
	return nil
}

func (s *Service) ListAISensitiveReplaceRules(ctx context.Context) ([]AISensitiveReplaceRule, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id, pattern, replace_type, replace_value, restore, description, priority, enabled, created_at, updated_at
		FROM portal_ai_sensitive_replace_rule
		ORDER BY priority DESC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AISensitiveReplaceRule, 0)
	for rows.Next() {
		item, err := scanAISensitiveReplaceRule(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) SaveAISensitiveReplaceRule(ctx context.Context, rule AISensitiveReplaceRule) (*AISensitiveReplaceRule, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}
	rule.Pattern = strings.TrimSpace(rule.Pattern)
	rule.ReplaceType = strings.ToLower(strings.TrimSpace(rule.ReplaceType))
	if rule.Pattern == "" {
		return nil, errors.New("pattern cannot be blank")
	}
	if rule.ReplaceType == "" {
		rule.ReplaceType = "replace"
	}
	if rule.ID == 0 && !rule.Enabled {
		rule.Enabled = true
	}
	enabled := 0
	if rule.Enabled {
		enabled = 1
	}
	restore := 0
	if rule.Restore {
		restore = 1
	}

	if rule.ID > 0 {
		_, err = db.ExecContext(ctx, `
			UPDATE portal_ai_sensitive_replace_rule
			SET pattern = ?, replace_type = ?, replace_value = ?, restore = ?, description = ?, priority = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?`,
			rule.Pattern,
			rule.ReplaceType,
			trimOrNil(rule.ReplaceValue),
			restore,
			trimOrNil(rule.Description),
			rule.Priority,
			enabled,
			rule.ID,
		)
	} else {
		result, execErr := db.ExecContext(ctx, `
			INSERT INTO portal_ai_sensitive_replace_rule (
				pattern, replace_type, replace_value, restore, description, priority, enabled
			) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			rule.Pattern,
			rule.ReplaceType,
			trimOrNil(rule.ReplaceValue),
			restore,
			trimOrNil(rule.Description),
			rule.Priority,
			enabled,
		)
		err = execErr
		if execErr == nil {
			rule.ID, _ = result.LastInsertId()
		}
	}
	if err != nil {
		return nil, err
	}
	return s.getAISensitiveReplaceRule(ctx, rule.ID)
}

func (s *Service) DeleteAISensitiveReplaceRule(ctx context.Context, id int64) error {
	db, err := s.db(ctx)
	if err != nil {
		return err
	}
	result, err := db.ExecContext(ctx, `DELETE FROM portal_ai_sensitive_replace_rule WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("replace rule not found: %d", id)
	}
	return nil
}

func (s *Service) ListAISensitiveAudits(ctx context.Context, query AISensitiveAuditQuery) ([]AISensitiveBlockAudit, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}
	statement := `
		SELECT id, request_id, route_name, consumer_name, display_name, blocked_at, blocked_by, request_phase,
			blocked_reason_json, match_type, matched_rule, matched_excerpt, provider_id, cost_usd
		FROM portal_ai_sensitive_block_audit
		WHERE 1 = 1`
	args := make([]any, 0)
	if strings.TrimSpace(query.ConsumerName) != "" {
		statement += ` AND consumer_name = ?`
		args = append(args, strings.TrimSpace(query.ConsumerName))
	}
	if strings.TrimSpace(query.DisplayName) != "" {
		statement += ` AND display_name = ?`
		args = append(args, strings.TrimSpace(query.DisplayName))
	}
	if strings.TrimSpace(query.RouteName) != "" {
		statement += ` AND route_name = ?`
		args = append(args, strings.TrimSpace(query.RouteName))
	}
	if strings.TrimSpace(query.MatchType) != "" {
		statement += ` AND match_type = ?`
		args = append(args, strings.TrimSpace(query.MatchType))
	}
	if strings.TrimSpace(query.StartTime) != "" {
		statement += ` AND blocked_at >= ?`
		args = append(args, strings.TrimSpace(query.StartTime))
	}
	if strings.TrimSpace(query.EndTime) != "" {
		statement += ` AND blocked_at <= ?`
		args = append(args, strings.TrimSpace(query.EndTime))
	}
	statement += ` ORDER BY blocked_at DESC`
	limit := query.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	statement += fmt.Sprintf(" LIMIT %d", limit)

	rows, err := db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AISensitiveBlockAudit, 0)
	for rows.Next() {
		item, err := scanAISensitiveAudit(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) GetAISensitiveSystemConfig(ctx context.Context) (*AISensitiveSystemConfig, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}
	config := &AISensitiveSystemConfig{}
	var (
		enabledInt int
		updatedBy  sql.NullString
		updatedAt  sql.NullTime
	)
	err = db.QueryRowContext(ctx, `
		SELECT system_deny_enabled, dictionary_text, updated_by, updated_at
		FROM portal_ai_sensitive_system_config
		WHERE config_key = 'default'`,
	).Scan(&enabledInt, &config.DictionaryText, &updatedBy, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return config, nil
		}
		return nil, err
	}
	config.SystemDenyEnabled = enabledInt > 0
	config.UpdatedBy = updatedBy.String
	if updatedAt.Valid {
		config.UpdatedAt = &updatedAt.Time
	}
	return config, nil
}

func (s *Service) SaveAISensitiveSystemConfig(ctx context.Context, config AISensitiveSystemConfig) (*AISensitiveSystemConfig, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}
	enabledInt := 0
	if config.SystemDenyEnabled {
		enabledInt = 1
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO portal_ai_sensitive_system_config (config_key, system_deny_enabled, dictionary_text, updated_by)
		VALUES ('default', ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			system_deny_enabled = VALUES(system_deny_enabled),
			dictionary_text = VALUES(dictionary_text),
			updated_by = VALUES(updated_by),
			updated_at = CURRENT_TIMESTAMP`,
		enabledInt,
		config.DictionaryText,
		nullIfEmpty(config.UpdatedBy),
	); err != nil {
		return nil, err
	}
	return s.GetAISensitiveSystemConfig(ctx)
}

func (s *Service) GetAISensitiveStatus(ctx context.Context) (*AISensitiveStatus, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}
	config, err := s.GetAISensitiveSystemConfig(ctx)
	if err != nil {
		return nil, err
	}
	status := &AISensitiveStatus{
		DBEnabled:                 true,
		DetectRuleCount:           queryCount(ctx, db, `SELECT COUNT(1) FROM portal_ai_sensitive_detect_rule`),
		ReplaceRuleCount:          queryCount(ctx, db, `SELECT COUNT(1) FROM portal_ai_sensitive_replace_rule`),
		AuditRecordCount:          queryCount(ctx, db, `SELECT COUNT(1) FROM portal_ai_sensitive_block_audit`),
		SystemDenyEnabled:         config.SystemDenyEnabled,
		SystemDictionaryWordCount: countDictionaryLines(config.DictionaryText),
		SystemDictionaryUpdatedAt: config.UpdatedAt,
		LastMigratedAt:            config.UpdatedAt,
	}

	if s.k8sClient != nil {
		items, err := s.k8sClient.ListResources(ctx, "ai-sensitive-projections")
		if err == nil {
			status.ProjectedInstanceCount = len(items)
		}
		item, err := s.k8sClient.GetResource(ctx, "ai-sensitive-projections", "default")
		if err == nil {
			status.LastReconciledAt = parseRFC3339Pointer(fmt.Sprint(item["projectedAt"]))
			status.LastError = strings.TrimSpace(fmt.Sprint(item["lastError"]))
		}
	}
	return status, nil
}

func (s *Service) ReconcileAISensitive(ctx context.Context) (*AISensitiveStatus, error) {
	if _, err := s.db(ctx); err != nil {
		return nil, err
	}
	if s.k8sClient == nil {
		return nil, errors.New("ai sensitive projection requires k8s client")
	}
	detectRules, err := s.ListAISensitiveDetectRules(ctx)
	if err != nil {
		return nil, err
	}
	replaceRules, err := s.ListAISensitiveReplaceRules(ctx)
	if err != nil {
		return nil, err
	}
	config, err := s.GetAISensitiveSystemConfig(ctx)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{
		"name":         "default",
		"detectRules":  detectRules,
		"replaceRules": replaceRules,
		"systemConfig": config,
		"projectedAt":  time.Now().UTC().Format(time.RFC3339),
	}
	if _, err := s.k8sClient.UpsertResource(ctx, "ai-sensitive-projections", "default", payload); err != nil {
		return nil, err
	}
	return s.GetAISensitiveStatus(ctx)
}

func (s *Service) getAISensitiveDetectRule(ctx context.Context, id int64) (*AISensitiveDetectRule, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}
	row := db.QueryRowContext(ctx, `
		SELECT id, pattern, match_type, description, priority, enabled, created_at, updated_at
		FROM portal_ai_sensitive_detect_rule
		WHERE id = ?`, id)
	item, err := scanAISensitiveDetectRule(row)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Service) getAISensitiveReplaceRule(ctx context.Context, id int64) (*AISensitiveReplaceRule, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}
	row := db.QueryRowContext(ctx, `
		SELECT id, pattern, replace_type, replace_value, restore, description, priority, enabled, created_at, updated_at
		FROM portal_ai_sensitive_replace_rule
		WHERE id = ?`, id)
	item, err := scanAISensitiveReplaceRule(row)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func scanAISensitiveDetectRule(scanner interface{ Scan(...any) error }) (AISensitiveDetectRule, error) {
	var (
		item      AISensitiveDetectRule
		enabled   int
		desc      sql.NullString
		createdAt sql.NullTime
		updatedAt sql.NullTime
	)
	if err := scanner.Scan(
		&item.ID,
		&item.Pattern,
		&item.MatchType,
		&desc,
		&item.Priority,
		&enabled,
		&createdAt,
		&updatedAt,
	); err != nil {
		return AISensitiveDetectRule{}, err
	}
	item.Enabled = enabled > 0
	item.Description = desc.String
	if createdAt.Valid {
		item.CreatedAt = &createdAt.Time
	}
	if updatedAt.Valid {
		item.UpdatedAt = &updatedAt.Time
	}
	return item, nil
}

func scanAISensitiveReplaceRule(scanner interface{ Scan(...any) error }) (AISensitiveReplaceRule, error) {
	var (
		item         AISensitiveReplaceRule
		restoreInt   int
		enabledInt   int
		replaceValue sql.NullString
		desc         sql.NullString
		createdAt    sql.NullTime
		updatedAt    sql.NullTime
	)
	if err := scanner.Scan(
		&item.ID,
		&item.Pattern,
		&item.ReplaceType,
		&replaceValue,
		&restoreInt,
		&desc,
		&item.Priority,
		&enabledInt,
		&createdAt,
		&updatedAt,
	); err != nil {
		return AISensitiveReplaceRule{}, err
	}
	item.ReplaceValue = replaceValue.String
	item.Restore = restoreInt > 0
	item.Enabled = enabledInt > 0
	item.Description = desc.String
	if createdAt.Valid {
		item.CreatedAt = &createdAt.Time
	}
	if updatedAt.Valid {
		item.UpdatedAt = &updatedAt.Time
	}
	return item, nil
}

func scanAISensitiveAudit(scanner interface{ Scan(...any) error }) (AISensitiveBlockAudit, error) {
	var (
		item              AISensitiveBlockAudit
		requestID         sql.NullString
		routeName         sql.NullString
		consumerName      sql.NullString
		displayName       sql.NullString
		blockedAt         sql.NullTime
		blockedBy         sql.NullString
		requestPhase      sql.NullString
		blockedReasonJSON sql.NullString
		matchType         sql.NullString
		matchedRule       sql.NullString
		matchedExcerpt    sql.NullString
		providerID        sql.NullInt64
		costUSD           sql.NullString
	)
	if err := scanner.Scan(
		&item.ID,
		&requestID,
		&routeName,
		&consumerName,
		&displayName,
		&blockedAt,
		&blockedBy,
		&requestPhase,
		&blockedReasonJSON,
		&matchType,
		&matchedRule,
		&matchedExcerpt,
		&providerID,
		&costUSD,
	); err != nil {
		return AISensitiveBlockAudit{}, err
	}
	item.RequestID = requestID.String
	item.RouteName = routeName.String
	item.ConsumerName = consumerName.String
	item.DisplayName = displayName.String
	item.BlockedBy = blockedBy.String
	item.RequestPhase = requestPhase.String
	item.BlockedReasonJSON = blockedReasonJSON.String
	item.MatchType = matchType.String
	item.MatchedRule = matchedRule.String
	item.MatchedExcerpt = matchedExcerpt.String
	if blockedAt.Valid {
		item.BlockedAt = &blockedAt.Time
	}
	if providerID.Valid {
		value := providerID.Int64
		item.ProviderID = &value
	}
	item.CostUSD = costUSD.String
	return item, nil
}

func queryCount(ctx context.Context, db *sql.DB, statement string, args ...any) int {
	var count int
	_ = db.QueryRowContext(ctx, statement, args...).Scan(&count)
	return count
}

func parseRFC3339Pointer(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return &parsed
}

func countDictionaryLines(value string) int {
	count := 0
	for _, line := range strings.Split(value, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}
