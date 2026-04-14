package platform

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/wooveep/aigateway-console/backend/internal/consts"
	"github.com/wooveep/aigateway-console/backend/internal/model/response"
	grafanaclient "github.com/wooveep/aigateway-console/backend/utility/clients/grafana"
	k8sclient "github.com/wooveep/aigateway-console/backend/utility/clients/k8s"
	portaldbclient "github.com/wooveep/aigateway-console/backend/utility/clients/portaldb"
)

func TestServiceHealthz(t *testing.T) {
	svc := New(k8sclient.NewMemoryClient(), grafanaclient.New(grafanaclient.Config{}), portaldbclient.New(portaldbclient.Config{}))

	status, err := svc.Healthz(context.Background())

	require.NoError(t, err)
	require.Equal(t, "ok", status.Status)
	require.Equal(t, consts.ServiceName, status.Service)
	require.Equal(t, consts.LegacyBackendDir, status.LegacyBackend)
}

func TestServiceSystemInfo(t *testing.T) {
	svc := New(k8sclient.NewMemoryClient(), grafanaclient.New(grafanaclient.Config{}), portaldbclient.New(portaldbclient.Config{}))

	info, err := svc.SystemInfo(context.Background())

	require.NoError(t, err)
	require.Equal(t, consts.ServiceName, info.Service)
	require.Equal(t, consts.PreferredProduct, info.PreferredNaming)
	require.Contains(t, info.BusinessLines, "gateway")
	require.Contains(t, info.BusinessLines, "portal")
}

func TestServiceInitializeAndLogin(t *testing.T) {
	svc := New(k8sclient.NewMemoryClient(), grafanaclient.New(grafanaclient.Config{}), portaldbclient.New(portaldbclient.Config{}))

	err := svc.InitializeSystem(context.Background(), &response.User{
		Username:    "admin",
		Password:    "secret",
		DisplayName: "Admin",
	}, map[string]any{
		"login.prompt": "Hello",
	})
	require.NoError(t, err)
	require.True(t, svc.IsSystemInitialized(context.Background()))

	user, token, err := svc.Login(context.Background(), "admin", "secret")
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.Equal(t, "admin", user.Username)

	validated, err := svc.ValidateSessionToken(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, "admin", validated.Username)
}

