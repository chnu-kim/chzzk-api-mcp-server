package chzzk

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ─── chzzk_generate_auth_code ─────────────────────────────────────────────────

func TestHandleGenerateAuthCode_Go(t *testing.T) {
	result, _, err := handleGenerateAuthCode(context.Background(), &mcp.CallToolRequest{}, GenerateAuthCodeInput{Language: "go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %s", result.Content[0].(*mcp.TextContent).Text)
	}

	code := result.Content[0].(*mcp.TextContent).Text
	for _, want := range []string{
		"package auth",
		"TokenResponse",
		"ExchangeCode",
		"RefreshToken",
		"RevokeToken",
		"AuthorizationURL",
		"CHZZK_CLIENT_ID",
		"CHZZK_CLIENT_SECRET",
		// Content-Type: JSON 전송
		"application/json",
	} {
		if !strings.Contains(code, want) {
			t.Errorf("generated Go auth code missing %q", want)
		}
	}
}

func TestHandleGenerateAuthCode_TypeScript(t *testing.T) {
	for _, lang := range []string{"typescript", "ts"} {
		result, _, err := handleGenerateAuthCode(context.Background(), &mcp.CallToolRequest{}, GenerateAuthCodeInput{Language: lang})
		if err != nil {
			t.Fatalf("lang=%s: unexpected error: %v", lang, err)
		}
		if result.IsError {
			t.Fatalf("lang=%s: unexpected error result", lang)
		}

		code := result.Content[0].(*mcp.TextContent).Text
		for _, want := range []string{
			"ChzzkAuthConfig",
			"exchangeCode",
			"refreshAccessToken",
			"revokeToken",
			"buildAuthorizationURL",
			"CHZZK_CLIENT_ID",
			// Content-Type: JSON 전송
			"application/json",
			// expiresIn: API가 String 반환 가능 → Number 정규화
			"Number(",
			// scope: 공백 구분 문자열 → 배열
			"split(",
			// revokeToken이 HTTP status뿐 아니라 API envelope code도 검사
			"ApiResponse<unknown>",
		} {
			if !strings.Contains(code, want) {
				t.Errorf("lang=%s: generated TS auth code missing %q", lang, want)
			}
		}
	}
}

func TestHandleGenerateAuthCode_UnsupportedLanguage(t *testing.T) {
	result, _, err := handleGenerateAuthCode(context.Background(), &mcp.CallToolRequest{}, GenerateAuthCodeInput{Language: "ruby"})
	if err != nil {
		t.Fatal("handler should not return error")
	}
	if !result.IsError {
		t.Error("expected IsError=true for unsupported language")
	}
}

// ─── chzzk_generate_api_client ────────────────────────────────────────────────

