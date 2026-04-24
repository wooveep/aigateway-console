package k8s

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/wooveep/aigateway-console/backend/internal/consts"
)

const (
	resourceConfigMapLabelKey  = "console.aigateway.io/type"
	resourceConfigMapLabelKind = "console.aigateway.io/kind"
	resourceConfigMapLabelApp  = "app.kubernetes.io/managed-by"
	resourceConfigMapType      = "resource"
	resourceConfigMapAppValue  = "aigateway-console"
	resourcePayloadKey         = "payload.json"
	defaultKubectlBin          = "kubectl"
	defaultK8sNamespace        = "aigateway-system"
	defaultResourcePrefix      = "aigw-console"
)

var ErrNotFound = errors.New("resource not found")

type Config struct {
	Enabled        bool
	Namespace      string
	KubectlBin     string
	KubeconfigPath string
	ResourcePrefix string
	IngressClass   string
}

type Client interface {
	Healthy(ctx context.Context) error
	Namespace() string
	IngressClass() string
	ReadSecret(ctx context.Context, name string) (map[string]string, error)
	UpsertSecret(ctx context.Context, name string, data map[string]string) error
	ReadConfigMap(ctx context.Context, name string) (map[string]string, error)
	UpsertConfigMap(ctx context.Context, name string, data map[string]string) error
	ListResources(ctx context.Context, kind string) ([]map[string]any, error)
	GetResource(ctx context.Context, kind, name string) (map[string]any, error)
	UpsertResource(ctx context.Context, kind, name string, data map[string]any) (map[string]any, error)
	DeleteResource(ctx context.Context, kind, name string) error
	SyncAIDataMaskingRuntime(ctx context.Context) error
	SyncAIModelRateLimitRuntime(ctx context.Context) error
	ResolveAIQuotaRedisServiceName(ctx context.Context) string
	ResolveAIQuotaRedisPassword(ctx context.Context) string
}

type RealClient struct {
	namespace      string
	kubectlBin     string
	kubeconfigPath string
	resourcePrefix string
	ingressClass   string
}

type MemoryClient struct {
	mu           sync.RWMutex
	secrets      map[string]map[string]string
	configMaps   map[string]map[string]string
	resources    map[string]map[string]map[string]any
	namespace    string
	ingressClass string
}

func New(cfg Config) Client {
	if cfg.Enabled {
		return &RealClient{
			namespace:      firstNonEmpty(cfg.Namespace, defaultK8sNamespace),
			kubectlBin:     firstNonEmpty(cfg.KubectlBin, defaultKubectlBin),
			kubeconfigPath: strings.TrimSpace(cfg.KubeconfigPath),
			resourcePrefix: firstNonEmpty(cfg.ResourcePrefix, defaultResourcePrefix),
			ingressClass:   firstNonEmpty(cfg.IngressClass, "aigateway"),
		}
	}
	return NewMemoryClient(Config{
		Namespace:    cfg.Namespace,
		IngressClass: cfg.IngressClass,
	})
}

func NewMemoryClient(configs ...Config) *MemoryClient {
	cfg := Config{}
	if len(configs) > 0 {
		cfg = configs[0]
	}
	client := &MemoryClient{
		secrets:      map[string]map[string]string{},
		configMaps:   map[string]map[string]string{},
		resources:    map[string]map[string]map[string]any{},
		namespace:    firstNonEmpty(cfg.Namespace, defaultK8sNamespace),
		ingressClass: firstNonEmpty(cfg.IngressClass, "aigateway"),
	}
	client.configMaps[consts.DefaultHigressConfigMapName] = map[string]string{
		"resourceVersion": "1",
		consts.DefaultHigressConfigDataKey: `tracing:
  enable: false
  sampling: 100
  timeout: 500
gzip:
  enable: true
  minContentLength: 1024
  contentType:
    - text/html
    - text/css
    - text/plain
    - text/xml
    - application/json
    - application/javascript
    - application/xhtml+xml
    - image/svg+xml
  disableOnEtagHeader: true
  memoryLevel: 5
  windowBits: 12
  chunkSize: 4096
  compressionLevel: BEST_COMPRESSION
  compressionStrategy: DEFAULT_STRATEGY
downstream:
  idleTimeout: 180
  maxRequestHeadersKb: 60
  connectionBufferLimits: 32768
  http2:
    maxConcurrentStreams: 100
    initialStreamWindowSize: 65535
    initialConnectionWindowSize: 1048576
  routeTimeout: 0
upstream:
  idleTimeout: 10
  connectionBufferLimits: 10485760
addXRealIpHeader: false
disableXEnvoyHeaders: false
`,
	}
	return client
}

