package portal

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	portaldbclient "github.com/wooveep/aigateway-console/backend/utility/clients/portaldb"
)

func TestListUsageStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	expectSchema(mock)
	rows := sqlmock.NewRows([]string{
		"consumer_name", "model_id", "request_count", "input_tokens", "output_tokens", "total_tokens",
		"cache_creation_input_tokens", "cache_creation_5m_input_tokens", "cache_creation_1h_input_tokens",
		"cache_read_input_tokens", "input_image_tokens", "output_image_tokens", "input_image_count", "output_image_count",
	}).AddRow(
		"alice", "qwen", 3, 120, 80, 0, 20, 10, 5, 7, 0, 0, 0, 0,
	)
	mock.ExpectQuery("FROM billing_usage_event").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	svc := New(portaldbclient.NewFromDB(portaldbclient.Config{Enabled: true, Driver: "mysql", AutoMigrate: true}, db))
	items, err := svc.ListUsageStats(context.Background(), UsageStatsQuery{})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, int64(227), items[0].TotalTokens)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListUsageEventsUnavailable(t *testing.T) {
	svc := New(portaldbclient.New(portaldbclient.Config{}))
	_, err := svc.ListUsageEvents(context.Background(), UsageEventsQuery{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unavailable")
}

func TestListDepartmentBillsWithDepartmentScope(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	expectSchema(mock)
	mock.ExpectQuery("SELECT path FROM org_department WHERE department_id = \\? LIMIT 1").
		WithArgs("dept-a").
		WillReturnRows(sqlmock.NewRows([]string{"path"}).AddRow("root/dept-a"))
	mock.ExpectQuery("SELECT department_id FROM org_department").
		WithArgs("dept-a", "root/dept-a", "root/dept-a/%").
		WillReturnRows(sqlmock.NewRows([]string{"department_id"}).AddRow("dept-a").AddRow("dept-b"))
	mock.ExpectQuery("FROM portal_usage_daily").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "dept-a", "dept-b").
		WillReturnRows(sqlmock.NewRows([]string{
			"department_id", "department_name", "department_path", "request_count", "total_tokens", "total_cost", "active_consumers",
		}).AddRow("dept-a", "Dept A", "root/dept-a", 10, 1000, 12.5, 3))

	svc := New(portaldbclient.NewFromDB(portaldbclient.Config{Enabled: true, Driver: "mysql", AutoMigrate: true}, db))
	includeChildren := true
	items, err := svc.ListDepartmentBills(context.Background(), DepartmentBillsQuery{
		DepartmentID:    "dept-a",
		IncludeChildren: &includeChildren,
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, int64(3), items[0].ActiveConsumers)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNormalizeStatsRangeDefaults(t *testing.T) {
	from, to := normalizeStatsRange(nil, nil, time.Hour)
	require.True(t, to.After(from))
	require.WithinDuration(t, to.Add(-time.Hour), from, 2*time.Second)
}
