package portaldb

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	portalshared "higress-portal-backend/schema/shared"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestEnsureSchemaAndMigrateLegacyDataAgainstSharedPortalSchema(t *testing.T) {
	ctx := context.Background()
	db := openMySQLForTest(t, ctx, "console_portaldb_it")

	require.NoError(t, portalshared.ApplyToSQL(ctx, db))

	_, err := db.ExecContext(ctx, `
		INSERT INTO org_department (department_id, name, parent_department_id, admin_consumer_name, path, level, sort_order, status)
		VALUES ('root', 'Root', NULL, NULL, 'Root', 0, 0, 'active')`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS portal_users (
			consumer_name VARCHAR(128) PRIMARY KEY,
			display_name VARCHAR(128) NULL,
			email VARCHAR(255) NULL,
			password_hash VARCHAR(255) NULL,
			status VARCHAR(16) NULL,
			source VARCHAR(16) NULL,
			user_level VARCHAR(16) NULL,
			department_id VARCHAR(64) NULL,
			parent_consumer_name VARCHAR(128) NULL,
			deleted TINYINT(1) NOT NULL DEFAULT 0,
			last_login_at DATETIME NULL
		)`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS portal_departments (
			department_id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(128) NOT NULL,
			parent_department_id VARCHAR(64) NULL,
			admin_consumer_name VARCHAR(128) NULL,
			deleted TINYINT(1) NOT NULL DEFAULT 0
		)`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS portal_asset_grant (
			asset_type VARCHAR(32) NOT NULL,
			asset_id VARCHAR(128) NOT NULL,
			subject_type VARCHAR(32) NOT NULL,
			subject_id VARCHAR(128) NOT NULL
		)`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS portal_ai_quota_user_policy (
			route_name VARCHAR(255) NOT NULL,
			consumer_name VARCHAR(128) NOT NULL,
			limit_total BIGINT NOT NULL DEFAULT 0,
			limit_5h BIGINT NOT NULL DEFAULT 0,
			limit_daily BIGINT NOT NULL DEFAULT 0,
			daily_reset_mode VARCHAR(16) NOT NULL DEFAULT 'fixed',
			daily_reset_time VARCHAR(5) NOT NULL DEFAULT '00:00',
			limit_weekly BIGINT NOT NULL DEFAULT 0,
			limit_monthly BIGINT NOT NULL DEFAULT 0,
			cost_reset_at DATETIME NULL,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO portal_departments (department_id, name, parent_department_id, admin_consumer_name, deleted)
		VALUES ('dept-eng', 'Engineering', 'root', 'demo', 0)`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO portal_users (
			consumer_name, display_name, email, password_hash, status, source, user_level, department_id, parent_consumer_name, deleted
		) VALUES ('demo', 'Demo', 'demo@example.com', 'hash', 'active', 'console', 'pro', 'dept-eng', NULL, 0)`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO portal_asset_grant (asset_type, asset_id, subject_type, subject_id)
		VALUES ('model_binding', 'binding-1', 'consumer', 'demo')`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO portal_ai_quota_user_policy (
			route_name, consumer_name, limit_total, limit_5h, limit_daily, daily_reset_mode, daily_reset_time, limit_weekly, limit_monthly, cost_reset_at
		) VALUES ('route-a', 'demo', 1000, 200, 300, 'fixed', '08:00', 400, 500, NULL)`)
	require.NoError(t, err)

	client := NewFromDB(Config{Enabled: true, Driver: "mysql", AutoMigrate: true}, db)
	require.NoError(t, client.EnsureSchema(ctx))
	require.NoError(t, client.MigrateLegacyData(ctx))

	assertTableExists(t, ctx, db, "portal_model_binding_price_version")
	assertTableExists(t, ctx, db, "portal_ai_sensitive_detect_rule")
	assertTableExists(t, ctx, db, "portal_ai_quota_balance")
	assertTableExists(t, ctx, db, "job_run_record")

	var (
		displayName string
		email       string
		userLevel   string
	)
	err = db.QueryRowContext(ctx, `
		SELECT display_name, email, user_level
		FROM portal_user
		WHERE consumer_name = 'demo'`).Scan(&displayName, &email, &userLevel)
	require.NoError(t, err)
	require.Equal(t, "Demo", displayName)
	require.Equal(t, "demo@example.com", email)
	require.Equal(t, "pro", userLevel)

	var departmentID string
	err = db.QueryRowContext(ctx, `
		SELECT department_id
		FROM org_account_membership
		WHERE consumer_name = 'demo'`).Scan(&departmentID)
	require.NoError(t, err)
	require.Equal(t, "dept-eng", departmentID)

	var grantCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM asset_grant
		WHERE asset_type = 'model_binding' AND asset_id = 'binding-1' AND subject_type = 'consumer' AND subject_id = 'demo'`).Scan(&grantCount)
	require.NoError(t, err)
	require.Equal(t, 1, grantCount)

	var quotaTotal int64
	err = db.QueryRowContext(ctx, `
		SELECT limit_total_micro_yuan
		FROM quota_policy_user
		WHERE consumer_name = 'demo'`).Scan(&quotaTotal)
	require.NoError(t, err)
	require.EqualValues(t, 1000, quotaTotal)
}

func openMySQLForTest(t *testing.T, ctx context.Context, databaseName string) *sql.DB {
	t.Helper()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mysql:8.4",
			ExposedPorts: []string{"3306/tcp"},
			Env: map[string]string{
				"MYSQL_ROOT_PASSWORD": "root",
				"MYSQL_DATABASE":      databaseName,
			},
			WaitingFor: wait.ForListeningPort("3306/tcp").WithStartupTimeout(90 * time.Second),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "3306/tcp")
	require.NoError(t, err)

	dsn := fmt.Sprintf("root:root@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=UTC", host, port.Port(), databaseName)
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	require.Eventually(t, func() bool {
		return db.PingContext(ctx) == nil
	}, 30*time.Second, 500*time.Millisecond)
	return db
}

func assertTableExists(t *testing.T, ctx context.Context, db *sql.DB, table string) {
	t.Helper()

	var count int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?`, table).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "expected table %s to exist", table)
}