func (c *MemoryClient) Healthy(ctx context.Context) error { return nil }
func (c *MemoryClient) Namespace() string                 { return c.namespace }
func (c *MemoryClient) IngressClass() string              { return c.ingressClass }

func (c *MemoryClient) ReadSecret(ctx context.Context, name string) (map[string]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, ok := c.secrets[name]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneStringMap(data), nil
}

func (c *MemoryClient) UpsertSecret(ctx context.Context, name string, data map[string]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.secrets[name] = cloneStringMap(data)
	return nil
}

func (c *MemoryClient) ReadConfigMap(ctx context.Context, name string) (map[string]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, ok := c.configMaps[name]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneStringMap(data), nil
}

func (c *MemoryClient) UpsertConfigMap(ctx context.Context, name string, data map[string]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.configMaps[name] = cloneStringMap(data)
	return nil
}

func (c *MemoryClient) ListResources(ctx context.Context, kind string) ([]map[string]any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	items := make([]map[string]any, 0)
	for _, item := range c.resources[kind] {
		items = append(items, cloneMap(item))
	}
	sort.Slice(items, func(i, j int) bool {
		return fmt.Sprint(items[i]["name"]) < fmt.Sprint(items[j]["name"])
	})
	return items, nil
}

func (c *MemoryClient) GetResource(ctx context.Context, kind, name string) (map[string]any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if item, ok := c.resources[kind][name]; ok {
		return cloneMap(item), nil
	}
	return nil, ErrNotFound
}

func (c *MemoryClient) UpsertResource(ctx context.Context, kind, name string, data map[string]any) (map[string]any, error) {
	c.mu.Lock()

	if _, ok := c.resources[kind]; !ok {
		c.resources[kind] = map[string]map[string]any{}
	}
	merged := cloneMap(data)
	merged["name"] = name
	if merged["version"] == nil {
		merged["version"] = nextVersion(c.resources[kind])
	}
	c.resources[kind][name] = merged
	c.mu.Unlock()

	if shouldSyncAIDataMaskingRuntime(kind, name) {
		if err := c.SyncAIDataMaskingRuntime(ctx); err != nil {
			return nil, err
		}
	}
	if routeName, ok := shouldSyncAIRouteManagedBuiltinRuntime(kind, name); ok {
		if err := c.SyncAIRouteManagedBuiltinRuntime(ctx, routeName); err != nil {
			return nil, err
		}
	}
	if shouldSyncAIModelRateLimitRuntime(kind, name) {
		if err := c.SyncAIModelRateLimitRuntime(ctx); err != nil {
			return nil, err
		}
	}
	return cloneMap(merged), nil
}

func (c *MemoryClient) DeleteResource(ctx context.Context, kind, name string) error {
	c.mu.Lock()

	if _, ok := c.resources[kind][name]; !ok {
		c.mu.Unlock()
		return ErrNotFound
	}
	delete(c.resources[kind], name)
	c.mu.Unlock()

	if shouldSyncAIDataMaskingRuntime(kind, name) {
		return c.SyncAIDataMaskingRuntime(ctx)
	}
	if routeName, ok := shouldSyncAIRouteManagedBuiltinRuntime(kind, name); ok {
		return c.SyncAIRouteManagedBuiltinRuntime(ctx, routeName)
	}
	if shouldSyncAIModelRateLimitRuntime(kind, name) {
		return c.SyncAIModelRateLimitRuntime(ctx)
	}
	return nil
}

func (c *MemoryClient) LoadBuiltinPluginRules(ctx context.Context, pluginName string) (map[string]map[string]any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := map[string]map[string]any{}
	items := c.resources[higressWasmPluginResource]
	for _, item := range items {
		labels, _ := nestedValue(item, "metadata", "labels").(map[string]any)
		if strings.TrimSpace(fmt.Sprint(labels[higressLabelWasmPluginName])) != strings.TrimSpace(pluginName) &&
			strings.TrimSpace(fmt.Sprint(item["name"])) != strings.TrimSpace(pluginName) {
			continue
		}
		for _, rule := range toMapSlice(nestedValue(item, "spec", "matchRules")) {
			for _, ingress := range normalizeStringSlice(rule["ingress"]) {
				result[ingress] = cloneMap(rule)
			}
		}
	}
	return result, nil
}

