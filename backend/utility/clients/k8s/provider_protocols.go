package k8s

import "strings"

const (
	AIProviderProtocolAuto               = "auto"
	AIProviderProtocolOpenAIV1          = "openai/v1"
	AIProviderProtocolAnthropicMessages = "anthropic/v1/messages"
	AIProviderProtocolOriginal          = "original"

	AIProviderDefaultOpenAIEndpoint    = "/v1/chat/completions"
	AIProviderDefaultAnthropicEndpoint = "/v1/messages"
	AIProviderDefaultResponsesEndpoint = "/v1/responses"
)

func NormalizeProviderProtocolCanonical(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, "_", "/")
	switch normalized {
	case "", "<nil>", AIProviderProtocolAuto:
		return AIProviderProtocolAuto
	case "openai", "openai/v1", "openai/v1/chatcompletions", "openai/v1/chat/completions":
		return AIProviderProtocolOpenAIV1
	case "anthropic", "claude", "anthropic/v1/messages", "/v1/messages":
		return AIProviderProtocolAnthropicMessages
	case "original":
		return AIProviderProtocolOriginal
	default:
		return strings.TrimSpace(value)
	}
}

func NormalizeBindingProtocolCanonical(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "", "<nil>", AIProviderProtocolAuto, "openai", "openai/v1", "openai/v1/chatcompletions", "openai/v1/chat/completions", "responses", "openai/v1/responses":
		return AIProviderProtocolOpenAIV1
	case "anthropic", "claude", "anthropic/v1/messages", "/v1/messages":
		return AIProviderProtocolAnthropicMessages
	case "original":
		return AIProviderProtocolOriginal
	default:
		return strings.TrimSpace(value)
	}
}

func IsSupportedProviderProtocol(value string) bool {
	switch NormalizeProviderProtocolCanonical(value) {
	case AIProviderProtocolAuto, AIProviderProtocolOpenAIV1, AIProviderProtocolAnthropicMessages, AIProviderProtocolOriginal:
		return true
	default:
		return false
	}
}

func IsSupportedBindingProtocol(value string) bool {
	switch NormalizeBindingProtocolCanonical(value) {
	case AIProviderProtocolOpenAIV1, AIProviderProtocolAnthropicMessages, AIProviderProtocolOriginal:
		return true
	default:
		return false
	}
}

func RecommendedEndpointForProtocol(protocol string) string {
	switch NormalizeBindingProtocolCanonical(protocol) {
	case AIProviderProtocolOpenAIV1:
		return AIProviderDefaultOpenAIEndpoint
	case AIProviderProtocolAnthropicMessages:
		return AIProviderDefaultAnthropicEndpoint
	default:
		return ""
	}
}
