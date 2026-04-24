package gateway

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	k8sclient "github.com/wooveep/aigateway-console/backend/utility/clients/k8s"
)

var providerDocsAliases = map[string][]string{
	"openai":    {"openai"},
	"claude":    {"claude"},
	"anthropic": {"claude"},
	"qwen":      {"qwen"},
	"ollama":    {"ollama"},
	"kimi":      {"kimi"},
	"zhipuai":   {"glm"},
	"glm":       {"glm"},
}

var providerProtocolFacts = map[string]map[string]any{
	"openai": {
		"nativeProtocols":     []string{k8sclient.AIProviderProtocolOpenAIV1},
		"compatibleProtocols": []string{k8sclient.AIProviderProtocolOpenAIV1, k8sclient.AIProviderProtocolAnthropicMessages},
		"claudeFallback":      true,
		"recommendedEndpoints": map[string]string{
			k8sclient.AIProviderProtocolOpenAIV1:          k8sclient.AIProviderDefaultOpenAIEndpoint,
			k8sclient.AIProviderProtocolAnthropicMessages: k8sclient.AIProviderDefaultAnthropicEndpoint,
			k8sclient.AIProviderProtocolOriginal:          k8sclient.AIProviderDefaultOpenAIEndpoint,
		},
		"capabilities": []string{"responses", "embeddings", "images", "audio", "realtime", "files", "batches", "models", "videos"},
	},
	"claude": {
		"nativeProtocols":     []string{k8sclient.AIProviderProtocolAnthropicMessages},
		"compatibleProtocols": []string{k8sclient.AIProviderProtocolOpenAIV1, k8sclient.AIProviderProtocolAnthropicMessages},
		"claudeFallback":      true,
		"recommendedEndpoints": map[string]string{
			k8sclient.AIProviderProtocolOpenAIV1:          k8sclient.AIProviderDefaultOpenAIEndpoint,
			k8sclient.AIProviderProtocolAnthropicMessages: k8sclient.AIProviderDefaultAnthropicEndpoint,
			k8sclient.AIProviderProtocolOriginal:          k8sclient.AIProviderDefaultAnthropicEndpoint,
		},
		"capabilities": []string{"embeddings", "models"},
	},
	"qwen": {
		"nativeProtocols":     []string{k8sclient.AIProviderProtocolOpenAIV1},
		"compatibleProtocols": []string{k8sclient.AIProviderProtocolOpenAIV1, k8sclient.AIProviderProtocolAnthropicMessages},
		"claudeFallback":      true,
		"recommendedEndpoints": map[string]string{
			k8sclient.AIProviderProtocolOpenAIV1:          k8sclient.AIProviderDefaultOpenAIEndpoint,
			k8sclient.AIProviderProtocolAnthropicMessages: k8sclient.AIProviderDefaultAnthropicEndpoint,
		},
		"capabilities": []string{"embeddings", "images", "audio", "models"},
	},
	"zhipuai": {
		"nativeProtocols":     []string{k8sclient.AIProviderProtocolOpenAIV1},
		"compatibleProtocols": []string{k8sclient.AIProviderProtocolOpenAIV1, k8sclient.AIProviderProtocolAnthropicMessages},
		"claudeFallback":      true,
		"recommendedEndpoints": map[string]string{
			k8sclient.AIProviderProtocolOpenAIV1:          k8sclient.AIProviderDefaultOpenAIEndpoint,
			k8sclient.AIProviderProtocolAnthropicMessages: k8sclient.AIProviderDefaultAnthropicEndpoint,
		},
		"capabilities": []string{"embeddings", "responses", "models"},
	},
	"ollama": {
		"nativeProtocols":     []string{k8sclient.AIProviderProtocolOriginal},
		"compatibleProtocols": []string{k8sclient.AIProviderProtocolOpenAIV1, k8sclient.AIProviderProtocolAnthropicMessages, k8sclient.AIProviderProtocolOriginal},
		"claudeFallback":      true,
		"recommendedEndpoints": map[string]string{
			k8sclient.AIProviderProtocolOpenAIV1:          k8sclient.AIProviderDefaultOpenAIEndpoint,
			k8sclient.AIProviderProtocolAnthropicMessages: k8sclient.AIProviderDefaultAnthropicEndpoint,
			k8sclient.AIProviderProtocolOriginal:          "/api/chat",
		},
		"capabilities": []string{"embeddings", "models"},
	},
}

func (s *Service) GetProviderProtocolDirectory(ctx context.Context) map[string]any {
	_ = ctx

	providerDocsStatus := map[string]string{}
	providerProtocolMatrix := map[string]any{}
	for providerType, facts := range providerProtocolFacts {
		clone := clonePayload(facts)
		docsStatus := providerDocsStatusForProvider(providerType)
		clone["docsStatus"] = docsStatus
		providerDocsStatus[providerType] = docsStatus
		providerProtocolMatrix[providerType] = clone
	}

	return map[string]any{
		"protocolOptions": map[string]any{
			"provider": []map[string]string{
				{"label": "Auto", "value": k8sclient.AIProviderProtocolAuto},
				{"label": "OpenAI", "value": k8sclient.AIProviderProtocolOpenAIV1},
				{"label": "Claude", "value": k8sclient.AIProviderProtocolAnthropicMessages},
				{"label": "Original", "value": k8sclient.AIProviderProtocolOriginal},
			},
			"binding": []map[string]string{
				{"label": "OpenAI", "value": k8sclient.AIProviderProtocolOpenAIV1},
				{"label": "Claude", "value": k8sclient.AIProviderProtocolAnthropicMessages},
				{"label": "Original", "value": k8sclient.AIProviderProtocolOriginal},
			},
		},
		"recommendedEndpoints": map[string]string{
			k8sclient.AIProviderProtocolOpenAIV1:          k8sclient.AIProviderDefaultOpenAIEndpoint,
			k8sclient.AIProviderProtocolAnthropicMessages: k8sclient.AIProviderDefaultAnthropicEndpoint,
			"openai/v1/responses":                         k8sclient.AIProviderDefaultResponsesEndpoint,
		},
		"providerProtocolMatrix": providerProtocolMatrix,
		"providerDocsStatus":     providerDocsStatus,
	}
}

func providerDocsStatusForProvider(providerType string) string {
	root := providerDocsRoot()
	if root == "" {
		return "pending"
	}
	for _, dirName := range providerDocsAliases[strings.ToLower(strings.TrimSpace(providerType))] {
		if docsDirHasOfficialFiles(filepath.Join(root, dirName)) {
			return "confirmed"
		}
	}
	return "pending"
}

func providerDocsRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	current := wd
	for range 8 {
		candidate := filepath.Join(current, "TASK", "projects", "aigateway-console", "provider-api-docs")
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return ""
}

func docsDirHasOfficialFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(entry.Name()))
		if name == "" || name == "readme.md" {
			continue
		}
		if strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".pdf") || strings.HasSuffix(name, ".html") {
			return true
		}
	}
	return false
}
