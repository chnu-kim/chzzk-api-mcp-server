package chzzkmcp

import (
	"context"

	"github.com/chanuuuu/chzzk-api-mcp-server/pkg/chzzk"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RunStdioServer starts the MCP server over stdin/stdout and blocks until the
// client disconnects or the context is cancelled.
func RunStdioServer(ctx context.Context) error {
	s := chzzk.NewMCPServer()
	return s.Run(ctx, &mcp.StdioTransport{})
}
