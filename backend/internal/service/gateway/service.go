package gateway

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/wooveep/aigateway-console/backend/internal/consts"
	k8sclient "github.com/wooveep/aigateway-console/backend/utility/clients/k8s"
)

type Service struct {
	k8sClient k8sclient.Client
}

func New(k8sClient k8sclient.Client) *Service {
	svc := &Service{k8sClient: k8sClient}
	svc.bootstrapDefaults(context.Background())
	return svc
}

func (s *Service) List(ctx context.Context, kind string) ([]map[string]any, error) {
	items, err := s.k8sClient.ListResources(ctx, kind)
	if err != nil {
		return nil, err
	}
	if kind == "wasm-plugins" {
		return s.mergeWasmPlugins(items), nil
	}
	return s.hydrateResources(kind, items), nil
}

func (s *Service) Get(ctx context.Context, kind, name string) (map[string]any, error) {
	item, err := s.k8sClient.GetResource(ctx, kind, name)
	if err == nil {
		return s.hydrateResource(kind, item), nil
	}
	if kind == "wasm-plugins" {
		if fallback, ok := s.loadBuiltinWasmPlugin(name); ok {
			return fallback, nil
		}
	}
	return nil, err
}

func (s *Service) Save(ctx context.Context, kind string, payload map[string]any) (map[string]any, error) {
	payload = clonePayload(payload)
	name := strings.TrimSpace(fmt.Sprint(payload["name"]))
	if name == "" {
		return nil, errors.New("name is required")
	}
	if s.isProtected(kind, name) {
		return nil, fmt.Errorf("%s %s is protected", kind, name)
	}
	if s.isInternalWriteBlocked(kind, name) {
		return nil, fmt.Errorf("%s %s is an internal resource", kind, name)
	}
	normalized, err := s.normalizeForSave(ctx, kind, payload)
	if err != nil {
		return nil, err
	}
	item, err := s.k8sClient.UpsertResource(ctx, kind, name, normalized)
	if err != nil {
		return nil, err
	}
	return s.hydrateResource(kind, item), nil
}

func (s *Service) Delete(ctx context.Context, kind, name string) error {
	if s.isProtected(kind, name) {
		return fmt.Errorf("%s %s is protected", kind, name)
	}
	if s.isInternalWriteBlocked(kind, name) {
		return fmt.Errorf("%s %s is an internal resource", kind, name)
	}
	return s.k8sClient.DeleteResource(ctx, kind, name)
}

func (s *Service) ListMcpConsumers(ctx context.Context, serverName string) ([]map[string]any, error) {
	item, err := s.Get(ctx, "mcp-servers", serverName)
	if err != nil {
		return nil, err
	}
	info, _ := item["consumerAuthInfo"].(map[string]any)
	allowed, _ := info["allowedConsumers"].([]any)

	result := make([]map[string]any, 0, len(allowed))
	for _, consumer := range allowed {
		result = append(result, map[string]any{
			"mcpServerName": serverName,
			"consumerName":  fmt.Sprint(consumer),
			"type":          fmt.Sprint(info["type"]),
		})
	}
	return result, nil
}

func (s *Service) SwaggerToMCPConfig(content string) string {
	return strings.TrimSpace(fmt.Sprintf("type: OPEN_API\nrawConfigurations: |\n  %s\n", strings.ReplaceAll(content, "\n", "\n  ")))
}

func (s *Service) GetPluginInstance(ctx context.Context, scope, target, pluginName string) (map[string]any, error) {
	return s.k8sClient.GetResource(ctx, pluginInstanceKind(scope, target), pluginName)
}

func (s *Service) SavePluginInstance(ctx context.Context, scope, target, pluginName string, payload map[string]any) (map[string]any, error) {
	if err := s.validatePluginInstance(ctx, scope, target, pluginName, payload); err != nil {
		return nil, err
	}
	payload["name"] = pluginName
	return s.k8sClient.UpsertResource(ctx, pluginInstanceKind(scope, target), pluginName, payload)
}

func (s *Service) DeletePluginInstance(ctx context.Context, scope, target, pluginName string) error {
	return s.k8sClient.DeleteResource(ctx, pluginInstanceKind(scope, target), pluginName)
}

func (s *Service) ListPluginInstances(ctx context.Context, scope, target string) ([]map[string]any, error) {
	return s.k8sClient.ListResources(ctx, pluginInstanceKind(scope, target))
}

func (s *Service) GetWasmPluginConfig(ctx context.Context, name string) (map[string]any, error) {
	item, err := s.Get(ctx, "wasm-plugins", name)
	if err != nil {
		return nil, err
	}
	config := map[string]any{
		"name": name,
		"type": "yaml",
	}
	if schema, ok := item["configSchema"]; ok {
		config["schema"] = schema
	} else if schema, ok := item["schema"]; ok {
		config["schema"] = schema
	} else {
		config["schema"] = map[string]any{}
	}
	return config, nil
}

func (s *Service) GetWasmPluginReadme(ctx context.Context, name string) (string, error) {
	item, err := s.Get(ctx, "wasm-plugins", name)
	if err != nil {
		return "", err
	}
	readme := stringValue(item["readme"])
	if readme == "" {
		readme = stringValue(item["documentation"])
	}
	if readme == "" {
		description := stringValue(item["description"])
		if description != "" {
			readme = "# " + name + "\n\n" + description
		}
	}
	if readme == "" {
		readme = "# " + name + "\n\nNo readme metadata is available yet for this plugin."
	}
	return readme, nil
}

func (s *Service) bootstrapDefaults(ctx context.Context) {
	_, _ = s.k8sClient.UpsertResource(ctx, "services", "aigateway-console.dns", map[string]any{
		"name":         "aigateway-console.dns",
		"namespace":    "aigateway-system",
		"port":         8080,
		"endpoints":    []string{"aigateway-console.aigateway-system.svc.cluster.local:8080"},
		"ingressClass": s.k8sClient.IngressClass(),
	})
	_, _ = s.k8sClient.UpsertResource(ctx, "proxy-servers", "default-http-proxy", map[string]any{
		"name":           "default-http-proxy",
		"type":           "HTTP",
		"serverAddress":  "127.0.0.1",
		"serverPort":     3128,
		"connectTimeout": 5000,
	})
}

func (s *Service) isProtected(kind, name string) bool {
	switch kind {
	case "domains":
		return name == consts.DefaultDomainName
	case "routes":
		return name == consts.DefaultRouteName
	case "tls-certificates":
		return name == consts.DefaultTLSCertificateName
	default:
		return false
	}
}

func (s *Service) isInternalWriteBlocked(kind, name string) bool {
	switch kind {
	case "routes", "service-sources", "proxy-servers":
		return strings.HasSuffix(name, consts.InternalResourceNameSuffix)
	default:
		return false
	}
}

func pluginInstanceKind(scope, target string) string {
	scope = strings.TrimSpace(scope)
	target = strings.TrimSpace(target)
	if target == "" {
		return fmt.Sprintf("%s-plugin-instances", scope)
	}
	return fmt.Sprintf("%s-plugin-instances:%s", scope, target)
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func clonePayload(src map[string]any) map[string]any {
	if src == nil {
		return map[string]any{}
	}
	return clonePayloadMap(src)
}
