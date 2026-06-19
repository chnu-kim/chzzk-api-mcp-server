package chzzk

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleListAPIs_AllEndpoints(t *testing.T) {
	result, _, err := handleListAPIs(context.Background(), &mcp.CallToolRequest{}, ListAPIsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected non-empty content")
	}

	text := result.Content[0].(*mcp.TextContent).Text
	var out ListAPIsOutput
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if out.Total != len(AllEndpoints) {
		t.Errorf("total: got %d, want %d", out.Total, len(AllEndpoints))
	}

	for _, cat := range Categories {
		if _, ok := out.Endpoints[string(cat)]; !ok {
			t.Errorf("category %q missing from output", cat)
		}
	}
}

func TestHandleListAPIs_CategoryFilter(t *testing.T) {
	result, _, err := handleListAPIs(context.Background(), &mcp.CallToolRequest{}, ListAPIsInput{Category: "live"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	var out ListAPIsOutput
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if _, ok := out.Endpoints["live"]; !ok {
		t.Error("live category missing")
	}
	if len(out.Endpoints) != 1 {
		t.Errorf("expected only live category, got %d categories", len(out.Endpoints))
	}
}

func TestHandleListAPIs_CategoryCaseInsensitive(t *testing.T) {
	result, _, err := handleListAPIs(context.Background(), &mcp.CallToolRequest{}, ListAPIsInput{Category: "LIVE"})
	if err != nil {
		t.Fatal("handler should not return error")
	}
	if result.IsError {
		t.Errorf("expected success for uppercase category 'LIVE': %s", result.Content[0].(*mcp.TextContent).Text)
	}
}

func TestHandleListAPIs_UnknownCategory(t *testing.T) {
	result, _, err := handleListAPIs(context.Background(), &mcp.CallToolRequest{}, ListAPIsInput{Category: "unknown"})
	if err != nil {
		t.Fatal("handler should not return error")
	}
	if !result.IsError {
		t.Error("expected IsError=true for unknown category")
	}
}

func TestHandleGetAPISpec_KnownEndpoint(t *testing.T) {
	result, _, err := handleGetAPISpec(context.Background(), &mcp.CallToolRequest{}, GetAPISpecInput{Endpoint: "GET /open/v1/lives"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %s", result.Content[0].(*mcp.TextContent).Text)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	var ep Endpoint
	if err := json.Unmarshal([]byte(text), &ep); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if ep.Method != "GET" || ep.Path != "/open/v1/lives" {
		t.Errorf("wrong endpoint: %s %s", ep.Method, ep.Path)
	}
	if ep.Category != CategoryLive {
		t.Errorf("wrong category: %s", ep.Category)
	}
	if len(ep.Response) == 0 {
		t.Error("expected non-empty response fields")
	}
}

func TestHandleGetAPISpec_CaseInsensitive(t *testing.T) {
	result, _, err := handleGetAPISpec(context.Background(), &mcp.CallToolRequest{}, GetAPISpecInput{Endpoint: "get /open/v1/lives"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("case-insensitive lookup failed: %s", result.Content[0].(*mcp.TextContent).Text)
	}
}

func TestHandleGetAPISpec_UnknownEndpoint(t *testing.T) {
	result, _, err := handleGetAPISpec(context.Background(), &mcp.CallToolRequest{}, GetAPISpecInput{Endpoint: "GET /no/such/path"})
	if err != nil {
		t.Fatal("handler should not return error")
	}
	if !result.IsError {
		t.Error("expected IsError=true for unknown endpoint")
	}
}

func TestHandleGetAPISpec_InvalidInput(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"whitespace only", "   "},
		{"no HTTP method", "/open/v1/lives"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, _, err := handleGetAPISpec(context.Background(), &mcp.CallToolRequest{}, GetAPISpecInput{Endpoint: tc.input})
			if err != nil {
				t.Fatal("handler should not return error")
			}
			if !result.IsError {
				t.Errorf("expected IsError=true for input %q", tc.input)
			}
			// suggestion section은 "\n\n유사한 엔드포인트:" 로 시작하므로
			// "\n\n" 이 없으면 제안 없음
			if strings.Contains(result.Content[0].(*mcp.TextContent).Text, "\n\n") {
				t.Errorf("input %q should produce no suggestion section", tc.input)
			}
		})
	}
}

func TestHandleGetAPISpec_SuggestionOnPartialMatch(t *testing.T) {
	result, _, err := handleGetAPISpec(context.Background(), &mcp.CallToolRequest{}, GetAPISpecInput{Endpoint: "GET /lives"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for incomplete path")
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "/open/v1/lives") {
		t.Error("expected suggestion to contain /open/v1/lives")
	}
}

func TestRegisterReferenceTools(t *testing.T) {
	s := NewMCPServer()
	if s == nil {
		t.Fatal("NewMCPServer returned nil")
	}
}
