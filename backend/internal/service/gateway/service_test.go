package gateway

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	k8sclient "github.com/alibaba/aigateway-group/aigateway-console/backend/utility/clients/k8s"
)

func TestProtectedResourceDeleteIsRejected(t *testing.T) {
	svc := New(k8sclient.NewMemoryClient())

	err := svc.Delete(context.Background(), "routes", "default")
	require.Error(t, err)
	require.Contains(t, err.Error(), "protected")
}

func TestInternalResourceWriteIsRejected(t *testing.T) {
	svc := New(k8sclient.NewMemoryClient())

	_, err := svc.Save(context.Background(), "proxy-servers", map[string]any{
		"name": "mesh.internal",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "internal resource")
}

func TestPluginInstanceRoundTrip(t *testing.T) {
	svc := New(k8sclient.NewMemoryClient())

	_, err := svc.SavePluginInstance(context.Background(), "route", "demo-route", "key-auth", map[string]any{
		"config": map[string]any{"consumers": []string{"demo"}},
	})
	require.NoError(t, err)

	item, err := svc.GetPluginInstance(context.Background(), "route", "demo-route", "key-auth")
	require.NoError(t, err)
	require.Equal(t, "key-auth", item["name"])

	list, err := svc.ListPluginInstances(context.Background(), "route", "demo-route")
	require.NoError(t, err)
	require.Len(t, list, 1)
}
