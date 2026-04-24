package gateway

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var builtinPluginResourceDirs = []string{
	"resource/public/plugin",
	"backend/resource/public/plugin",
	"backend-java-legacy/sdk/src/main/resources/plugins",
	"backend/sdk/src/main/resources/plugins",
}

var builtinPluginFallbacks = map[string]map[string]any{
	"ai-data-masking": {
		"name":        "ai-data-masking",
		"builtIn":     true,
		"internal":    true,
		"title":       "AI Data Masking",
		"description": "Provides AI request and response masking for AI routes.",
		"category":    "ai",
		"version":     "2.0.0",
		"phase":       "AUTHN",
		"priority":    100,
		"x-title-i18n": map[string]any{
			"zh-CN": "AI 数据脱敏",
		},
		"x-description-i18n": map[string]any{
			"zh-CN": "提供 AI 路由的请求与响应脱敏能力。",
		},
	},
	"ai-quota": {
		"name":        "ai-quota",
		"builtIn":     true,
		"internal":    true,
		"title":       "AI Quota",
		"description": "Provides AI quota control and billing runtime enforcement for AI routes.",
		"category":    "ai",
		"version":     "1.1.0",
		"phase":       "AUTHN",
		"priority":    280,
		"x-title-i18n": map[string]any{
			"zh-CN": "AI 配额",
		},
		"x-description-i18n": map[string]any{
			"zh-CN": "提供 AI 路由的配额控制与计费运行时校验能力。",
		},
	},
	"cluster-key-rate-limit": {
		"name":        "cluster-key-rate-limit",
		"builtIn":     true,
		"internal":    true,
		"title":       "Cluster Key Rate Limit",
		"description": "Provides cluster-wide per-key request rate limiting for AI route runtime rules.",
		"category":    "traffic",
		"version":     "2.0.0",
		"phase":       "UNSPECIFIED_PHASE",
		"priority":    20,
		"x-title-i18n": map[string]any{
			"zh-CN": "集群 Key 限流",
		},
		"x-description-i18n": map[string]any{
			"zh-CN": "提供 AI 路由运行时的集群级按 Key 请求限流能力。",
		},
	},
	"ai-token-ratelimit": {
		"name":        "ai-token-ratelimit",
		"builtIn":     true,
		"internal":    true,
		"title":       "AI Token Ratelimit",
		"description": "Provides per-consumer AI token rate limiting for AI route runtime rules.",
		"category":    "ai",
		"version":     "2.0.0",
		"phase":       "UNSPECIFIED_PHASE",
		"priority":    600,
		"x-title-i18n": map[string]any{
			"zh-CN": "AI Token 限流",
		},
		"x-description-i18n": map[string]any{
			"zh-CN": "提供 AI 路由运行时按 consumer 的 token 限流能力。",
		},
	},
}

var aiRoutePluginDependencies = map[string][]string{}

func (s *Service) mergeWasmPlugins(items []map[string]any) []map[string]any {
	return s.mergeWasmPluginsWithLang(items, "")
}

func (s *Service) mergeWasmPluginsWithLang(items []map[string]any, lang string) []map[string]any {
	index := map[string]map[string]any{}
	for _, item := range items {
		hydrated := s.hydrateResource("wasm-plugins", item)
		index[strings.TrimSpace(strings.ToLower(stringValue(hydrated["name"])))] = hydrated
	}
	for _, builtin := range s.listBuiltinWasmPlugins() {
		key := strings.TrimSpace(strings.ToLower(stringValue(builtin["name"])))
		if key == "" {
			continue
		}
		if existing, ok := index[key]; ok {
			index[key] = mergeMaps(builtin, existing)
			continue
		}
		index[key] = builtin
	}
	result := make([]map[string]any, 0, len(index))
	for _, item := range index {
		result = append(result, s.localizeWasmPlugin(clonePayload(item), lang))
	}
	sort.Slice(result, func(i, j int) bool {
		return stringValue(result[i]["name"]) < stringValue(result[j]["name"])
	})
	return result
}

func (s *Service) listBuiltinWasmPlugins() []map[string]any {
	roots := resolveBuiltinPluginRoots()
	index := map[string]map[string]any{}
	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			key := strings.TrimSpace(strings.ToLower(entry.Name()))
			if key == "" || index[key] != nil {
				continue
			}
			if item, ok := loadBuiltinWasmPluginFromRoot(root, entry.Name()); ok {
				index[key] = item
			}
		}
	}
	for key, item := range builtinPluginFallbacks {
		if index[key] != nil {
			continue
		}
		index[key] = clonePayload(item)
	}
	items := make([]map[string]any, 0, len(index))
	for _, item := range index {
		items = append(items, item)
	}
	return items
}

func (s *Service) loadBuiltinWasmPlugin(name string) (map[string]any, bool) {
	for _, root := range resolveBuiltinPluginRoots() {
		if item, ok := loadBuiltinWasmPluginFromRoot(root, name); ok {
			return item, true
		}
	}
	if item, ok := builtinPluginFallbacks[strings.TrimSpace(strings.ToLower(name))]; ok {
		return clonePayload(item), true
	}
	return nil, false
}

