package chzzk

import (
	"strings"
)

func wsClientGo(events map[string]bool) string {
	var sb strings.Builder

	sb.WriteString(`// WebSocket 클라이언트 — 치지직 실시간 이벤트 수신
// 의존성: go get github.com/gorilla/websocket
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

const baseURL = "https://openapi.chzzk.naver.com"

type apiResponse[T any] struct {
	Code    int     ` + "`" + `json:"code"` + "`" + `
	Message *string ` + "`" + `json:"message"` + "`" + `
	Content T       ` + "`" + `json:"content"` + "`" + `
}

type sessionAuthContent struct {
	WebSocketURL string ` + "`" + `json:"webSocketUrl"` + "`" + `
}

type sessionContent struct {
	SessionKey string ` + "`" + `json:"sessionKey"` + "`" + `
}

type sessionListContent struct {
	Data []sessionContent ` + "`" + `json:"data"` + "`" + `
}

// getWebSocketURL calls GET /open/v1/sessions/auth and returns the webSocketUrl.
func getWebSocketURL(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/open/v1/sessions/auth", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result apiResponse[sessionAuthContent]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Code != http.StatusOK {
		msg := "unknown error"
		if result.Message != nil {
			msg = *result.Message
		}
		return "", fmt.Errorf("sessions/auth error %d: %s", result.Code, msg)
	}
	return result.Content.WebSocketURL, nil
}

// getSessionKey calls GET /open/v1/sessions and returns the first active session key.
func getSessionKey(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/open/v1/sessions", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result apiResponse[sessionListContent]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Content.Data) == 0 {
		return "", fmt.Errorf("no active session found")
	}
	return result.Content.Data[0].SessionKey, nil
}

`)

	if events["chat"] {
		sb.WriteString(`// subscribeChat calls POST /open/v1/sessions/events/subscribe/chat.
func subscribeChat(ctx context.Context, accessToken, sessionKey string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		baseURL+"/open/v1/sessions/events/subscribe/chat?sessionKey="+sessionKey, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

`)
	}

	if events["donation"] {
		sb.WriteString(`// subscribeDonation calls POST /open/v1/sessions/events/subscribe/donation.
func subscribeDonation(ctx context.Context, accessToken, sessionKey string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		baseURL+"/open/v1/sessions/events/subscribe/donation?sessionKey="+sessionKey, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

`)
	}

	sb.WriteString(`func main() {
	accessToken := os.Getenv("CHZZK_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("CHZZK_ACCESS_TOKEN is required")
	}
	ctx := context.Background()

	// 1. WebSocket URL 발급
	wsURL, err := getWebSocketURL(ctx, accessToken)
	if err != nil {
		log.Fatalf("getWebSocketURL: %v", err)
	}

	// 2. WebSocket 연결
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		log.Fatalf("websocket.Dial: %v", err)
	}
	defer conn.Close()
	log.Println("WebSocket connected:", wsURL)

	// 3. 세션 키 조회 (연결 직후 GET /open/v1/sessions)
	sessionKey, err := getSessionKey(ctx, accessToken)
	if err != nil {
		log.Fatalf("getSessionKey: %v", err)
	}

	// 4. 이벤트 구독
`)

	if events["chat"] {
		sb.WriteString(`	if err := subscribeChat(ctx, accessToken, sessionKey); err != nil {
		log.Fatalf("subscribeChat: %v", err)
	}
`)
	}
	if events["donation"] {
		sb.WriteString(`	if err := subscribeDonation(ctx, accessToken, sessionKey); err != nil {
		log.Fatalf("subscribeDonation: %v", err)
	}
`)
	}

	sb.WriteString(`
	// 5. 이벤트 수신 루프
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("ReadMessage: %v", err)
			return
		}
		var event map[string]any
		if err := json.Unmarshal(msg, &event); err != nil {
			log.Printf("unmarshal: %v — raw: %s", err, msg)
			continue
		}
		fmt.Printf("event: %v\n", event)
	}
}
`)

	return sb.String()
}

