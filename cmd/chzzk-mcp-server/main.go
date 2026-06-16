package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/chanuuuu/chzzk-api-mcp-server/internal/chzzkmcp"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "chzzk-mcp-server",
	Short: "Chzzk API MCP 서버",
	Long: `치지직(Chzzk) Open API 연동 서비스 개발을 돕는 MCP 서버.
API 레퍼런스 조회 및 Go/TypeScript 코드 생성 도구를 제공합니다.`,
}

var stdioCmd = &cobra.Command{
	Use:   "stdio",
	Short: "stdin/stdout 트랜스포트로 MCP 서버 실행",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		return chzzkmcp.RunStdioServer(ctx)
	},
}

func init() {
	rootCmd.AddCommand(stdioCmd)
}

func main() {
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
