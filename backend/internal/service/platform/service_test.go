package platform

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/alibaba/aigateway-group/aigateway-console/backend/internal/consts"
	"github.com/alibaba/aigateway-group/aigateway-console/backend/internal/model/response"
	grafanaclient "github.com/alibaba/aigateway-group/aigateway-console/backend/utility/clients/grafana"
	k8sclient "github.com/alibaba/aigateway-group/aigateway-console/backend/utility/clients/k8s"
	portaldbclient "github.com/alibaba/aigateway-group/aigateway-console/backend/utility/clients/portaldb"
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
