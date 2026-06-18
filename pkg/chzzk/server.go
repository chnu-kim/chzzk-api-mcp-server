package chzzk

import "github.com/modelcontextprotocol/go-sdk/mcp"

const (
	ServerName    = "chzzk-api-mcp-server"
	ServerVersion = "0.1.0"
)

// NewMCPServer creates and configures the Chzzk MCP server with all tools registered.
func NewMCPServer() *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    ServerName,
		Version: ServerVersion,
	}, &mcp.ServerOptions{
		Instructions: "치지직(Chzzk) Open API 연동 서비스 개발을 돕는 API 레퍼런스 조회 및 코드 생성 MCP 서버입니다. " +
			"API 레퍼런스 조회(chzzk_list_apis, chzzk_get_api_spec)와 " +
			"코드 생성(chzzk_generate_auth_code, chzzk_generate_api_client, chzzk_scaffold_project)을 지원합니다.",
	})

	RegisterReferenceTools(s)
	RegisterCodegenTools(s)

	return s
}
