package portal

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	portaldbclient "github.com/alibaba/aigateway-group/aigateway-console/backend/utility/clients/portaldb"
)

func TestCreateInviteCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	expectSchema(mock)
	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO portal_invite_code (invite_code, status, expires_at)
		VALUES (?, ?, ?)`)).
		WithArgs(sqlmock.AnyArg(), "active", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT invite_code, status, expires_at, used_by_consumer, used_at, created_at
		FROM portal_invite_code
		WHERE invite_code = ?`)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"invite_code", "status", "expires_at", "used_by_consumer", "used_at", "created_at",
		}).AddRow("ABCD1234", "active", time.Now(), nil, nil, time.Now()))

	svc := New(portaldbclient.NewFromDB(portaldbclient.Config{Enabled: true, Driver: "mysql"}, db))
	item, err := svc.CreateInviteCode(context.Background(), 7)
	require.NoError(t, err)
	require.Equal(t, "active", item.Status)
	require.NotEmpty(t, item.InviteCode)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateAccountStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	expectSchema(mock)
	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE portal_users
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE consumer_name = ? AND deleted = 0`)).
		WithArgs("disabled", "demo").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT department_id, name, parent_department_id
		FROM portal_departments
		WHERE deleted = 0`)).
		WillReturnRows(sqlmock.NewRows([]string{"department_id", "name", "parent_department_id"}))
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT consumer_name, display_name, email, status, user_level, source, department_id,
			parent_consumer_name, is_department_admin, last_login_at, temp_password
		FROM portal_users
		WHERE deleted = 0
		ORDER BY consumer_name ASC`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"consumer_name", "display_name", "email", "status", "user_level", "source",
			"department_id", "parent_consumer_name", "is_department_admin", "last_login_at", "temp_password",
		}).AddRow("demo", "Demo", "demo@example.com", "disabled", "normal", "console", nil, nil, false, nil, nil))

	svc := New(portaldbclient.NewFromDB(portaldbclient.Config{Enabled: true, Driver: "mysql"}, db))
	item, err := svc.UpdateAccountStatus(context.Background(), "demo", "disabled")
	require.NoError(t, err)
	require.Equal(t, "disabled", item.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func expectSchema(mock sqlmock.Sqlmock) {
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS portal_departments").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS portal_users").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS portal_invite_code").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS portal_asset_grant").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS portal_model_asset").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS portal_model_binding").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS portal_model_binding_price_version").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS portal_agent_catalog").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS portal_ai_sensitive_detect_rule").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS portal_ai_sensitive_replace_rule").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS portal_ai_sensitive_system_config").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS portal_ai_sensitive_block_audit").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS job_run_record").WillReturnResult(sqlmock.NewResult(0, 0))
}
