package portaldb

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	portalshared "higress-portal-backend/schema/shared"
)

func TestSQLClientHealthyRetriesSchemaAfterStartupFailure(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectPing()
	for range portalshared.RequiredTables() {
		mock.ExpectQuery(regexp.QuoteMeta(Rebind("postgres", tableExistenceQuery("postgres")))).
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	}
	for range ensureSchemaDDLs("postgres") {
		mock.ExpectExec("CREATE TABLE IF NOT EXISTS").
			WillReturnResult(sqlmock.NewResult(0, 0))
	}

	client := &SQLClient{
		config: Config{
			Enabled:     true,
			Driver:      "postgres",
			AutoMigrate: false,
		},
		db:  db,
		err: errors.New("dial tcp aigateway-core-postgresql-pgpool:5432: connect: connection refused"),
	}

	require.NoError(t, client.Healthy(context.Background()))
	require.NoError(t, client.err)
	require.NoError(t, mock.ExpectationsWereMet())
}
