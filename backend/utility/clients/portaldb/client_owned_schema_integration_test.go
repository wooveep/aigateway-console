//go:build integration
// +build integration

package portaldb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	portalshared "higress-portal-backend/schema/shared"
)

func TestEnsureSchemaCreatesConsoleOwnedTablesWhenAutoMigrateDisabled(t *testing.T) {
	ctx := context.Background()
	db := openPostgresForTest(t, ctx, "console_portaldb_owned_schema_it")

	require.NoError(t, portalshared.ApplyToSQLWithDriver(ctx, db, "postgres"))

	client := NewFromDB(Config{Enabled: true, Driver: "postgres", AutoMigrate: false}, db)
	require.NoError(t, client.EnsureSchema(ctx))

	assertTableExists(t, ctx, db, "portal_ai_quota_balance")
	assertTableExists(t, ctx, db, "portal_ai_quota_schedule_rule")
	assertTableExists(t, ctx, db, "job_run_record")
}
