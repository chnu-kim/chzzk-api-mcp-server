# chzzk-api-mcp-server

치지직(Chzzk) Open API 연동 서비스 개발을 돕는 MCP(Model Context Protocol) 서버.

Claude 등 MCP 클라이언트에서 Chzzk API 레퍼런스를 조회하고, Go/TypeScript 연동 코드를 즉시 생성할 수 있습니다.

## 제공 도구

| 도구 | 설명 |
|------|------|
| `chzzk_list_apis` | 카테고리별 API 엔드포인트 목록 조회 |
| `chzzk_get_api_spec` | 특정 엔드포인트 상세 스펙 (파라미터·응답·인증) |
| `chzzk_generate_auth_code` | OAuth2 인증 플로우 완성 코드 생성 |
| `chzzk_generate_api_client` | 지정 엔드포인트 타입 안전 HTTP 클라이언트 생성 |
| `chzzk_scaffold_project` | 프로젝트 보일러플레이트 생성 |

지원 언어: **Go**, **TypeScript**

## 설치

```bash
git clone https://github.com/chanuuuu/chzzk-api-mcp-server
cd chzzk-api-mcp-server
go build -o chzzk-mcp-server ./cmd/chzzk-mcp-server/
```

## Claude Desktop 연동

`~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "chzzk": {
      "command": "/path/to/chzzk-mcp-server",
      "args": ["stdio"]
    }
  }
}
```

## 사용 예시

**API 목록 조회**
```
chzzk_list_apis(category: "live")
→ GET /open/v1/lives, GET /open/v1/lives/setting, PATCH /open/v1/lives/setting ...
```

**엔드포인트 스펙 확인**
```
chzzk_get_api_spec(endpoint: "GET /open/v1/lives")
→ 쿼리 파라미터(size, next), 응답 필드(liveId, liveTitle, concurrentUserCount...), 인증 타입 반환
```

**Go 클라이언트 생성**
```
chzzk_generate_api_client(language: "go", endpoints: ["GET /open/v1/lives", "POST /open/v1/chats/send"])
→ 타입이 정의된 완성 Go 클라이언트 코드 반환
```

**프로젝트 스캐폴드**
```
chzzk_scaffold_project(language: "go", project_name: "my-chzzk-bot", features: ["auth", "live", "chat"])
→ 디렉토리 구조 + 핵심 파일 코드 반환
```

## Chzzk API 카테고리

`auth` · `user` · `channel` · `category` · `live` · `chat` · `session` · `drops` · `restriction`

- Base URL: `https://openapi.chzzk.naver.com`
- 공식 문서: https://chzzk.gitbook.io/chzzk

## 개발

```bash
go test ./...        # 테스트 실행
go build ./...       # 빌드 검증
```

새 엔드포인트 추가: `pkg/chzzk/api_reference.go`의 `AllEndpoints` 슬라이스에 `Endpoint` 항목 추가.
새 도구 추가: `pkg/chzzk/tools_*.go` 파일에 핸들러 구현 후 `RegisterXxxTools()`에 등록.