func wsClientTypeScript(events map[string]bool) string {
	bt := "`"
	var sb strings.Builder

	sb.WriteString(`// WebSocket 클라이언트 — 치지직 실시간 이벤트 수신
// Node.js 18+ (fetch/WebSocket 내장) 또는 ws 패키지 사용
// 환경 변수: CHZZK_ACCESS_TOKEN

const BASE_URL = "https://openapi.chzzk.naver.com";

interface ApiResponse<T> {
  code: number;
  message?: string;
  content: T;
}

interface SessionAuthContent {
  webSocketUrl: string;
}

interface SessionContent {
  sessionKey: string;
  connectedDate: string;
  subscribedEvents: string[];
}

async function getWebSocketUrl(accessToken: string): Promise<string> {
  const resp = await fetch(` + bt + `${BASE_URL}/open/v1/sessions/auth` + bt + `, {
    headers: { Authorization: ` + bt + `Bearer ${accessToken}` + bt + ` },
  });
  const json: ApiResponse<SessionAuthContent> = await resp.json();
  if (json.code !== 200) throw new Error(` + bt + `sessions/auth error ${json.code}: ${json.message}` + bt + `);
  return json.content.webSocketUrl;
}

async function getSessionKey(accessToken: string): Promise<string> {
  const resp = await fetch(` + bt + `${BASE_URL}/open/v1/sessions` + bt + `, {
    headers: { Authorization: ` + bt + `Bearer ${accessToken}` + bt + ` },
  });
  const json: ApiResponse<{ data: SessionContent[] }> = await resp.json();
  if (!json.content.data.length) throw new Error("no active session found");
  return json.content.data[0].sessionKey;
}

`)

	if events["chat"] {
		sb.WriteString(`async function subscribeChat(accessToken: string, sessionKey: string): Promise<void> {
  await fetch(` + bt + `${BASE_URL}/open/v1/sessions/events/subscribe/chat?sessionKey=${sessionKey}` + bt + `, {
    method: "POST",
    headers: { Authorization: ` + bt + `Bearer ${accessToken}` + bt + ` },
  });
}

`)
	}

	if events["donation"] {
		sb.WriteString(`async function subscribeDonation(accessToken: string, sessionKey: string): Promise<void> {
  await fetch(` + bt + `${BASE_URL}/open/v1/sessions/events/subscribe/donation?sessionKey=${sessionKey}` + bt + `, {
    method: "POST",
    headers: { Authorization: ` + bt + `Bearer ${accessToken}` + bt + ` },
  });
}

`)
	}

	sb.WriteString(`async function main() {
  const accessToken = process.env.CHZZK_ACCESS_TOKEN;
  if (!accessToken) throw new Error("CHZZK_ACCESS_TOKEN is required");

  // 1. WebSocket URL 발급
  const wsUrl = await getWebSocketUrl(accessToken);

  // 2. WebSocket 연결
  const ws = new WebSocket(wsUrl);

  ws.addEventListener("open", async () => {
    console.log("WebSocket connected:", wsUrl);

    // 3. 세션 키 조회
    const sessionKey = await getSessionKey(accessToken);

    // 4. 이벤트 구독
`)

	if events["chat"] {
		sb.WriteString(`    await subscribeChat(accessToken, sessionKey);
`)
	}
	if events["donation"] {
		sb.WriteString(`    await subscribeDonation(accessToken, sessionKey);
`)
	}

	sb.WriteString(`  });

  // 5. 이벤트 수신
  ws.addEventListener("message", (event) => {
    const data = JSON.parse(event.data as string);
    console.log("event:", data);
  });

  ws.addEventListener("error", (err) => console.error("WebSocket error:", err));
  ws.addEventListener("close", () => console.log("WebSocket closed"));
}

main().catch(console.error);
`)

	return sb.String()
}