func TestHandleGenerateAPIClient_Go(t *testing.T) {
	result, _, err := handleGenerateAPIClient(context.Background(), &mcp.CallToolRequest{}, GenerateAPIClientInput{
		Language:  "go",
		Endpoints: []string{"GET /open/v1/lives", "POST /open/v1/chats/send"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("error result: %s", result.Content[0].(*mcp.TextContent).Text)
	}

	code := result.Content[0].(*mcp.TextContent).Text
	for _, want := range []string{
		"package chzzk",
		"Client",
		"NewClient",
		"baseURL",
		"Lives",  // from GET /open/v1/lives method name
		"Chats",  // from POST /open/v1/chats/send
	} {
		if !strings.Contains(code, want) {
			t.Errorf("Go client missing %q", want)
		}
	}
}

func TestHandleGenerateAPIClient_Go_BodyParams(t *testing.T) {
	result, _, err := handleGenerateAPIClient(context.Background(), &mcp.CallToolRequest{}, GenerateAPIClientInput{
		Language:  "go",
		Endpoints: []string{"POST /open/v1/chats/send"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("error result: %s", result.Content[0].(*mcp.TextContent).Text)
	}

	code := result.Content[0].(*mcp.TextContent).Text
	for _, want := range []string{
		"json.Marshal",   // body 직렬화
		`"message"`,      // body 필드 json 태그
		"string(bodyBytes)", // Marshal 결과를 do()에 전달
	} {
		if !strings.Contains(code, want) {
			t.Errorf("Go client body serialization missing %q\ngenerated code:\n%s", want, code)
		}
	}
}

func TestHandleGenerateAPIClient_TypeScript(t *testing.T) {
	for _, lang := range []string{"typescript", "ts"} {
		result, _, err := handleGenerateAPIClient(context.Background(), &mcp.CallToolRequest{}, GenerateAPIClientInput{
			Language:  lang,
			Endpoints: []string{"GET /open/v1/channels"},
		})
		if err != nil {
			t.Fatalf("lang=%s: unexpected error: %v", lang, err)
		}
		if result.IsError {
			t.Fatalf("lang=%s: error result: %s", lang, result.Content[0].(*mcp.TextContent).Text)
		}

		code := result.Content[0].(*mcp.TextContent).Text
		for _, want := range []string{
			"ChzzkClient",
			"BASE_URL",
			"request",
		} {
			if !strings.Contains(code, want) {
				t.Errorf("lang=%s: TS client missing %q", lang, want)
			}
		}
	}
}

func TestHandleGenerateAPIClient_EmptyEndpoints(t *testing.T) {
	result, _, err := handleGenerateAPIClient(context.Background(), &mcp.CallToolRequest{}, GenerateAPIClientInput{
		Language:  "go",
		Endpoints: []string{},
	})
	if err != nil {
		t.Fatal("handler should not error")
	}
	if !result.IsError {
		t.Error("expected IsError=true for empty endpoints")
	}
}

func TestHandleGenerateAPIClient_UnknownEndpoint(t *testing.T) {
	result, _, err := handleGenerateAPIClient(context.Background(), &mcp.CallToolRequest{}, GenerateAPIClientInput{
		Language:  "go",
		Endpoints: []string{"GET /no/such/endpoint"},
	})
	if err != nil {
		t.Fatal("handler should not error")
	}
	if !result.IsError {
		t.Error("expected IsError=true for unknown endpoint")
	}
}

func TestHandleGenerateAPIClient_UnsupportedLanguage(t *testing.T) {
	result, _, err := handleGenerateAPIClient(context.Background(), &mcp.CallToolRequest{}, GenerateAPIClientInput{
		Language:  "python",
		Endpoints: []string{"GET /open/v1/lives"},
	})
	if err != nil {
		t.Fatal("handler should not error")
	}
	if !result.IsError {
		t.Error("expected IsError=true for unsupported language")
	}
}

// ─── chzzk_scaffold_project ───────────────────────────────────────────────────

func TestHandleScaffoldProject_Go(t *testing.T) {
	result, _, err := handleScaffoldProject(context.Background(), &mcp.CallToolRequest{}, ScaffoldProjectInput{
		Language:    "go",
		ProjectName: "my-bot",
		Features:    []string{"auth", "live", "chat", "channel", "session"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("error result: %s", result.Content[0].(*mcp.TextContent).Text)
	}

	out := result.Content[0].(*mcp.TextContent).Text
	for _, want := range []string{
		"my-bot",
		"go.mod",
		"main.go",
		"CHZZK_CLIENT_ID",
		"auth/",
		"live/",
		"chat/",
		"channel/",
		"session/",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Go scaffold missing %q", want)
		}
	}
}

func TestHandleScaffoldProject_TypeScript(t *testing.T) {
	result, _, err := handleScaffoldProject(context.Background(), &mcp.CallToolRequest{}, ScaffoldProjectInput{
		Language:    "typescript",
		ProjectName: "chzzk-ts-app",
		Features:    []string{"auth", "live"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("error result: %s", result.Content[0].(*mcp.TextContent).Text)
	}

	out := result.Content[0].(*mcp.TextContent).Text
	for _, want := range []string{
		"chzzk-ts-app",
		"package.json",
		"tsconfig.json",
		"client.ts",
		"CHZZK_CLIENT_ID",
		"ChzzkClient",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("TS scaffold missing %q", want)
		}
	}
}

func TestHandleScaffoldProject_DefaultValues(t *testing.T) {
	result, _, err := handleScaffoldProject(context.Background(), &mcp.CallToolRequest{}, ScaffoldProjectInput{
		Language: "go",
		// ProjectName and Features omitted → defaults applied
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("error: %s", result.Content[0].(*mcp.TextContent).Text)
	}
	out := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(out, "chzzk-app") {
		t.Error("expected default project name 'chzzk-app'")
	}
}

func TestHandleScaffoldProject_UnsupportedLanguage(t *testing.T) {
	result, _, err := handleScaffoldProject(context.Background(), &mcp.CallToolRequest{}, ScaffoldProjectInput{
		Language: "rust",
	})
	if err != nil {
		t.Fatal("handler should not error")
	}
	if !result.IsError {
		t.Error("expected IsError=true for unsupported language")
	}
}

// ─── api_reference helpers ────────────────────────────────────────────────────

func TestFindEndpoint(t *testing.T) {
	ep, ok := FindEndpoint("GET /open/v1/lives")
	if !ok {
		t.Fatal("expected to find GET /open/v1/lives")
	}
	if ep.Category != CategoryLive {
		t.Errorf("wrong category: %s", ep.Category)
	}
}

func TestEndpointsByCategory(t *testing.T) {
	chatEps := EndpointsByCategory(CategoryChat)
	if len(chatEps) == 0 {
		t.Error("expected chat endpoints")
	}
	for _, ep := range chatEps {
		if ep.Category != CategoryChat {
			t.Errorf("wrong category for %s %s", ep.Method, ep.Path)
		}
	}
}

func TestAllEndpointsHaveRequiredFields(t *testing.T) {
	for _, ep := range AllEndpoints {
		if ep.Method == "" {
			t.Errorf("endpoint missing method: %+v", ep)
		}
		if ep.Path == "" {
			t.Errorf("endpoint missing path: %+v", ep)
		}
		if ep.Category == "" {
			t.Errorf("endpoint missing category: %s %s", ep.Method, ep.Path)
		}
		if ep.Description == "" {
			t.Errorf("endpoint missing description: %s %s", ep.Method, ep.Path)
		}
		if ep.AuthType == "" {
			t.Errorf("endpoint missing auth_type: %s %s", ep.Method, ep.Path)
		}
	}
}
