package chzzk

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ─── chzzk_generate_auth_code ─────────────────────────────────────────────────

type GenerateAuthCodeInput struct {
	Language string `json:"language" jsonschema:"코드 생성 언어. go 또는 typescript"`
}

func handleGenerateAuthCode(_ context.Context, _ *mcp.CallToolRequest, input GenerateAuthCodeInput) (*mcp.CallToolResult, any, error) {
	switch strings.ToLower(input.Language) {
	case "go":
		return textResult(authCodeGo()), nil, nil
	case "typescript", "ts":
		return textResult(authCodeTypeScript()), nil, nil
	default:
		return errorResult(fmt.Sprintf("지원하지 않는 언어: %q. 지원 언어: go, typescript", input.Language)), nil, nil
	}
}

// ─── chzzk_generate_api_client ────────────────────────────────────────────────

type GenerateAPIClientInput struct {
	Language  string   `json:"language" jsonschema:"코드 생성 언어. go 또는 typescript"`
	Endpoints []string `json:"endpoints" jsonschema:"클라이언트를 생성할 엔드포인트 목록. 'METHOD /path' 형식. 예: ['GET /open/v1/lives', 'POST /open/v1/chats/send']"`
}

func handleGenerateAPIClient(_ context.Context, _ *mcp.CallToolRequest, input GenerateAPIClientInput) (*mcp.CallToolResult, any, error) {
	if len(input.Endpoints) == 0 {
		return errorResult("endpoints는 최소 한 개 이상이어야 합니다"), nil, nil
	}

	var eps []Endpoint
	var notFound []string
	for _, key := range input.Endpoints {
		ep, ok := FindEndpoint(strings.TrimSpace(key))
		if !ok {
			notFound = append(notFound, key)
		} else {
			eps = append(eps, ep)
		}
	}
	if len(notFound) > 0 {
		return errorResult(fmt.Sprintf("찾을 수 없는 엔드포인트: %s\nchzzk_list_apis 도구로 올바른 엔드포인트를 확인하세요.", strings.Join(notFound, ", "))), nil, nil
	}

	switch strings.ToLower(input.Language) {
	case "go":
		return textResult(apiClientGo(eps)), nil, nil
	case "typescript", "ts":
		return textResult(apiClientTypeScript(eps)), nil, nil
	default:
		return errorResult(fmt.Sprintf("지원하지 않는 언어: %q. 지원 언어: go, typescript", input.Language)), nil, nil
	}
}

// ─── chzzk_scaffold_project ───────────────────────────────────────────────────

type ScaffoldProjectInput struct {
	Language    string   `json:"language" jsonschema:"프로젝트 언어. go 또는 typescript"`
	ProjectName string   `json:"project_name" jsonschema:"프로젝트 이름 (예: my-chzzk-bot)"`
	Features    []string `json:"features" jsonschema:"포함할 기능 목록. auth, live, chat, channel, session 중 선택. 예: ['auth', 'live', 'chat']"`
}

func handleScaffoldProject(_ context.Context, _ *mcp.CallToolRequest, input ScaffoldProjectInput) (*mcp.CallToolResult, any, error) {
	if input.ProjectName == "" {
		input.ProjectName = "chzzk-app"
	}
	if len(input.Features) == 0 {
		input.Features = []string{"auth", "live", "chat"}
	}

	featureSet := make(map[string]bool)
	for _, f := range input.Features {
		featureSet[strings.ToLower(f)] = true
	}

	switch strings.ToLower(input.Language) {
	case "go":
		return textResult(scaffoldGo(input.ProjectName, featureSet)), nil, nil
	case "typescript", "ts":
		return textResult(scaffoldTypeScript(input.ProjectName, featureSet)), nil, nil
	default:
		return errorResult(fmt.Sprintf("지원하지 않는 언어: %q. 지원 언어: go, typescript", input.Language)), nil, nil
	}
}

// ─── chzzk_generate_websocket_client ──────────────────────────────────────────

type GenerateWebSocketClientInput struct {
	Language string   `json:"language" jsonschema:"코드 생성 언어. go 또는 typescript"`
	Events   []string `json:"events" jsonschema:"구독할 이벤트 목록. chat, donation 중 선택. 예: ['chat', 'donation']. 미지정 시 chat, donation 모두 포함"`
}

func handleGenerateWebSocketClient(_ context.Context, _ *mcp.CallToolRequest, input GenerateWebSocketClientInput) (*mcp.CallToolResult, any, error) {
	if len(input.Events) == 0 {
		input.Events = supportedWSEvents
	}

	eventSet := make(map[string]bool)
	for _, e := range input.Events {
		e = strings.ToLower(e)
		valid := false
		for _, s := range supportedWSEvents {
			if s == e {
				valid = true
				break
			}
		}
		if !valid {
			return errorResult(fmt.Sprintf("지원하지 않는 이벤트: %q. 지원 이벤트: %s", e, strings.Join(supportedWSEvents, ", "))), nil, nil
		}
		eventSet[e] = true
	}

	switch strings.ToLower(input.Language) {
	case "go":
		return textResult(wsClientGo(eventSet)), nil, nil
	case "typescript", "ts":
		return textResult(wsClientTypeScript(eventSet)), nil, nil
	default:
		return errorResult(fmt.Sprintf("지원하지 않는 언어: %q. 지원 언어: go, typescript", input.Language)), nil, nil
	}
}

// RegisterCodegenTools adds code generation tools to the MCP server.
func RegisterCodegenTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name: "chzzk_generate_auth_code",
		Description: "치지직 OAuth2 인증 플로우 완성 코드를 생성합니다. " +
			"인가 코드 요청 URL 생성, 토큰 발급/갱신/폐기 함수, 콜백 서버(Go)를 포함합니다. " +
			"환경 변수: CHZZK_CLIENT_ID, CHZZK_CLIENT_SECRET, CHZZK_REDIRECT_URI. " +
			"지원 언어: go, typescript",
	}, handleGenerateAuthCode)

	mcp.AddTool(s, &mcp.Tool{
		Name: "chzzk_generate_api_client",
		Description: "지정한 치지직 API 엔드포인트에 대한 타입 안전 HTTP 클라이언트 코드를 생성합니다. " +
			"endpoints는 'METHOD /path' 형식으로 전달하세요 (예: 'GET /open/v1/lives'). " +
			"사용 가능한 엔드포인트는 chzzk_list_apis 또는 chzzk_get_api_spec으로 확인하세요. " +
			"지원 언어: go, typescript",
	}, handleGenerateAPIClient)

	mcp.AddTool(s, &mcp.Tool{
		Name: "chzzk_scaffold_project",
		Description: "치지직 API를 연동하는 서비스의 프로젝트 보일러플레이트를 생성합니다. " +
			"선택한 features(auth, live, chat, channel, session)에 맞는 디렉토리 구조, " +
			"기본 클라이언트, 설정 파일, 환경 변수 예시를 포함합니다. " +
			"지원 언어: go, typescript",
	}, handleScaffoldProject)

	mcp.AddTool(s, &mcp.Tool{
		Name: "chzzk_generate_websocket_client",
		Description: "치지직 WebSocket 실시간 이벤트 클라이언트 코드를 생성합니다. " +
			"Session API로 WebSocket URL을 발급받고, 지정한 이벤트(chat, donation)를 구독하는 " +
			"완성된 클라이언트 코드를 반환합니다. " +
			"인증: Access Token (CHZZK_ACCESS_TOKEN 환경 변수). " +
			"지원 언어: go, typescript",
	}, handleGenerateWebSocketClient)
}