func (c *MemoryClient) UpsertBuiltinWasmPluginExecution(ctx context.Context, pluginName, phase string, priority int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	internalName := builtinWasmPluginResourceName(pluginName)
	existing := cloneMap(c.resources[higressWasmPluginResource][internalName])
	if existing == nil {
		manifest, ok := builtinWasmPluginManifest(pluginName, c.namespace)
		if !ok {
			return ErrNotFound
		}
		existing = manifest
	}
	spec := ensureMap(existing, "spec")
	if strings.TrimSpace(phase) != "" {
		spec["phase"] = strings.TrimSpace(phase)
	}
	if priority > 0 {
		spec["priority"] = priority
	}
	if _, ok := c.resources[higressWasmPluginResource]; !ok {
		c.resources[higressWasmPluginResource] = map[string]map[string]any{}
	}
	c.resources[higressWasmPluginResource][internalName] = existing
	return nil
}

func (c *MemoryClient) SyncAIRouteManagedBuiltinRuntime(ctx context.Context, routeName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	publicIngress := aiRouteIngressName(routeName)
	internalIngress := aiRouteInternalIngressName(routeName)
	if _, ok := c.resources["ai-routes"][routeName]; !ok {
		_ = c.removeBuiltinPluginRuleLocked(higressWasmPluginNameAIStatistics, map[string][]string{"ingress": {publicIngress}})
		_ = c.removeBuiltinPluginRuleLocked(higressWasmPluginNameAIStatistics, map[string][]string{"ingress": {internalIngress}})
		_ = c.removeBuiltinPluginRuleLocked(higressWasmPluginNameAIQuota, map[string][]string{"ingress": {publicIngress}})
		_ = c.removeBuiltinPluginRuleLocked(higressWasmPluginNameAIQuota, map[string][]string{"ingress": {internalIngress}})
		return nil
	}

	if err := c.syncAIRouteManagedRuleLocked(routeName, higressWasmPluginNameAIStatistics, publicIngress, true, map[string]any{
		higressAIStatisticsDefaultAttrsKey: true,
	}); err != nil {
		return err
	}
	if err := c.syncAIRouteManagedRuleLocked(routeName, higressWasmPluginNameAIStatistics, internalIngress, true, map[string]any{
		higressAIStatisticsDefaultAttrsKey: true,
	}); err != nil {
		return err
	}

	quotaConfig := buildAIQuotaRuleConfig(routeName, c.ResolveAIQuotaRedisServiceName(ctx), c.ResolveAIQuotaRedisPassword(ctx))
	if err := c.syncAIRouteManagedRuleLocked(routeName, higressWasmPluginNameAIQuota, publicIngress, true, quotaConfig); err != nil {
		return err
	}
	if err := c.syncAIRouteManagedRuleLocked(routeName, higressWasmPluginNameAIQuota, internalIngress, true, quotaConfig); err != nil {
		return err
	}

	return nil
}

func (c *MemoryClient) SyncAIDataMaskingRuntime(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	projection := cloneMap(c.resources["ai-sensitive-projections"]["default"])
	desiredRules := c.buildDesiredAIDataMaskingRulesLocked(projection)
	pluginName := builtinWasmPluginResourceName(higressWasmPluginNameAIDataMasking)
	existing := cloneMap(c.resources[higressWasmPluginResource][pluginName])
	if existing == nil && len(desiredRules) == 0 {
		return nil
	}
	if existing == nil {
		manifest, ok := builtinWasmPluginManifest(higressWasmPluginNameAIDataMasking, c.namespace)
		if !ok {
			return ErrNotFound
		}
		existing = manifest
	}
	spec := ensureMap(existing, "spec")
	spec["matchRules"] = syncAIDataMaskingMatchRules(toMapSlice(spec["matchRules"]), desiredRules)

	if _, ok := c.resources[higressWasmPluginResource]; !ok {
		c.resources[higressWasmPluginResource] = map[string]map[string]any{}
	}
	c.resources[higressWasmPluginResource][pluginName] = existing
	return nil
}

