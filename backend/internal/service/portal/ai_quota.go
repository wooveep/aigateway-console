package portal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

const builtinQuotaAdminConsumer = "administrator"

type AIQuotaMenuState struct {
	Enabled           bool `json:"enabled"`
	EnabledRouteCount int  `json:"enabledRouteCount"`
}

type AIQuotaRouteSummary struct {
	RouteName         string   `json:"routeName"`
	Domains           []string `json:"domains,omitempty"`
	Path              string   `json:"path,omitempty"`
	RedisKeyPrefix    string   `json:"redisKeyPrefix"`
	AdminConsumer     string   `json:"adminConsumer"`
	AdminPath         string   `json:"adminPath"`
	QuotaUnit         string   `json:"quotaUnit,omitempty"`
	ScheduleRuleCount int      `json:"scheduleRuleCount"`
}

type AIQuotaConsumerQuota struct {
	ConsumerName string `json:"consumerName"`
	Quota        int64  `json:"quota"`
}

type AIQuotaUserPolicy struct {
	ConsumerName   string     `json:"consumerName"`
	LimitTotal     int64      `json:"limitTotal"`
	Limit5h        int64      `json:"limit5h"`
	LimitDaily     int64      `json:"limitDaily"`
	DailyResetMode string     `json:"dailyResetMode"`
	DailyResetTime string     `json:"dailyResetTime"`
	LimitWeekly    int64      `json:"limitWeekly"`
	LimitMonthly   int64      `json:"limitMonthly"`
	CostResetAt    *time.Time `json:"costResetAt,omitempty"`
}

type AIQuotaUserPolicyRequest struct {
	LimitTotal     int64      `json:"limitTotal"`
	Limit5h        int64      `json:"limit5h"`
	LimitDaily     int64      `json:"limitDaily"`
	DailyResetMode string     `json:"dailyResetMode"`
	DailyResetTime string     `json:"dailyResetTime"`
	LimitWeekly    int64      `json:"limitWeekly"`
	LimitMonthly   int64      `json:"limitMonthly"`
	CostResetAt    *time.Time `json:"costResetAt"`
}

type AIQuotaScheduleRule struct {
	ID            string     `json:"id"`
	ConsumerName  string     `json:"consumerName"`
	Action        string     `json:"action"`
	Cron          string     `json:"cron"`
	Value         int64      `json:"value"`
	Enabled       bool       `json:"enabled"`
	CreatedAt     *time.Time `json:"createdAt,omitempty"`
	UpdatedAt     *time.Time `json:"updatedAt,omitempty"`
	LastAppliedAt *time.Time `json:"lastAppliedAt,omitempty"`
	LastError     string     `json:"lastError,omitempty"`
}

type AIQuotaScheduleRuleRequest struct {
	ID           string `json:"id"`
	ConsumerName string `json:"consumerName"`
	Action       string `json:"action"`
	Cron         string `json:"cron"`
	Value        int64  `json:"value"`
	Enabled      *bool  `json:"enabled"`
}

func (s *Service) GetAIQuotaMenuState(ctx context.Context) (*AIQuotaMenuState, error) {
	routes, err := s.ListAIQuotaRoutes(ctx)
	if err != nil {
		return nil, err
	}
	return &AIQuotaMenuState{
		Enabled:           len(routes) > 0,
		EnabledRouteCount: len(routes),
	}, nil
}

