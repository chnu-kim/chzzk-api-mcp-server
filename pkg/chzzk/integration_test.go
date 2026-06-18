package chzzk

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func connectIntegration(t *testing.T) (*mcp.ClientSession, func()) {
	t.Helper()
	ctx := context.Background()

	server := NewMCPServer()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.1"}, nil)

	t1, t2 := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, t1, nil); err != nil {
		t.Fatal("server.Connect:", err)
	}
	cs, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatal("client.Connect:", err)
	}
	return cs, func() { cs.Close() }
}

func TestIntegration_ListTools_FiveToolsRegistered(t *testing.T) {
	cs, cleanup := connectIntegration(t)
	defer cleanup()

	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal("ListTools:", err)
	}

	want := []string{
		"chzzk_list_apis",
		"chzzk_get_api_spec",
		"chzzk_generate_auth_code",
		"chzzk_generate_api_client",
		"chzzk_scaffold_project",
	}
	got := make(map[string]bool, len(res.Tools))
	for _, tool := range res.Tools {
		got[tool.Name] = true
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("tool %q not found in ListTools result", name)
		}
	}
	if len(res.Tools) != len(want) {
		t.Errorf("expected %d tools, got %d", len(want), len(res.Tools))
	}
}

func TestIntegration_ListTools_InputSchemasNotNil(t *testing.T) {
	cs, cleanup := connectIntegration(t)
	defer cleanup()

	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal("ListTools:", err)
	}
	for _, tool := range res.Tools {
		if tool.InputSchema == nil {
			t.Errorf("tool %q has nil InputSchema", tool.Name)
		}
	}
}

func TestIntegration_CallTool_ListAPIs_ValidJSON(t *testing.T) {
	cs, cleanup := connectIntegration(t)
	defer cleanup()

	result, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "chzzk_list_apis",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatal("CallTool:", err)
	}
	if result.IsError {
		t.Fatalf("unexpected IsError=true: %v", result.Content)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	var parsed any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Errorf("response is not valid JSON: %v\n%s", err, text)
	}
}

func TestIntegration_CallTool_GetAPISpec(t *testing.T) {
	cs, cleanup := connectIntegration(t)
	defer cleanup()

	result, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "chzzk_get_api_spec",
		Arguments: map[string]any{"endpoint": "GET /open/v1/lives"},
	})
	if err != nil {
		t.Fatal("CallTool:", err)
	}
	if result.IsError {
		text := result.Content[0].(*mcp.TextContent).Text
		t.Fatalf("unexpected IsError=true: %s", text)
	}
}

func TestIntegration_CallTool_GenerateAuthCode_Go(t *testing.T) {
	cs, cleanup := connectIntegration(t)
	defer cleanup()

	result, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "chzzk_generate_auth_code",
		Arguments: map[string]any{"language": "go"},
	})
	if err != nil {
		t.Fatal("CallTool:", err)
	}
	if result.IsError {
		text := result.Content[0].(*mcp.TextContent).Text
		t.Fatalf("unexpected IsError=true: %s", text)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "package auth") {
		t.Errorf("expected 'package auth' in response, got:\n%s", text)
	}
}