func TestServiceLoadsPersistedAdminState(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	hash, err := hashAdminPassword("secret")
	require.NoError(t, err)

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS console_system_state").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT initialized, admin_username, admin_display_name, admin_password_hash, configs_json").
		WithArgs(consts.DefaultAdminStateKey).
		WillReturnRows(sqlmock.NewRows([]string{
			"initialized", "admin_username", "admin_display_name", "admin_password_hash", "configs_json",
		}).AddRow(true, "admin", "Admin", hash, `{"system.initialized":true,"login.prompt":"hello"}`))

	svc := New(
		k8sclient.NewMemoryClient(),
		grafanaclient.New(grafanaclient.Config{}),
		portaldbclient.NewFromDB(portaldbclient.Config{Enabled: true, Driver: "mysql"}, db),
	)

	require.True(t, svc.IsSystemInitialized(context.Background()))
	configs := svc.GetConfigs(context.Background())
	require.Equal(t, "hello", configs["login.prompt"])

	user, token, err := svc.Login(context.Background(), "admin", "secret")
	require.NoError(t, err)
	require.Equal(t, "admin", user.Username)

	validated, err := svc.ValidateSessionToken(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, "admin", validated.Username)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceMigratesLegacySecretStateToDatabase(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	client := k8sclient.NewMemoryClient()
	require.NoError(t, client.UpsertSecret(context.Background(), consts.DefaultSecretName, map[string]string{
		"adminUsername":    "admin",
		"adminPassword":    "secret",
		"adminDisplayName": "Admin",
	}))

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS console_system_state").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT initialized, admin_username, admin_display_name, admin_password_hash, configs_json").
		WithArgs(consts.DefaultAdminStateKey).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("INSERT INTO console_system_state").
		WithArgs(
			consts.DefaultAdminStateKey,
			true,
			"admin",
			"Admin",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	svc := New(
		client,
		grafanaclient.New(grafanaclient.Config{}),
		portaldbclient.NewFromDB(portaldbclient.Config{Enabled: true, Driver: "mysql"}, db),
	)

	require.True(t, svc.IsSystemInitialized(context.Background()))
	user, token, err := svc.Login(context.Background(), "admin", "secret")
	require.NoError(t, err)
	require.Equal(t, "admin", user.Username)

	validated, err := svc.ValidateSessionToken(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, "admin", validated.Username)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceSetAndGetAIGatewayConfig(t *testing.T) {
	svc := New(k8sclient.NewMemoryClient(), grafanaclient.New(grafanaclient.Config{}), portaldbclient.New(portaldbclient.Config{}))

	current, err := svc.GetAIGatewayConfig(context.Background())
	require.NoError(t, err)
	require.Contains(t, current, "ConfigMap")
	require.Contains(t, current, "aigateway:")

	updated, err := svc.SetAIGatewayConfig(context.Background(), `
apiVersion: v1
kind: ConfigMap
metadata:
  name: aigateway-console
  resourceVersion: "2"
data:
  aigateway: |
    enabled: true
  mesh: |
    defaultConfig: {}
  meshNetworks: |
    networks: {}
`)
	require.NoError(t, err)
	require.Contains(t, updated, "resourceVersion: \"2\"")
}

func TestServiceMigratesLegacyConfigKey(t *testing.T) {
	client := k8sclient.NewMemoryClient()
	svc := New(client, grafanaclient.New(grafanaclient.Config{}), portaldbclient.New(portaldbclient.Config{}))

	current, err := svc.GetAIGatewayConfig(context.Background())

	require.NoError(t, err)
	require.Contains(t, current, "aigateway:")
	require.NotContains(t, current, "higress:")

	stored, err := client.ReadConfigMap(context.Background(), consts.DefaultConfigMapName)
	require.NoError(t, err)
	require.Equal(t, "enabled: true\n", stored["aigateway"])
	require.Empty(t, stored["higress"])
}

func TestBootstrapDefaultResourcesLockedDoesNotOverwriteExistingResources(t *testing.T) {
	client := k8sclient.NewMemoryClient()
	_, err := client.UpsertResource(context.Background(), "tls-certificates", consts.DefaultTLSCertificateName, map[string]any{
		"name": "default",
		"cert": "existing-cert",
		"key":  "existing-key",
	})
	require.NoError(t, err)
	_, err = client.UpsertResource(context.Background(), "domains", consts.DefaultDomainName, map[string]any{
		"name":           consts.DefaultDomainName,
		"certIdentifier": consts.DefaultTLSCertificateName,
		"enableHttps":    "force",
	})
	require.NoError(t, err)
	_, err = client.UpsertResource(context.Background(), "routes", consts.DefaultRouteName, map[string]any{
		"name":    consts.DefaultRouteName,
		"domains": []string{consts.DefaultDomainName},
		"path":    map[string]any{"matchType": "EQUAL", "matchValue": "/"},
		"services": []map[string]any{{
			"name": "existing.default.svc.cluster.local",
			"port": 8080,
		}},
	})
	require.NoError(t, err)

	svc := New(client, grafanaclient.New(grafanaclient.Config{}), portaldbclient.New(portaldbclient.Config{}))
	svc.adminUser = &response.User{
		Username:    "admin",
		DisplayName: "Admin",
	}

	require.NoError(t, svc.bootstrapDefaultResourcesLocked(context.Background(), "secret"))

	tls, err := client.GetResource(context.Background(), "tls-certificates", consts.DefaultTLSCertificateName)
	require.NoError(t, err)
	require.Equal(t, "existing-cert", tls["cert"])

	domain, err := client.GetResource(context.Background(), "domains", consts.DefaultDomainName)
	require.NoError(t, err)
	require.Equal(t, "force", domain["enableHttps"])

	route, err := client.GetResource(context.Background(), "routes", consts.DefaultRouteName)
	require.NoError(t, err)
	require.Equal(t, "existing.default.svc.cluster.local", route["services"].([]any)[0].(map[string]any)["name"])
}