func (s *Service) ListAIQuotaRoutes(ctx context.Context) ([]AIQuotaRouteSummary, error) {
	if s.k8sClient == nil {
		return []AIQuotaRouteSummary{}, nil
	}
	rawRoutes, err := s.k8sClient.ListResources(ctx, "ai-routes")
	if err != nil {
		return nil, err
	}

	scheduleCounts := map[string]int{}
	if db, err := s.db(ctx); err == nil {
		rows, queryErr := db.QueryContext(ctx, `
			SELECT route_name, COUNT(1)
			FROM portal_ai_quota_schedule_rule
			GROUP BY route_name`)
		if queryErr == nil {
			defer rows.Close()
			for rows.Next() {
				var (
					routeName string
					count     int
				)
				if scanErr := rows.Scan(&routeName, &count); scanErr != nil {
					return nil, scanErr
				}
				scheduleCounts[routeName] = count
			}
			if rows.Err() != nil {
				return nil, rows.Err()
			}
		}
	}

	items := make([]AIQuotaRouteSummary, 0, len(rawRoutes))
	for _, item := range rawRoutes {
		name := strings.TrimSpace(fmt.Sprint(item["name"]))
		if name == "" {
			continue
		}
		summary := AIQuotaRouteSummary{
			RouteName:         name,
			Domains:           readStringSlice(item["domains"]),
			Path:              firstNonEmpty(extractMatchValue(item["pathPredicate"]), extractMatchValue(item["path"]), "/"+name),
			RedisKeyPrefix:    firstNonEmpty(strings.TrimSpace(fmt.Sprint(item["redisKeyPrefix"])), "aigateway:quota:"+name),
			AdminConsumer:     firstNonEmpty(strings.TrimSpace(fmt.Sprint(item["adminConsumer"])), builtinQuotaAdminConsumer),
			AdminPath:         firstNonEmpty(strings.TrimSpace(fmt.Sprint(item["adminPath"])), "/v1/ai/quotas/routes/"+name+"/consumers"),
			QuotaUnit:         firstNonEmpty(strings.TrimSpace(fmt.Sprint(item["quotaUnit"])), "amount"),
			ScheduleRuleCount: scheduleCounts[name],
		}
		items = append(items, summary)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].RouteName < items[j].RouteName
	})
	return items, nil
}

func (s *Service) ListAIQuotaConsumers(ctx context.Context, routeName string) ([]AIQuotaConsumerQuota, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}
	routeName = strings.TrimSpace(routeName)
	if routeName == "" {
		return nil, errors.New("routeName cannot be blank")
	}

	balanceRows, err := db.QueryContext(ctx, `
		SELECT consumer_name, quota
		FROM portal_ai_quota_balance
		WHERE route_name = ?`, routeName)
	if err != nil {
		return nil, err
	}
	defer balanceRows.Close()

	balances := map[string]int64{}
	for balanceRows.Next() {
		var (
			consumerName string
			quota        int64
		)
		if err := balanceRows.Scan(&consumerName, &quota); err != nil {
			return nil, err
		}
		balances[consumerName] = quota
	}
	if err := balanceRows.Err(); err != nil {
		return nil, err
	}

	names := map[string]struct{}{
		builtinQuotaAdminConsumer: {},
	}
	rows, err := db.QueryContext(ctx, `
		SELECT consumer_name
		FROM portal_user
		WHERE COALESCE(is_deleted, 0) = 0
		ORDER BY consumer_name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var consumerName string
		if err := rows.Scan(&consumerName); err != nil {
			return nil, err
		}
		names[consumerName] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for consumerName := range balances {
		names[consumerName] = struct{}{}
	}

	consumers := make([]AIQuotaConsumerQuota, 0, len(names))
	for consumerName := range names {
		consumers = append(consumers, AIQuotaConsumerQuota{
			ConsumerName: consumerName,
			Quota:        balances[consumerName],
		})
	}
	sort.Slice(consumers, func(i, j int) bool {
		return consumers[i].ConsumerName < consumers[j].ConsumerName
	})
	return consumers, nil
}

func (s *Service) RefreshAIQuota(ctx context.Context, routeName, consumerName string, value int64) (*AIQuotaConsumerQuota, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}
	routeName = strings.TrimSpace(routeName)
	consumerName = strings.TrimSpace(consumerName)
	if routeName == "" || consumerName == "" {
		return nil, errors.New("routeName and consumerName are required")
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO portal_ai_quota_balance (route_name, consumer_name, quota)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE quota = VALUES(quota), updated_at = CURRENT_TIMESTAMP`,
		routeName, consumerName, value,
	); err != nil {
		return nil, err
	}
	return &AIQuotaConsumerQuota{ConsumerName: consumerName, Quota: value}, nil
}

func (s *Service) DeltaAIQuota(ctx context.Context, routeName, consumerName string, delta int64) (*AIQuotaConsumerQuota, error) {
	current, err := s.getAIQuotaBalance(ctx, routeName, consumerName)
	if err != nil {
		return nil, err
	}
	return s.RefreshAIQuota(ctx, routeName, consumerName, current+delta)
}