func loadBuiltinWasmPluginFromRoot(base, name string) (map[string]any, bool) {
	root := filepath.Join(base, name)
	specBytes, err := os.ReadFile(filepath.Join(root, "spec.yaml"))
	if err != nil {
		return nil, false
	}
	var spec map[string]any
	if err := yaml.Unmarshal(specBytes, &spec); err != nil {
		return nil, false
	}

	info, _ := spec["info"].(map[string]any)
	specNode, _ := spec["spec"].(map[string]any)
	item := map[string]any{
		"name":        name,
		"builtIn":     true,
		"internal":    true,
		"title":       stringValue(info["title"]),
		"description": stringValue(info["description"]),
		"category":    stringValue(info["category"]),
		"icon":        stringValue(info["iconUrl"]),
		"iconUrl":     stringValue(info["iconUrl"]),
		"version":     stringValue(info["version"]),
		"phase":       strings.ToUpper(stringValue(specNode["phase"])),
		"priority":    specNode["priority"],
	}
	if titleI18n, ok := info["x-title-i18n"].(map[string]any); ok && len(titleI18n) > 0 {
		item["x-title-i18n"] = clonePayload(titleI18n)
	}
	if descI18n, ok := info["x-description-i18n"].(map[string]any); ok && len(descI18n) > 0 {
		item["x-description-i18n"] = clonePayload(descI18n)
	}
	if schema, ok := specNode["configSchema"]; ok {
		item["configSchema"] = schema
	}
	if schema, ok := specNode["routeConfigSchema"]; ok {
		item["routeConfigSchema"] = schema
		if item["configSchema"] == nil {
			item["configSchema"] = schema
		}
	}
	if readme := firstExistingTextFile(filepath.Join(root, "README_EN.md"), filepath.Join(root, "README.md")); strings.TrimSpace(readme) != "" {
		item["readme"] = readme
		item["documentation"] = readme
	}
	return item, true
}

func (s *Service) localizeWasmPlugin(item map[string]any, lang string) map[string]any {
	if item == nil {
		return nil
	}
	requiredPlugins := aiRoutePluginDependenciesForPlugin(stringValue(item["name"]))
	item["requiredPlugins"] = requiredPlugins
	item["dependencySource"] = "ai-route-explicit-map"
	if len(requiredPlugins) == 0 {
		item["dependencyEnabled"] = false
	}
	locale := normalizePluginLocale(lang)
	if locale == "" {
		return item
	}
	if title := localizedPluginText(item, "title", locale); title != "" {
		item["title"] = title
	}
	if description := localizedPluginText(item, "description", locale); description != "" {
		item["description"] = description
	}
	return item
}

func aiRoutePluginDependenciesForPlugin(pluginName string) []string {
	items := aiRoutePluginDependencies[strings.TrimSpace(pluginName)]
	if len(items) == 0 {
		return []string{}
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	sort.Strings(result)
	return result
}

func normalizePluginLocale(lang string) string {
	switch strings.TrimSpace(lang) {
	case "zh-CN", "en-US":
		return strings.TrimSpace(lang)
	default:
		return ""
	}
}

func localizedPluginText(item map[string]any, key, locale string) string {
	if item == nil || locale == "" {
		return ""
	}
	i18nMap, _ := item["x-"+key+"-i18n"].(map[string]any)
	if len(i18nMap) == 0 {
		return stringValue(item[key])
	}
	if value := stringValue(i18nMap[locale]); value != "" {
		return value
	}
	return stringValue(item[key])
}

func resolveBuiltinPluginRoots() []string {
	roots := make([]string, 0, len(builtinPluginResourceDirs))
	seen := map[string]struct{}{}
	for _, resourceDir := range builtinPluginResourceDirs {
		for _, candidate := range relativeResourceCandidates(resourceDir) {
			if isBuiltinPluginRoot(candidate) {
				cleaned := filepath.Clean(candidate)
				if _, ok := seen[cleaned]; ok {
					continue
				}
				seen[cleaned] = struct{}{}
				roots = append(roots, cleaned)
				break
			}
		}
	}
	return roots
}

func relativeResourceCandidates(resourceDir string) []string {
	candidates := []string{
		resourceDir,
		filepath.Join("..", resourceDir),
		filepath.Join("..", "..", resourceDir),
		filepath.Join("..", "..", "..", resourceDir),
		filepath.Join("..", "..", "..", "..", resourceDir),
		filepath.Join("..", "..", "..", "..", "..", resourceDir),
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		base := filepath.Dir(file)
		candidates = append(candidates,
			filepath.Join(base, resourceDir),
			filepath.Join(base, "..", resourceDir),
			filepath.Join(base, "..", "..", resourceDir),
			filepath.Join(base, "..", "..", "..", resourceDir),
			filepath.Join(base, "..", "..", "..", "..", resourceDir),
			filepath.Join(base, "..", "..", "..", "..", "..", resourceDir),
		)
	}
	return candidates
}

func isBuiltinPluginRoot(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(path, entry.Name(), "spec.yaml")); err == nil {
			return true
		}
	}
	return false
}

func firstExistingTextFile(paths ...string) string {
	for _, path := range paths {
		if bytes, err := os.ReadFile(path); err == nil {
			content := strings.TrimSpace(string(bytes))
			if content != "" {
				return content
			}
		}
	}
	return ""
}

func mergeMaps(base, override map[string]any) map[string]any {
	result := clonePayload(base)
	for key, value := range override {
		result[key] = value
	}
	return result
}
