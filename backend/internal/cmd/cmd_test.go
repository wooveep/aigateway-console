package cmd

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/stretchr/testify/require"
)

func TestManifestConfigLoads(t *testing.T) {
	ctx := context.Background()

	require.Equal(t, ":8080", g.Cfg().MustGet(ctx, "server.address").String())
	require.Equal(t, "aigateway-console", g.Cfg().MustGet(ctx, "app.name").String())
	require.NotEmpty(t, g.Cfg().MustGet(ctx, "naming.preferred").String())
}