func (s *Service) GetAIQuotaUserPolicy(ctx context.Context, routeName, consumerName string) (*AIQuotaUserPolicy, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}
	routeName = strings.TrimSpace(routeName)
	consumerName = strings.TrimSpace(consumerName)
	if routeName == "" || consumerName == "" {
		return nil, errors.New("routeName and consumerName are required")
	}

	policy := defaultAIQuotaPolicy(consumerName)
	var costResetAt sql.NullTime
	err = db.QueryRowContext(ctx, `
		SELECT limit_total_micro_yuan, limit_5h_micro_yuan, limit_daily_micro_yuan, daily_reset_mode, daily_reset_time,
			limit_weekly_micro_yuan, limit_monthly_micro_yuan, cost_reset_at
		FROM quota_policy_user
		WHERE consumer_name = ?`,
		consumerName,
	).Scan(
		&policy.LimitTotal,
		&policy.Limit5h,
		&policy.LimitDaily,
		&policy.DailyResetMode,
		&policy.DailyResetTime,
		&policy.LimitWeekly,
		&policy.LimitMonthly,
		&costResetAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return policy, nil
		}
		return nil, err
	}
	if costResetAt.Valid {
		policy.CostResetAt = &costResetAt.Time
	}
	return policy, nil
}

func (s *Service) SaveAIQuotaUserPolicy(
	ctx context.Context,
	routeName, consumerName string,
	request AIQuotaUserPolicyRequest,
) (*AIQuotaUserPolicy, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}
	routeName = strings.TrimSpace(routeName)
	consumerName = strings.TrimSpace(consumerName)
	if routeName == "" || consumerName == "" {
		return nil, errors.New("routeName and consumerName are required")
	}

	mode := firstNonEmpty(strings.TrimSpace(request.DailyResetMode), "fixed")
	resetTime := firstNonEmpty(strings.TrimSpace(request.DailyResetTime), "00:00")
	if _, err := db.ExecContext(ctx, `
		INSERT INTO quota_policy_user (
			consumer_name, limit_total_micro_yuan, limit_5h_micro_yuan, limit_daily_micro_yuan,
			daily_reset_mode, daily_reset_time, limit_weekly_micro_yuan, limit_monthly_micro_yuan, cost_reset_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			limit_total_micro_yuan = VALUES(limit_total_micro_yuan),
			limit_5h_micro_yuan = VALUES(limit_5h_micro_yuan),
			limit_daily_micro_yuan = VALUES(limit_daily_micro_yuan),
			daily_reset_mode = VALUES(daily_reset_mode),
			daily_reset_time = VALUES(daily_reset_time),
			limit_weekly_micro_yuan = VALUES(limit_weekly_micro_yuan),
			limit_monthly_micro_yuan = VALUES(limit_monthly_micro_yuan),
			cost_reset_at = VALUES(cost_reset_at),
			updated_at = CURRENT_TIMESTAMP`,
		consumerName,
		request.LimitTotal,
		request.Limit5h,
		request.LimitDaily,
		mode,
		resetTime,
		request.LimitWeekly,
		request.LimitMonthly,
		request.CostResetAt,
	); err != nil {
		return nil, err
	}
	return s.GetAIQuotaUserPolicy(ctx, routeName, consumerName)
}