func (c *MemoryClient) SyncAIModelRateLimitRuntime(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	projection := cloneMap(c.resources["ai-model-rate-limit-projections"]["default"])
	desiredRules := buildAIModelRateLimitRulesFromProjection(projection)
	for pluginName, desired := range desiredRules {
		internalName := builtinWasmPluginResourceName(pluginName)
		existing := cloneMap(c.resources[higressWasmPluginResource][internalName])
		if existing == nil && len(desired) == 0 {
			continue
		}
		if existing == nil {
			manifest, ok := builtinWasmPluginManifest(pluginName, c.namespace)
			if !ok {
				return ErrNotFound
			}
			existing = manifest
		}
		spec := ensureMap(existing, "spec")
		spec["matchRules"] = syncAIModelRateLimitMatchRules(toMapSlice(spec["matchRules"]), desired, modelRateLimitRulePrefix(pluginName))

		if _, ok := c.resources[higressWasmPluginResource]; !ok {
			c.resources[higressWasmPluginResource] = map[string]map[string]any{}
		}
		c.resources[higressWasmPluginResource][internalName] = existing
	}
	return nil
}

func (c *MemoryClient) ResolveAIQuotaRedisServiceName(ctx context.Context) string {
	return fmt.Sprintf("%s.%s.svc.%s", higressAIQuotaRedisServiceDefault, c.namespace, higressPluginServerClusterDomainDefault)
}

func (c *MemoryClient) ResolveAIQuotaRedisPassword(ctx context.Context) string {
	secret, ok := c.secrets[higressAIQuotaRedisSecretDefault]
	if ok {
		if password := strings.TrimSpace(secret[higressAIQuotaRedisPasswordKey]); password != "" {
			return password
		}
	}
	return higressAIQuotaRedisPasswordDefault
}

func (c *MemoryClient) syncAIRouteManagedRuleLocked(
	routeName string,
	pluginName string,
	ingressName string,
	defaultEnabled bool,
	baseConfig map[string]any,
) error {
	override := c.loadAIRoutePluginInstanceOverrideLocked(routeName, pluginName)
	config := cloneMap(baseConfig)
	if override != nil && len(override.Config) > 0 {
		for key, value := range override.Config {
			config[key] = value
		}
	}
	enabled := defaultEnabled
	if override != nil {
		enabled = override.Enabled
	}
	if !enabled {
		_ = c.removeBuiltinPluginRuleLocked(pluginName, map[string][]string{"ingress": {ingressName}})
		return nil
	}
	return c.upsertBuiltinPluginRuleLocked(pluginName, map[string][]string{"ingress": {ingressName}}, config, true)
}

func (c *MemoryClient) loadAIRoutePluginInstanceOverrideLocked(routeName, pluginName string) *aiRoutePluginInstanceOverride {
	target := aiRouteIngressName(routeName)
	item := cloneMap(c.resources[routePluginInstanceKind(target)][pluginName])
	if item == nil {
		return nil
	}
	return &aiRoutePluginInstanceOverride{
		Enabled: boolValue(firstNonNil(item["enabled"], item["runtimeEnabled"])),
		Config:  parsePluginInstanceOverrideConfig(item),
	}
}

func (c *MemoryClient) upsertBuiltinPluginRuleLocked(pluginName string, targets map[string][]string, config map[string]any, enabled bool) error {
	return c.mutateBuiltinWasmPluginLocked(pluginName, func(plugin map[string]any) error {
		spec := ensureMap(plugin, "spec")
		matchRules := toMapSlice(spec["matchRules"])
		nextRule := map[string]any{
			"config":        cloneMap(config),
			"configDisable": !enabled,
		}
		for targetType, targetValues := range targets {
			if len(targetValues) > 0 {
				nextRule[targetType] = targetValues
			}
		}
		next := make([]map[string]any, 0, len(matchRules)+1)
		replaced := false
		for _, rule := range matchRules {
			if wasmRuleMatchesTargets(rule, targets) {
				if !replaced {
					next = append(next, nextRule)
					replaced = true
				}
				continue
			}
			next = append(next, rule)
		}
		if !replaced {
			next = append(next, nextRule)
		}
		sort.Slice(next, func(i, j int) bool {
			return wasmRuleLess(next[i], next[j])
		})
		spec["matchRules"] = next
		return nil
	})
}

