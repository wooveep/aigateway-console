package portal

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	k8sclient "github.com/wooveep/aigateway-console/backend/utility/clients/k8s"
	portaldbclient "github.com/wooveep/aigateway-console/backend/utility/clients/portaldb"
)

func TestReplaceAssetGrants(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	expectSchema(mock)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM asset_grant WHERE asset_type = ? AND asset_id = ?`)).
		WithArgs("model_binding", "binding-1").
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta(`
			INSERT INTO asset_grant (asset_type, asset_id, subject_type, subject_id)
			VALUES (?, ?, ?, ?)`)).
		WithArgs("model_binding", "binding-1", "consumer", "demo").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`
			INSERT INTO asset_grant (asset_type, asset_id, subject_type, subject_id)
			VALUES (?, ?, ?, ?)`)).
		WithArgs("model_binding", "binding-1", "user_level", "pro").
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT asset_type, asset_id, subject_type, subject_id
		FROM asset_grant
		WHERE asset_type = ? AND asset_id = ?
		ORDER BY subject_type, subject_id`)).
		WithArgs("model_binding", "binding-1").
		WillReturnRows(sqlmock.NewRows([]string{"asset_type", "asset_id", "subject_type", "subject_id"}).
			AddRow("model_binding", "binding-1", "consumer", "demo").
			AddRow("model_binding", "binding-1", "user_level", "pro"))

	svc := New(portaldbclient.NewFromDB(portaldbclient.Config{Enabled: true, Driver: "mysql", AutoMigrate: true}, db))
	items, err := svc.ReplaceAssetGrants(context.Background(), "model_binding", "binding-1", []AssetGrantRecord{
		{SubjectType: "consumer", SubjectID: "demo"},
		{SubjectType: "user_level", SubjectID: "pro"},
	})
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAgentCatalogOptionsFromK8s(t *testing.T) {
	client := k8sclient.NewMemoryClient()
	_, err := client.UpsertResource(context.Background(), "mcp-servers", "knowledge-hub", map[string]any{
		"name":        "knowledge-hub",
		"description": "Search and docs",
		"type":        "OPEN_API",
		"domains":     []any{"docs.internal"},
		"consumerAuthInfo": map[string]any{
			"enable": true,
			"type":   "key-auth",
		},
		"updatedAt": time.Now().UTC().Format(time.RFC3339),
	})
	require.NoError(t, err)

	svc := New(nil, client)
	options, err := svc.GetAgentCatalogOptions(context.Background())
	require.NoError(t, err)
	require.Len(t, options.Servers, 1)
	require.Equal(t, "knowledge-hub", options.Servers[0].McpServerName)
	require.Equal(t, "key-auth", options.Servers[0].AuthType)
	require.NotNil(t, options.Servers[0].AuthEnabled)
	require.True(t, *options.Servers[0].AuthEnabled)
}

func TestGetModelAssetOptionsSupportsStructuredProviderModels(t *testing.T) {
	client := k8sclient.NewMemoryClient()
	_, err := client.UpsertResource(context.Background(), "ai-providers", "doubao", map[string]any{
		"name": "doubao",
		"models": []map[string]any{
			{"modelId": "doubao-seed-2-0-pro-260215", "targetModel": "doubao-seed-2-0-pro-260215", "label": "Doubao Pro"},
		},
	})
	require.NoError(t, err)

	svc := New(nil, client)
	options, err := svc.GetModelAssetOptions(context.Background())
	require.NoError(t, err)
	require.Len(t, options.ProviderModels, 1)
	require.Equal(t, "doubao", options.ProviderModels[0].ProviderName)
	require.Len(t, options.ProviderModels[0].Models, 1)
	require.Equal(t, "doubao-seed-2-0-pro-260215", options.ProviderModels[0].Models[0].ModelID)
}
