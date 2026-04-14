package portal

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type UsageStatsQuery struct {
	From *int64
	To   *int64
}

type UsageEventsQuery struct {
	From            *int64
	To              *int64
	ConsumerName    string
	DepartmentID    string
	IncludeChildren *bool
	APIKeyID        string
	ModelID         string
	RouteName       string
	RequestStatus   string
	UsageStatus     string
	PageNum         int
	PageSize        int
}

type DepartmentBillsQuery struct {
	From            *int64
	To              *int64
	DepartmentID    string
	IncludeChildren *bool
}

type PortalUsageStatRecord struct {
	ConsumerName               string `json:"consumerName"`
	ModelName                  string `json:"modelName"`
	RequestCount               int64  `json:"requestCount"`
	InputTokens                int64  `json:"inputTokens"`
	OutputTokens               int64  `json:"outputTokens"`
	TotalTokens                int64  `json:"totalTokens"`
	CacheCreationInputTokens   int64  `json:"cacheCreationInputTokens"`
	CacheCreation5mInputTokens int64  `json:"cacheCreation5mInputTokens"`
	CacheCreation1hInputTokens int64  `json:"cacheCreation1hInputTokens"`
	CacheReadInputTokens       int64  `json:"cacheReadInputTokens"`
	InputImageTokens           int64  `json:"inputImageTokens"`
	OutputImageTokens          int64  `json:"outputImageTokens"`
	InputImageCount            int64  `json:"inputImageCount"`
	OutputImageCount           int64  `json:"outputImageCount"`
}

type PortalUsageEventRecord struct {
	EventID        string `json:"eventId"`
	RequestID      string `json:"requestId"`
	TraceID        string `json:"traceId"`
	ConsumerName   string `json:"consumerName"`
	DepartmentID   string `json:"departmentId"`
	DepartmentPath string `json:"departmentPath"`
	APIKeyID       string `json:"apiKeyId"`
	ModelID        string `json:"modelId"`
	PriceVersionID *int64 `json:"priceVersionId,omitempty"`
	RouteName      string `json:"routeName"`
	RequestKind    string `json:"requestKind"`
	RequestStatus  string `json:"requestStatus"`
	UsageStatus    string `json:"usageStatus"`
	HTTPStatus     *int   `json:"httpStatus,omitempty"`
	InputTokens    int64  `json:"inputTokens"`
	OutputTokens   int64  `json:"outputTokens"`
	TotalTokens    int64  `json:"totalTokens"`
	RequestCount   int64  `json:"requestCount"`
	CostMicroYuan  int64  `json:"costMicroYuan"`
	OccurredAt     string `json:"occurredAt,omitempty"`
}

type PortalDepartmentBillRecord struct {
	DepartmentID    string  `json:"departmentId"`
	DepartmentName  string  `json:"departmentName"`
	DepartmentPath  string  `json:"departmentPath"`
	RequestCount    int64   `json:"requestCount"`
	TotalTokens     int64   `json:"totalTokens"`
	TotalCost       float64 `json:"totalCost"`
	ActiveConsumers int64   `json:"activeConsumers"`
}

