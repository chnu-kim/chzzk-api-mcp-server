# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 개요

치지직(Chzzk) Open API 연동 서비스 개발을 돕는 MCP 서버. 두 기능이 동등하게 핵심이다:

- **API 레퍼런스 조회** — 36개 엔드포인트 데이터를 내장해 파라미터·응답 스펙을 즉시 반환
- **코드 생성** — OAuth2 인증 플로우, 타입 안전 HTTP 클라이언트, 프로젝트 스캐폴드 생성 (Go/TypeScript)

런타임에 Chzzk API를 직접 호출하지 않는다. 모든 데이터는 `api_reference.go`에 정적으로 내장되어 있다.

## 주요 명령어

```bash
# 빌드
go build ./...
go build -o chzzk-mcp-server ./cmd/chzzk-mcp-server/

# 테스트
go test ./...
go test ./pkg/chzzk/...               # 패키지 단위
go test ./pkg/chzzk/... -run TestName # 특정 테스트
go test ./pkg/chzzk/... -v            # 상세 출력

# 실행
./chzzk-mcp-server stdio

# 의존성 정리
go mod tidy
```

## 아키텍처

요청 흐름: `MCP 클라이언트 → cmd(Cobra CLI) → internal/chzzkmcp → pkg/chzzk`

```
cmd/chzzk-mcp-server/main.go   Cobra CLI. stdio 서브커맨드만 존재.
internal/chzzkmcp/server.go    RunStdioServer() — mcp.StdioTransport로 서버 실행.
pkg/chzzk/
  api_reference.go             모든 Chzzk API 데이터. Endpoint 슬라이스(AllEndpoints),
                               카테고리별 조회(EndpointsByCategory), 키 조회(FindEndpoint).
                               새 엔드포인트는 여기에만 추가하면 된다.
  server.go                    NewMCPServer() — RegisterReferenceTools + RegisterCodegenTools 호출.
  tools_reference.go           chzzk_list_apis, chzzk_get_api_spec 핸들러 + 등록 함수.
  tools_codegen.go             chzzk_generate_auth_code, chzzk_generate_api_client,
                               chzzk_scaffold_project 핸들러 + 등록 함수.
                               코드 생성 로직(authCodeGo, apiClientGo, scaffoldGo 등)도 포함.
```

## 도구 등록 패턴

새 MCP 도구를 추가할 때의 패턴:

```go
// 1. Input 구조체 정의 (jsonschema 태그로 스키마 자동 생성)
type MyToolInput struct {
    Language string `json:"language" jsonschema:"설명"`
}

// 2. 핸들러 함수 (3번째 반환값은 handler 오류, 도구 오류는 IsError:true로 반환)
func handleMyTool(_ context.Context, _ *mcp.CallToolRequest, input MyToolInput) (*mcp.CallToolResult, any, error) {
    return textResult("결과"), nil, nil  // 성공
    return errorResult("오류 메시지"), nil, nil  // 도구 오류
}

// 3. Register 함수에서 mcp.AddTool로 등록
mcp.AddTool(s, &mcp.Tool{Name: "chzzk_my_tool", Description: "..."}, handleMyTool)
```

`textResult()` / `errorResult()` 헬퍼는 `tools_codegen.go`에 정의되어 있다.

## Go 코드 내 백틱(`) 처리

TypeScript 템플릿 리터럴(`` `${...}` ``)을 Go 문자열에 포함할 때는 `bt := "` `` `` `"` 변수를 사용한다 (Go raw string 안에 backtick을 넣을 수 없기 때문):

```go
bt := "`"
sb.WriteString("let url = " + bt + "${BASE_URL}${path}" + bt + ";")
```

double-quoted Go 문자열 안에서 `"` 를 포함한 TypeScript 코드를 쓸 때는 `\"` 로 이스케이프한다.

## API 레퍼런스 데이터 구조

`api_reference.go`의 `Endpoint` 구조체가 핵심 데이터 모델이다:
- `AuthType`: `none` / `client_credentials` / `access_token`
- `Key()`: `"METHOD /path"` 형식 → `FindEndpoint()`의 조회 키
- `AllEndpoints` 슬라이스 선언 순서가 `chzzk_list_apis` 반환 순서를 결정한다

## MCP SDK

`github.com/modelcontextprotocol/go-sdk v1.6.1` 사용. 핵심 API:
- `mcp.NewServer(impl, opts)` — 서버 생성
- `mcp.AddTool[In, Out](server, tool, handler)` — 제네릭 도구 등록 (Input 스키마 자동 추론)
- `mcp.StdioTransport{}` — stdio 트랜스포트
- `mcp.CallToolResult{IsError: true}` — 도구 레벨 오류 (handler error는 프로토콜 오류)
