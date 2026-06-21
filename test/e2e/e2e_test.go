//go:build e2e

// E2E 테스트: 실제 바이너리를 subprocess로 실행해 배포 경로 전체를 검증한다.
//
// 실행: go test -tags e2e ./test/e2e/... -v
package e2e

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/chanuuuu/chzzk-api-mcp-server/pkg/chzzk"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var serverBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "chzzk-e2e-*")
	if err != nil {
		panic("tmpdir: " + err.Error())
	}
	serverBin = filepath.Join(dir, "chzzk-mcp-server")
	cmd := exec.Command("go", "build", "-o", serverBin,
		"github.com/chanuuuu/chzzk-api-mcp-server/cmd/chzzk-mcp-server")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("go build: " + err.Error())
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func connect(t *testing.T) *mcp.ClientSession {
	t.Helper()
	cmd := exec.Command(serverBin, "stdio")
	client := mcp.NewClient(&mcp.Implementation{Name: "e2e-client", Version: "v0.0.1"}, nil)
	session, err := client.Connect(context.Background(), &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

// TestInitialize — Initialize 핸드셰이크 성공, ServerInfo.Name 검증.
func TestInitialize(t *testing.T) {
	session := connect(t)
	res := session.InitializeResult()
	if res.ServerInfo.Name != chzzk.ServerName {
		t.Errorf("ServerInfo.Name = %q, want %q", res.ServerInfo.Name, chzzk.ServerName)
	}
}

// TestListTools — 6개 도구가 모두 등록되어 있는지 검증.
func TestListTools(t *testing.T) {
	session := connect(t)
	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	const want = 6
	if len(res.Tools) != want {
		t.Errorf("tool count = %d, want %d", len(res.Tools), want)
	}
}

// TestCallTool_ListAPIs — chzzk_list_apis 호출이 정상 응답을 반환하는지 검증.
func TestCallTool_ListAPIs(t *testing.T) {
	session := connect(t)
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "chzzk_list_apis",
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %v", res.Content)
	}
	if len(res.Content) == 0 {
		t.Fatal("empty content")
	}
}

// TestStdinEOF_ExitsClean — stdin EOF 시 서버가 exit code 0으로 종료하는지 검증.
func TestStdinEOF_ExitsClean(t *testing.T) {
	cmd := exec.Command(serverBin, "stdio")
	cmd.Stdout = io.Discard
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cmd.Process.Kill() }) //nolint

	stdin.Close()

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("exit error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("process did not exit within 5s after stdin EOF")
	}
}
