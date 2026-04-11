package middleware

import "testing"

func TestAnonymousPathMatching(t *testing.T) {
	tests := map[string]bool{
		"/healthz":             true,
		"/healthz/ready":       true,
		"/landing":             true,
		"/session/login":       true,
		"/system/info":         true,
		"/system/config":       true,
		"/system/init":         true,
		"/dashboard/info":      false,
		"/user/info":           false,
		"/v1/routes":           false,
		"/mcp-templates/index": true,
	}

	for path, expected := range tests {
		if actual := isAnonymousPath(path); actual != expected {
			t.Fatalf("path %s anonymous=%v, want %v", path, actual, expected)
		}
	}
}