func (c *MemoryClient) removeBuiltinPluginRuleLocked(pluginName string, targets map[string][]string) error {
	return c.mutateBuiltinWasmPluginLocked(pluginName, func(plugin map[string]any) error {
		spec := ensureMap(plugin, "spec")
		matchRules := toMapSlice(spec["matchRules"])
		next := make([]map[string]any, 0, len(matchRules))
		found := false
		for _, rule := range matchRules {
			if wasmRuleMatchesTargets(rule, targets) {
				found = true
				continue
			}
			next = append(next, rule)
		}
		if !found {
			return ErrNotFound
		}
		spec["matchRules"] = next
		return nil
	})
}

func (c *MemoryClient) mutateBuiltinWasmPluginLocked(pluginName string, mutate func(map[string]any) error) error {
	internalName := builtinWasmPluginResourceName(pluginName)
	existing := cloneMap(c.resources[higressWasmPluginResource][internalName])
	if existing == nil {
		manifest, ok := builtinWasmPluginManifest(pluginName, c.namespace)
		if !ok {
			return ErrNotFound
		}
		existing = manifest
	}
	syncBuiltinWasmPluginBaseSpec(existing, pluginName, c.namespace)
	if err := mutate(existing); err != nil {
		return err
	}
	if _, ok := c.resources[higressWasmPluginResource]; !ok {
		c.resources[higressWasmPluginResource] = map[string]map[string]any{}
	}
	c.resources[higressWasmPluginResource][internalName] = existing
	return nil
}

func (c *RealClient) Healthy(ctx context.Context) error {
	_, err := c.run(ctx, nil, "get", "namespace", c.namespace, "-o", "name")
	return err
}

func (c *RealClient) Namespace() string {
	return c.namespace
}

func (c *RealClient) IngressClass() string {
	return c.ingressClass
}

func (c *RealClient) ReadSecret(ctx context.Context, name string) (map[string]string, error) {
	body, err := c.run(ctx, nil, "get", "secret", name, "-o", "json")
	if err != nil {
		return nil, err
	}
	var payload struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	result := map[string]string{}
	for key, value := range payload.Data {
		decoded, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, err
		}
		result[key] = string(decoded)
	}
	return result, nil
}

func (c *RealClient) UpsertSecret(ctx context.Context, name string, data map[string]string) error {
	manifest, err := yaml.Marshal(map[string]any{
		"apiVersion": "v1",
		"kind":       "Secret",
		"metadata": map[string]any{
			"name":      name,
			"namespace": c.namespace,
		},
		"type":       "Opaque",
		"stringData": cloneStringMap(data),
	})
	if err != nil {
		return err
	}
	_, err = c.run(ctx, manifest, "apply", "-f", "-")
	return err
}

func (c *RealClient) ReadConfigMap(ctx context.Context, name string) (map[string]string, error) {
	body, err := c.run(ctx, nil, "get", "configmap", name, "-o", "json")
	if err != nil {
		return nil, err
	}
	var payload struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	return cloneStringMap(payload.Data), nil
}

func (c *RealClient) UpsertConfigMap(ctx context.Context, name string, data map[string]string) error {
	manifest, err := yaml.Marshal(map[string]any{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]any{
			"name":      name,
			"namespace": c.namespace,
		},
		"data": cloneStringMap(data),
	})
	if err != nil {
		return err
	}
	_, err = c.run(ctx, manifest, "apply", "-f", "-")
	return err
}

