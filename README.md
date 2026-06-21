# chzzk-api-mcp-server

치지직(Chzzk) Open API 연동 서비스 개발을 돕는 MCP(Model Context Protocol) 서버.

Claude 등 MCP 클라이언트에서 두 가지 방식으로 개발 속도를 높입니다.

- **API 레퍼런스 조회** — 엔드포인트·파라미터·응답 스펙을 문서 없이 즉시 확인
- **코드 생성** — OAuth2 인증 플로우, 타입 안전 HTTP 클라이언트, 프로젝트 스캐폴드를 언어별로 즉시 생성

## 제공 도구

| 도구 | 설명 |
|------|------|
| `chzzk_list_apis` | 카테고리별 API 엔드포인트 목록 조회 |
| `chzzk_get_api_spec` | 특정 엔드포인트 상세 스펙 (파라미터·응답·인증) |
| `chzzk_generate_auth_code` | OAuth2 인증 플로우 완성 코드 생성 |
| `chzzk_generate_api_client` | 지정 엔드포인트 타입 안전 HTTP 클라이언트 생성 |
| `chzzk_scaffold_project` | 프로젝트 보일러플레이트 생성 |
| `chzzk_generate_websocket_client` | 채팅·후원 실시간 이벤트를 수신하는 WebSocket 클라이언트 코드 생성 |

지원 언어: **Go**, **TypeScript**

## 설치

**go install** (권장)

```bash
go install github.com/chanuuuu/chzzk-api-mcp-server/cmd/chzzk-mcp-server@latest
```

**소스 빌드**

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

**WebSocket 실시간 이벤트 클라이언트**
```
chzzk_generate_websocket_client(language: "go", events: ["chat", "donation"])
→ Session API로 URL 발급 후 채팅·후원 이벤트를 구독하는 완성 클라이언트 코드 반환
```

## Chzzk API 카테고리

`auth` · `user` · `channel` · `category` · `live` · `chat` · `session` · `drops` · `restriction`

- Base URL: `https://openapi.chzzk.naver.com`
- 공식 문서: https://chzzk.gitbook.io/chzzk

## Roadmap

- [x] **WebSocket / 실시간 이벤트** — 채팅·후원 이벤트를 수신하는 WebSocket 클라이언트 코드 생성 (`chzzk_generate_websocket_client`)
- [ ] **Python 지원** — Go/TypeScript 외 Python 클라이언트 코드 생성

## 개발

```bash
go test ./...        # 테스트 실행
go build ./...       # 빌드 검증
```

새 엔드포인트 추가: `pkg/chzzk/api_reference.go`의 `AllEndpoints` 슬라이스에 `Endpoint` 항목 추가.
새 도구 추가: `pkg/chzzk/tools_*.go` 파일에 핸들러 구현 후 `RegisterXxxTools()`에 등록.
