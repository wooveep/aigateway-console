package portal

import "testing"

func TestNormalizeBindingCanonicalProtocolAndEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		binding  ModelAssetBinding
		protocol string
		endpoint string
	}{
		{
			name: "responses protocol collapses into openai with responses endpoint",
			binding: ModelAssetBinding{
				ModelID:      "gpt-4.1",
				ProviderName: "openai",
				TargetModel:  "gpt-4.1",
				Protocol:     "responses",
			},
			protocol: "openai/v1",
			endpoint: "/v1/responses",
		},
		{
			name: "anthropic alias normalizes to canonical messages protocol",
			binding: ModelAssetBinding{
				ModelID:      "claude-sonnet",
				ProviderName: "claude",
				TargetModel:  "claude-sonnet",
				Protocol:     "anthropic",
			},
			protocol: "anthropic/v1/messages",
			endpoint: "/v1/messages",
		},
		{
			name: "original protocol keeps explicit endpoint empty",
			binding: ModelAssetBinding{
				ModelID:      "ollama-llama3",
				ProviderName: "ollama",
				TargetModel:  "llama3",
				Protocol:     "original",
			},
			protocol: "original",
			endpoint: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			normalized, err := normalizeBinding(tt.binding)
			if err != nil {
				t.Fatalf("normalizeBinding() error = %v", err)
			}
			if normalized.Protocol != tt.protocol {
				t.Fatalf("protocol = %q, want %q", normalized.Protocol, tt.protocol)
			}
			if normalized.Endpoint != tt.endpoint {
				t.Fatalf("endpoint = %q, want %q", normalized.Endpoint, tt.endpoint)
			}
		})
	}
}