func (c *RealClient) ListResources(ctx context.Context, kind string) ([]map[string]any, error) {
	if isControlPlaneKind(kind) {
		return c.listControlPlaneResources(ctx, kind)
	}
	kindSlug := labelValue(kind)
	body, err := c.run(ctx, nil, "get", "configmap", "-l",
		fmt.Sprintf("%s=%s,%s=%s", resourceConfigMapLabelKey, resourceConfigMapType, resourceConfigMapLabelKind, kindSlug),
		"-o", "json")
	if err != nil {
		return nil, err
	}
	var payload struct {
		Items []struct {
			Data map[string]string `json:"data"`
		} `json:"items"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(payload.Items))
	for _, item := range payload.Items {
		resource, err := decodeResourcePayload(item.Data)
		if err != nil {
			return nil, err
		}
		items = append(items, resource)
	}
	sort.Slice(items, func(i, j int) bool {
		return fmt.Sprint(items[i]["name"]) < fmt.Sprint(items[j]["name"])
	})
	return items, nil
}

func (c *RealClient) GetResource(ctx context.Context, kind, name string) (map[string]any, error) {
	if isControlPlaneKind(kind) {
		return c.getControlPlaneResource(ctx, kind, name)
	}
	body, err := c.run(ctx, nil, "get", "configmap", c.resourceConfigMapName(kind, name), "-o", "json")
	if err != nil {
		return nil, err
	}
	var payload struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	return decodeResourcePayload(payload.Data)
}

func (c *RealClient) UpsertResource(ctx context.Context, kind, name string, data map[string]any) (map[string]any, error) {
	if isControlPlaneKind(kind) {
		return c.upsertControlPlaneResource(ctx, kind, name, data)
	}
	item := cloneMap(data)
	item["name"] = name
	if item["version"] == nil {
		existing, err := c.GetResource(ctx, kind, name)
		if err == nil {
			item["version"] = incrementVersion(existing["version"])
		} else if !errors.Is(err, ErrNotFound) {
			return nil, err
		} else {
			item["version"] = "1"
		}
	}

	payloadBytes, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	manifest, err := yaml.Marshal(map[string]any{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]any{
			"name":      c.resourceConfigMapName(kind, name),
			"namespace": c.namespace,
			"labels": map[string]string{
				resourceConfigMapLabelKey:  resourceConfigMapType,
				resourceConfigMapLabelKind: labelValue(kind),
				resourceConfigMapLabelApp:  resourceConfigMapAppValue,
			},
			"annotations": map[string]string{
				"console.aigateway.io/original-kind": kind,
				"console.aigateway.io/original-name": name,
			},
		},
		"data": map[string]string{
			resourcePayloadKey: string(payloadBytes),
		},
	})
	if err != nil {
		return nil, err
	}
	if _, err := c.run(ctx, manifest, "apply", "-f", "-"); err != nil {
		return nil, err
	}
	if shouldSyncAIDataMaskingRuntime(kind, name) {
		if err := c.SyncAIDataMaskingRuntime(ctx); err != nil {
			return nil, err
		}
	}
	if shouldSyncAIModelRateLimitRuntime(kind, name) {
		if err := c.SyncAIModelRateLimitRuntime(ctx); err != nil {
			return nil, err
		}
	}
	return cloneMap(item), nil
}

func (c *RealClient) DeleteResource(ctx context.Context, kind, name string) error {
	if isControlPlaneKind(kind) {
		return c.deleteControlPlaneResource(ctx, kind, name)
	}
	_, err := c.run(ctx, nil, "delete", "configmap", c.resourceConfigMapName(kind, name), "--ignore-not-found=false")
	if err != nil {
		return err
	}
	if shouldSyncAIDataMaskingRuntime(kind, name) {
		return c.SyncAIDataMaskingRuntime(ctx)
	}
	if shouldSyncAIModelRateLimitRuntime(kind, name) {
		return c.SyncAIModelRateLimitRuntime(ctx)
	}
	return nil
}

func shouldSyncAIDataMaskingRuntime(kind, name string) bool {
	trimmedKind := strings.TrimSpace(kind)
	trimmedName := strings.TrimSpace(name)
	if trimmedKind == "ai-sensitive-projections" && trimmedName == "default" {
		return true
	}
	if strings.HasPrefix(trimmedKind, "route-plugin-instances:") && trimmedName == higressWasmPluginNameAIDataMasking {
		return true
	}
	return false
}

func shouldSyncAIModelRateLimitRuntime(kind, name string) bool {
	trimmedKind := strings.TrimSpace(kind)
	trimmedName := strings.TrimSpace(name)
	if trimmedKind == "ai-model-rate-limit-projections" && trimmedName == "default" {
		return true
	}
	if strings.HasPrefix(trimmedKind, "route-plugin-instances:") &&
		(trimmedName == higressWasmPluginNameClusterKeyRateLimit || trimmedName == higressWasmPluginNameAITokenRateLimit) {
		return true
	}
	return false
}

func shouldSyncAIRouteManagedBuiltinRuntime(kind, name string) (string, bool) {
	trimmedKind := strings.TrimSpace(kind)
	trimmedName := strings.TrimSpace(name)
	if !strings.HasPrefix(trimmedKind, "route-plugin-instances:") {
		return "", false
	}
	if trimmedName != higressWasmPluginNameAIStatistics && trimmedName != higressWasmPluginNameAIQuota {
		return "", false
	}
	target := strings.TrimPrefix(trimmedKind, "route-plugin-instances:")
	return aiRouteNameFromPluginTarget(target)
}

func ParseConfigMapYAML(raw string) (map[string]string, error) {
	var parsed map[string]any
	if err := yaml.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}

	metadata, _ := parsed["metadata"].(map[string]any)
	dataNode, _ := parsed["data"].(map[string]any)
	if len(dataNode) == 0 {
		return nil, errors.New("config map data is empty")
	}

	result := map[string]string{}
	if metadata != nil {
		if rv, ok := metadata["resourceVersion"].(string); ok {
			result["resourceVersion"] = rv
		}
	}
	for key, value := range dataNode {
		result[key] = fmt.Sprint(value)
	}
	return result, nil
}

func RenderConfigMapYAML(name string, data map[string]string) (string, error) {
	payload := map[string]any{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]any{
			"name":            name,
			"resourceVersion": data["resourceVersion"],
		},
		"data": map[string]any{
			"aigateway":    data["aigateway"],
			"mesh":         data["mesh"],
			"meshNetworks": data["meshNetworks"],
		},
	}
	bytes, err := yaml.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func cloneStringMap(src map[string]string) map[string]string {
	dst := map[string]string{}
	maps.Copy(dst, src)
	return dst
}

func cloneMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	bytes, _ := json.Marshal(src)
	dst := map[string]any{}
	_ = json.Unmarshal(bytes, &dst)
	return dst
}

func nextVersion(items map[string]map[string]any) string {
	return fmt.Sprintf("%d", len(items)+1)
}

func (c *RealClient) run(ctx context.Context, stdin []byte, args ...string) ([]byte, error) {
	cmdArgs := make([]string, 0, len(args)+4)
	if c.kubeconfigPath != "" {
		cmdArgs = append(cmdArgs, "--kubeconfig", c.kubeconfigPath)
	}
	if c.namespace != "" {
		cmdArgs = append(cmdArgs, "-n", c.namespace)
	}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.CommandContext(ctx, c.kubectlBin, cmdArgs...)
	if len(stdin) > 0 {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, normalizeKubectlErr(output, err)
	}
	return output, nil
}

func (c *RealClient) resourceConfigMapName(kind, name string) string {
	raw := strings.ToLower(strings.TrimSpace(c.resourcePrefix + "-" + labelValue(kind) + "-" + labelValue(name)))
	if len(raw) <= 63 {
		return raw
	}
	sum := sha1.Sum([]byte(raw))
	hash := hex.EncodeToString(sum[:])[:10]
	prefix := raw[:52]
	prefix = strings.TrimRight(prefix, "-")
	return prefix + "-" + hash
}

func decodeResourcePayload(data map[string]string) (map[string]any, error) {
	raw := strings.TrimSpace(data[resourcePayloadKey])
	if raw == "" {
		return nil, errors.New("resource payload is empty")
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func incrementVersion(value any) string {
	current := strings.TrimSpace(fmt.Sprint(value))
	number, err := strconv.Atoi(current)
	if err != nil || number <= 0 {
		return "1"
	}
	return strconv.Itoa(number + 1)
}

func labelValue(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return "default"
	}

	var builder strings.Builder
	lastDash := false
	for _, r := range normalized {
		valid := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if valid {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteByte('-')
			lastDash = true
		}
	}
	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return "default"
	}
	if len(result) <= 63 {
		return result
	}
	sum := sha1.Sum([]byte(result))
	hash := hex.EncodeToString(sum[:])[:10]
	return result[:52] + "-" + hash
}

func normalizeKubectlErr(output []byte, err error) error {
	message := strings.TrimSpace(string(output))
	if strings.Contains(message, "NotFound") || strings.Contains(message, "not found") {
		return ErrNotFound
	}
	if message == "" {
		return err
	}
	return fmt.Errorf("%w: %s", err, message)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
