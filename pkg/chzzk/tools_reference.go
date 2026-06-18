package chzzk

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ─── chzzk_list_apis ─────────────────────────────────────────────────────────

type ListAPIsInput struct {
	Category string `json:"category,omitempty" jsonschema:"치지직 API 카테고리 필터 (auth|user|channel|category|live|chat|session|drops|restriction). 생략하면 전체 반환."`
}

type APISummary struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Category    string `json:"category"`
	Description string `json:"description"`
	AuthType    string `json:"auth_type"`
	Scope       string `json:"scope,omitempty"`
}

type ListAPIsOutput struct {
	Total     int                    `json:"total"`
	Endpoints map[string][]APISummary `json:"endpoints"`
}

func handleListAPIs(_ context.Context, _ *mcp.CallToolRequest, input ListAPIsInput) (*mcp.CallToolResult, any, error) {
	var filtered []Endpoint
	if input.Category == "" {
		filtered = AllEndpoints
	} else {
		cat := Category(strings.ToLower(input.Category))
		for _, e := range AllEndpoints {
			if e.Category == cat {
				filtered = append(filtered, e)
			}
		}
		if len(filtered) == 0 {
			return errorResult(fmt.Sprintf("알 수 없는 카테고리: %q. 사용 가능한 카테고리: auth, user, channel, category, live, chat, session, drops, restriction", input.Category)), nil, nil
		}
	}

	grouped := make(map[string][]APISummary)
	for _, e := range filtered {
		cat := string(e.Category)
		grouped[cat] = append(grouped[cat], APISummary{
			Method:      e.Method,
			Path:        e.Path,
			Category:    cat,
			Description: e.Description,
			AuthType:    string(e.AuthType),
			Scope:       e.Scope,
		})
	}

	out := ListAPIsOutput{
		Total:     len(filtered),
		Endpoints: grouped,
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return textResult(string(b)), nil, nil
}

// ─── chzzk_get_api_spec ───────────────────────────────────────────────────────

type GetAPISpecInput struct {
	Endpoint string `json:"endpoint" jsonschema:"조회할 엔드포인트. 'METHOD /path' 형식. 예: 'GET /open/v1/lives', 'POST /open/v1/chats/send'"`
}

func handleGetAPISpec(_ context.Context, _ *mcp.CallToolRequest, input GetAPISpecInput) (*mcp.CallToolResult, any, error) {
	key := strings.TrimSpace(input.Endpoint)

	ep, ok := FindEndpoint(key)
	if !ok {
		upper := strings.ToUpper(key)
		for _, e := range AllEndpoints {
			if strings.ToUpper(e.Key()) == upper {
				ep = e
				ok = true
				break
			}
		}
	}

	if !ok {
		var suggestions []string
		parts := strings.SplitN(key, " ", 2)
		searchPath := ""
		if len(parts) == 2 {
			searchPath = strings.ToLower(parts[1])
		}
		for _, e := range AllEndpoints {
			if searchPath != "" && strings.Contains(strings.ToLower(e.Path), searchPath) {
				suggestions = append(suggestions, e.Key())
			}
		}
		msg := fmt.Sprintf("엔드포인트를 찾을 수 없습니다: %q", key)
		if len(suggestions) > 0 {
			msg += "\n\n유사한 엔드포인트:\n  " + strings.Join(suggestions, "\n  ")
		}
		return errorResult(msg), nil, nil
	}

	b, err := json.MarshalIndent(ep, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return textResult(string(b)), nil, nil
}

// RegisterReferenceTools adds reference tools to the MCP server.
func RegisterReferenceTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name: "chzzk_list_apis",
		Description: "치지직(Chzzk) Open API 전체 엔드포인트 목록을 카테고리별로 반환합니다. " +
			"category 파라미터로 필터링할 수 있습니다 (auth, user, channel, category, live, chat, session, drops, restriction). " +
			"각 엔드포인트의 HTTP 메서드, 경로, 설명, 인증 타입을 포함합니다.",
	}, handleListAPIs)

	mcp.AddTool(s, &mcp.Tool{
		Name: "chzzk_get_api_spec",
		Description: "특정 치지직 API 엔드포인트의 상세 스펙을 반환합니다. " +
			"요청 파라미터(query/path/body), 응답 필드, 인증 방식, 스코프를 포함합니다. " +
			"'METHOD /path' 형식으로 전달하세요. 예: 'GET /open/v1/lives', 'POST /auth/v1/token'",
	}, handleGetAPISpec)
}