func (s *Service) ListAIQuotaScheduleRules(ctx context.Context, routeName, consumerName string) ([]AIQuotaScheduleRule, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}
	routeName = strings.TrimSpace(routeName)
	if routeName == "" {
		return nil, errors.New("routeName cannot be blank")
	}

	statement := `
		SELECT id, consumer_name, action, cron, value, enabled, created_at, updated_at, last_applied_at, last_error
		FROM portal_ai_quota_schedule_rule
		WHERE route_name = ?`
	args := []any{routeName}
	if strings.TrimSpace(consumerName) != "" {
		statement += ` AND consumer_name = ?`
		args = append(args, strings.TrimSpace(consumerName))
	}
	statement += ` ORDER BY consumer_name ASC, created_at DESC`

	rows, err := db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AIQuotaScheduleRule, 0)
	for rows.Next() {
		var (
			item          AIQuotaScheduleRule
			enabled       int
			createdAt     sql.NullTime
			updatedAt     sql.NullTime
			lastAppliedAt sql.NullTime
			lastError     sql.NullString
		)
		if err := rows.Scan(
			&item.ID,
			&item.ConsumerName,
			&item.Action,
			&item.Cron,
			&item.Value,
			&enabled,
			&createdAt,
			&updatedAt,
			&lastAppliedAt,
			&lastError,
		); err != nil {
			return nil, err
		}
		item.Enabled = enabled > 0
		if createdAt.Valid {
			item.CreatedAt = &createdAt.Time
		}
		if updatedAt.Valid {
			item.UpdatedAt = &updatedAt.Time
		}
		if lastAppliedAt.Valid {
			item.LastAppliedAt = &lastAppliedAt.Time
		}
		item.LastError = lastError.String
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) SaveAIQuotaScheduleRule(
	ctx context.Context,
	routeName string,
	request AIQuotaScheduleRuleRequest,
) (*AIQuotaScheduleRule, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}
	routeName = strings.TrimSpace(routeName)
	consumerName := strings.TrimSpace(request.ConsumerName)
	if routeName == "" || consumerName == "" {
		return nil, errors.New("routeName and consumerName are required")
	}
	action := strings.ToUpper(strings.TrimSpace(request.Action))
	if action != "REFRESH" && action != "DELTA" {
		return nil, errors.New("action must be REFRESH or DELTA")
	}
	cron := strings.TrimSpace(request.Cron)
	if cron == "" {
		return nil, errors.New("cron cannot be blank")
	}
	ruleID := strings.TrimSpace(request.ID)
	if ruleID == "" {
		ruleID = "quota-rule-" + uuid.NewString()[:8]
	}
	enabled := true
	if request.Enabled != nil {
		enabled = *request.Enabled
	}
	enabledValue := 0
	if enabled {
		enabledValue = 1
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO portal_ai_quota_schedule_rule (
			id, route_name, consumer_name, action, cron, value, enabled
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			consumer_name = VALUES(consumer_name),
			action = VALUES(action),
			cron = VALUES(cron),
			value = VALUES(value),
			enabled = VALUES(enabled),
			updated_at = CURRENT_TIMESTAMP`,
		ruleID, routeName, consumerName, action, cron, request.Value, enabledValue,
	); err != nil {
		return nil, err
	}
	return s.getAIQuotaScheduleRule(ctx, routeName, ruleID)
}

func (s *Service) DeleteAIQuotaScheduleRule(ctx context.Context, routeName, ruleID string) error {
	db, err := s.db(ctx)
	if err != nil {
		return err
	}
	routeName = strings.TrimSpace(routeName)
	ruleID = strings.TrimSpace(ruleID)
	if routeName == "" || ruleID == "" {
		return errors.New("routeName and ruleId are required")
	}
	result, err := db.ExecContext(ctx, `
		DELETE FROM portal_ai_quota_schedule_rule
		WHERE route_name = ? AND id = ?`,
		routeName, ruleID,
	)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("schedule rule not found: %s", ruleID)
	}
	return nil
}

func (s *Service) getAIQuotaBalance(ctx context.Context, routeName, consumerName string) (int64, error) {
	db, err := s.db(ctx)
	if err != nil {
		return 0, err
	}
	var quota int64
	err = db.QueryRowContext(ctx, `
		SELECT quota
		FROM portal_ai_quota_balance
		WHERE route_name = ? AND consumer_name = ?`,
		strings.TrimSpace(routeName),
		strings.TrimSpace(consumerName),
	).Scan(&quota)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return quota, nil
}

func (s *Service) getAIQuotaScheduleRule(ctx context.Context, routeName, ruleID string) (*AIQuotaScheduleRule, error) {
	items, err := s.ListAIQuotaScheduleRules(ctx, routeName, "")
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.ID == ruleID {
			return &item, nil
		}
	}
	return nil, fmt.Errorf("schedule rule not found: %s", ruleID)
}

func defaultAIQuotaPolicy(consumerName string) *AIQuotaUserPolicy {
	return &AIQuotaUserPolicy{
		ConsumerName:   consumerName,
		DailyResetMode: "fixed",
		DailyResetTime: "00:00",
	}
}

func extractMatchValue(value any) string {
	record, ok := value.(map[string]any)
	if !ok {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(record["matchValue"]))
}

func readStringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		result := append([]string{}, typed...)
		sort.Strings(result)
		return result
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			trimmed := strings.TrimSpace(fmt.Sprint(item))
			if trimmed == "" {
				continue
			}
			items = append(items, trimmed)
		}
		sort.Strings(items)
		return items
	default:
		return []string{}
	}
}