func (s *Service) ListUsageStats(ctx context.Context, query UsageStatsQuery) ([]PortalUsageStatRecord, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}

	fromTime, toTime := normalizeStatsRange(query.From, query.To, time.Hour)
	rows, err := db.QueryContext(ctx, `
		SELECT consumer_name, model_id,
			COALESCE(SUM(request_count), 0) AS request_count,
			COALESCE(SUM(input_tokens), 0) AS input_tokens,
			COALESCE(SUM(output_tokens), 0) AS output_tokens,
			COALESCE(SUM(total_tokens), 0) AS total_tokens,
			COALESCE(SUM(cache_creation_input_tokens), 0) AS cache_creation_input_tokens,
			COALESCE(SUM(cache_creation_5m_input_tokens), 0) AS cache_creation_5m_input_tokens,
			COALESCE(SUM(cache_creation_1h_input_tokens), 0) AS cache_creation_1h_input_tokens,
			COALESCE(SUM(cache_read_input_tokens), 0) AS cache_read_input_tokens,
			COALESCE(SUM(input_image_tokens), 0) AS input_image_tokens,
			COALESCE(SUM(output_image_tokens), 0) AS output_image_tokens,
			COALESCE(SUM(input_image_count), 0) AS input_image_count,
			COALESCE(SUM(output_image_count), 0) AS output_image_count
		FROM billing_usage_event
		WHERE request_status = 'success'
			AND usage_status = 'parsed'
			AND occurred_at >= ?
			AND occurred_at < ?
		GROUP BY consumer_name, model_id
		ORDER BY consumer_name ASC, model_id ASC`,
		fromTime,
		toTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query portal billing usage stats: %w", err)
	}
	defer rows.Close()

	items := make([]PortalUsageStatRecord, 0)
	for rows.Next() {
		var item PortalUsageStatRecord
		if err := rows.Scan(
			&item.ConsumerName,
			&item.ModelName,
			&item.RequestCount,
			&item.InputTokens,
			&item.OutputTokens,
			&item.TotalTokens,
			&item.CacheCreationInputTokens,
			&item.CacheCreation5mInputTokens,
			&item.CacheCreation1hInputTokens,
			&item.CacheReadInputTokens,
			&item.InputImageTokens,
			&item.OutputImageTokens,
			&item.InputImageCount,
			&item.OutputImageCount,
		); err != nil {
			return nil, err
		}
		if item.TotalTokens <= 0 {
			cacheCreationEffectiveTokens := maxInt64(
				item.CacheCreationInputTokens,
				item.CacheCreation5mInputTokens+item.CacheCreation1hInputTokens,
			)
			item.TotalTokens = item.InputTokens + item.OutputTokens + cacheCreationEffectiveTokens +
				item.CacheReadInputTokens + item.InputImageTokens + item.OutputImageTokens
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) ListUsageEvents(ctx context.Context, query UsageEventsQuery) ([]PortalUsageEventRecord, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}

	fromTime, toTime := normalizeStatsRange(query.From, query.To, 24*time.Hour)
	pageNum := query.PageNum
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}

	statement := strings.Builder{}
	statement.WriteString(`SELECT event_id, request_id, trace_id, consumer_name, department_id,
		department_path, api_key_id, model_id, price_version_id, route_name, request_kind,
		request_status, usage_status, http_status, input_tokens, output_tokens, total_tokens,
		request_count, cost_micro_yuan, occurred_at
		FROM billing_usage_event
		WHERE occurred_at >= ? AND occurred_at < ?`)
	args := []any{fromTime, toTime}
	appendPortalStatsEqualsFilter(&statement, &args, "consumer_name", query.ConsumerName)
	appendPortalStatsEqualsFilter(&statement, &args, "api_key_id", query.APIKeyID)
	appendPortalStatsEqualsFilter(&statement, &args, "model_id", query.ModelID)
	appendPortalStatsEqualsFilter(&statement, &args, "route_name", query.RouteName)
	appendPortalStatsEqualsFilter(&statement, &args, "request_status", query.RequestStatus)
	appendPortalStatsEqualsFilter(&statement, &args, "usage_status", query.UsageStatus)
	if err := s.appendPortalStatsDepartmentScope(ctx, db, &statement, &args, query.DepartmentID, query.IncludeChildren); err != nil {
		return nil, err
	}
	statement.WriteString(` ORDER BY occurred_at DESC, id DESC LIMIT ? OFFSET ?`)
	args = append(args, pageSize, (pageNum-1)*pageSize)

	rows, err := db.QueryContext(ctx, statement.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query portal billing usage events: %w", err)
	}
	defer rows.Close()

	items := make([]PortalUsageEventRecord, 0)
	for rows.Next() {
		var item PortalUsageEventRecord
		var priceVersionID sql.NullInt64
		var httpStatus sql.NullInt64
		var occurredAt sql.NullTime
		if err := rows.Scan(
			&item.EventID,
			&item.RequestID,
			&item.TraceID,
			&item.ConsumerName,
			&item.DepartmentID,
			&item.DepartmentPath,
			&item.APIKeyID,
			&item.ModelID,
			&priceVersionID,
			&item.RouteName,
			&item.RequestKind,
			&item.RequestStatus,
			&item.UsageStatus,
			&httpStatus,
			&item.InputTokens,
			&item.OutputTokens,
			&item.TotalTokens,
			&item.RequestCount,
			&item.CostMicroYuan,
			&occurredAt,
		); err != nil {
			return nil, err
		}
		if priceVersionID.Valid {
			value := priceVersionID.Int64
			item.PriceVersionID = &value
		}
		if httpStatus.Valid {
			value := int(httpStatus.Int64)
			item.HTTPStatus = &value
		}
		if occurredAt.Valid {
			item.OccurredAt = occurredAt.Time.Format(time.RFC3339)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) ListDepartmentBills(ctx context.Context, query DepartmentBillsQuery) ([]PortalDepartmentBillRecord, error) {
	db, err := s.db(ctx)
	if err != nil {
		return nil, err
	}

	fromTime, toTime := normalizeStatsRange(query.From, query.To, 30*24*time.Hour)
	statement := strings.Builder{}
	statement.WriteString(`SELECT d.department_id, COALESCE(o.name, '') AS department_name,
		COALESCE(d.department_path, '') AS department_path, COALESCE(SUM(d.request_count), 0) AS request_count,
		COALESCE(SUM(d.total_tokens), 0) AS total_tokens, COALESCE(SUM(d.cost_amount), 0) AS total_cost,
		COUNT(DISTINCT d.consumer_name) AS active_consumers
		FROM portal_usage_daily d
		LEFT JOIN org_department o ON o.department_id = d.department_id
		WHERE d.billing_date >= ? AND d.billing_date <= ?`)
	args := []any{fromTime.Format("2006-01-02"), toTime.Format("2006-01-02")}
	if err := s.appendPortalStatsDepartmentScope(ctx, db, &statement, &args, query.DepartmentID, query.IncludeChildren); err != nil {
		return nil, err
	}
	statement.WriteString(` GROUP BY d.department_id, o.name, d.department_path ORDER BY d.department_path ASC, d.department_id ASC`)

	rows, err := db.QueryContext(ctx, statement.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query portal department bills: %w", err)
	}
	defer rows.Close()

	items := make([]PortalDepartmentBillRecord, 0)
	for rows.Next() {
		var item PortalDepartmentBillRecord
		if err := rows.Scan(
			&item.DepartmentID,
			&item.DepartmentName,
			&item.DepartmentPath,
			&item.RequestCount,
			&item.TotalTokens,
			&item.TotalCost,
			&item.ActiveConsumers,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func normalizeStatsRange(fromMillis, toMillis *int64, fallback time.Duration) (time.Time, time.Time) {
	now := time.Now()
	toTime := now
	if toMillis != nil && *toMillis > 0 {
		toTime = time.UnixMilli(*toMillis)
	}
	fromTime := toTime.Add(-fallback)
	if fromMillis != nil && *fromMillis > 0 && *fromMillis < toTime.UnixMilli() {
		fromTime = time.UnixMilli(*fromMillis)
	}
	return fromTime, toTime
}

func appendPortalStatsEqualsFilter(statement *strings.Builder, args *[]any, columnName, value string) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return
	}
	statement.WriteString(" AND ")
	statement.WriteString(columnName)
	statement.WriteString(" = ?")
	*args = append(*args, normalized)
}

func (s *Service) appendPortalStatsDepartmentScope(
	ctx context.Context,
	db *sql.DB,
	statement *strings.Builder,
	args *[]any,
	departmentID string,
	includeChildren *bool,
) error {
	normalizedDepartmentID := strings.TrimSpace(departmentID)
	if normalizedDepartmentID == "" {
		return nil
	}
	shouldIncludeChildren := true
	if includeChildren != nil {
		shouldIncludeChildren = *includeChildren
	}
	departmentIDs, err := s.resolvePortalStatsDepartmentIDs(ctx, db, normalizedDepartmentID, shouldIncludeChildren)
	if err != nil {
		return err
	}
	if len(departmentIDs) == 0 {
		statement.WriteString(" AND department_id = ''")
		return nil
	}
	statement.WriteString(" AND department_id IN (")
	for index, item := range departmentIDs {
		if index > 0 {
			statement.WriteString(", ")
		}
		statement.WriteString("?")
		*args = append(*args, item)
	}
	statement.WriteString(")")
	return nil
}

func (s *Service) resolvePortalStatsDepartmentIDs(
	ctx context.Context,
	db *sql.DB,
	departmentID string,
	includeChildren bool,
) ([]string, error) {
	if !includeChildren {
		return []string{departmentID}, nil
	}

	var rootPath string
	if err := db.QueryRowContext(ctx,
		`SELECT path FROM org_department WHERE department_id = ? LIMIT 1`,
		departmentID,
	).Scan(&rootPath); err != nil {
		if err == sql.ErrNoRows {
			return []string{departmentID}, nil
		}
		return nil, err
	}
	if strings.TrimSpace(rootPath) == "" {
		return []string{departmentID}, nil
	}

	rows, err := db.QueryContext(ctx,
		`SELECT department_id FROM org_department
		WHERE department_id = ? OR path = ? OR path LIKE ?
		ORDER BY path ASC`,
		departmentID,
		rootPath,
		rootPath+"/%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]string, 0)
	for rows.Next() {
		var item string
		if err := rows.Scan(&item); err != nil {
			return nil, err
		}
		if strings.TrimSpace(item) != "" {
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		return []string{departmentID}, nil
	}
	return items, rows.Err()
}

func maxInt64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}
