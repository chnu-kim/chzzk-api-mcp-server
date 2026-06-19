package chzzk

import (
	"go/format"
	"strings"
	"testing"
)

// ─── tsMethodForEndpoint ──────────────────────────────────────────────────────

func TestTsMethodForEndpoint_OptionalQueryParams(t *testing.T) {
	ep := Endpoint{
		Method:      "GET",
		Path:        "/test",
		QueryParams: []Param{{Name: "page", Type: "int", Required: false}},
	}
	code := tsMethodForEndpoint(ep)
	if !strings.Contains(code, "!== undefined") {
		t.Errorf("expected optional param guard ('!== undefined') in TS method; got:\n%s", code)
	}
}

func TestTsMethodForEndpoint_BodyAndAccessToken(t *testing.T) {
	ep := Endpoint{
		Method:     "POST",
		Path:       "/test",
		AuthType:   AuthTypeAccessToken,
		BodyParams: []Param{{Name: "message", Type: "string", Required: true}},
		Response:   []ResponseField{{Name: "ok", Type: "bool"}},
	}
	code := tsMethodForEndpoint(ep)
	if !strings.Contains(code, "body:") {
		t.Errorf("expected 'body:' for BodyParams endpoint; got:\n%s", code)
	}
	if strings.Contains(code, "const query") {
		t.Errorf("unexpected 'const query' block when no QueryParams exist; got:\n%s", code)
	}
	if !strings.Contains(code, "auth: true") {
		t.Errorf("expected 'auth: true' for access_token endpoint; got:\n%s", code)
	}
}

func TestTsMethodForEndpoint_VoidResponse(t *testing.T) {
	ep := Endpoint{Method: "GET", Path: "/test"}
	code := tsMethodForEndpoint(ep)
	if !strings.Contains(code, "Promise<void>") {
		t.Errorf("expected 'Promise<void>' for response-less endpoint; got:\n%s", code)
	}
}

func TestTsMethodForEndpoint_RequiredAndOptionalQueryParams(t *testing.T) {
	ep := Endpoint{
		Method: "GET",
		Path:   "/test",
		QueryParams: []Param{
			{Name: "keyword", Type: "string", Required: true},
			{Name: "size", Type: "int", Required: false},
		},
	}
	code := tsMethodForEndpoint(ep)

	if !strings.Contains(code, "keyword: string") {
		t.Errorf("expected required param in function signature without '?'; got:\n%s", code)
	}
	if !strings.Contains(code, "size?: ") {
		t.Errorf("expected optional param with '?' in function signature; got:\n%s", code)
	}
	if strings.Contains(code, "if (keyword !== undefined)") {
		t.Errorf("required param should not have an undefined guard; got:\n%s", code)
	}
	if !strings.Contains(code, "!== undefined") {
		t.Errorf("expected undefined guard for optional param; got:\n%s", code)
	}
}

// ─── goFieldName ──────────────────────────────────────────────────────────────

func TestGoFieldName(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		// camelCase 실제 API 응답 필드명 (Id/Url suffix → acronym 변환)
		{"channelId", "ChannelID"},
		{"liveId", "LiveID"},
		{"categoryId", "CategoryID"},
		{"channelImageUrl", "ChannelImageURL"},
		{"liveThumbnailImageUrl", "LiveThumbnailImageURL"},
		{"posterImageUrl", "PosterImageURL"},
		// snake_case (기존 동작 보존)
		{"channel_id", "ChannelID"},
		{"channel_url", "ChannelURL"},
		// 단독 두문자어
		{"id", "ID"},
		{"url", "URL"},
		// 두문자어 아닌 필드 — suffix에 id/url을 포함하지 않으면 변환 없음
		{"expiresIn", "ExpiresIn"},
		{"accessToken", "AccessToken"},
		{"followerCount", "FollowerCount"},
		{"concurrentUserCount", "ConcurrentUserCount"},
		{"verifiedMark", "VerifiedMark"},
		// "id/url"을 부분 포함하지만 suffix가 아닌 경우 → 변환 없음 (false positive 방지)
		{"studio", "Studio"},
		{"valid", "Valid"},
		{"periods", "Periods"},
	}
	for _, tc := range cases {
		if got := goFieldName(tc.input); got != tc.want {
			t.Errorf("goFieldName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ─── goParamName ─────────────────────────────────────────────────────────────

func TestGoParamName(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"channelId", "channelId"},
		{"channel_id", "channelId"},
		{"page", "page"},
		{"size", "size"},
		{"messageTime", "messageTime"},
		{"minFollowerMinute", "minFollowerMinute"},
		{"allowSubscriberInFollowerMode", "allowSubscriberInFollowerMode"},
		{"channelIds[]", "channelIds"},
		// 점으로 끝나는 이름 처리 (TrimSuffix 경로)
		{"param.", "param"},
	}
	for _, tc := range cases {
		if got := goParamName(tc.input); got != tc.want {
			t.Errorf("goParamName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ─── goMethodName ─────────────────────────────────────────────────────────────

func TestGoMethodName(t *testing.T) {
	cases := []struct {
		method string
		path   string
		want   string
	}{
		{"GET", "/open/v1/lives", "GetLives"},
		{"POST", "/open/v1/chats/send", "CreateChatsSend"},
		{"GET", "/open/v1/channels", "GetChannels"},
		{"PUT", "/open/v1/chats/settings", "UpdateChatsSettings"},
		{"PATCH", "/open/v1/chats/settings", "UpdateChatsSettings"},
		{"DELETE", "/open/v1/sessions/events/subscribe/chat", "DeleteSessionsEventsSubscribeChat"},
		{"POST", "/auth/v1/token", "CreateToken"},
		// 알 수 없는 HTTP 메서드 — prefix 없이 path만 사용
		{"HEAD", "/open/v1/lives", "Lives"},
	}
	for _, tc := range cases {
		ep := Endpoint{Method: tc.method, Path: tc.path}
		if got := goMethodName(ep); got != tc.want {
			t.Errorf("goMethodName(%s %s) = %q, want %q", tc.method, tc.path, got, tc.want)
		}
	}
}

// ─── goType / goZero / tsType ─────────────────────────────────────────────────

func TestGoType(t *testing.T) {
	cases := []struct{ input, want string }{
		{"int", "int"},
		{"long", "int"},
		{"bool", "bool"},
		{"string[]", "[]string"},
		{"string", "string"},
		{"", "string"}, // default (알 수 없는 타입)
	}
	for _, tc := range cases {
		if got := goType(tc.input); got != tc.want {
			t.Errorf("goType(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestGoZero(t *testing.T) {
	cases := []struct{ input, want string }{
		{"int", "0"},
		{"long", "0"},
		{"bool", "false"},
		{"string", `""`},
		{"string[]", `""`},
		{"", `""`}, // default (알 수 없는 타입)
	}
	for _, tc := range cases {
		if got := goZero(tc.input); got != tc.want {
			t.Errorf("goZero(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestTsType(t *testing.T) {
	cases := []struct{ input, want string }{
		{"int", "number"},
		{"long", "number"},
		{"bool", "boolean"},
		{"string[]", "string[]"},
		{"string", "string"},
		{"", "string"}, // default (알 수 없는 타입)
	}
	for _, tc := range cases {
		if got := tsType(tc.input); got != tc.want {
			t.Errorf("tsType(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ─── Go 코드 생성 결과 문법 유효성 (go/format 파서 사용) ─────────────────────

// mustClientGoCode는 단일 엔드포인트로 Go 클라이언트 코드를 생성하고 문법 유효성을 검증한다.
func mustClientGoCode(t *testing.T, key string) string {
	t.Helper()
	ep, ok := FindEndpoint(key)
	if !ok {
		t.Fatalf("endpoint not found: %s", key)
	}
	code := apiClientGo([]Endpoint{ep})
	if _, err := format.Source([]byte(code)); err != nil {
		t.Fatalf("generated Go code is not valid Go: %v", err)
	}
	return code
}

func TestApiClientGoSyntaxValid(t *testing.T) {
	endpoints := []string{
		"GET /open/v1/channels",           // string[], string, int, bool Response 필드
		"GET /open/v1/channels/followers",  // int optional QueryParams, access_token 인증
		"GET /open/v1/lives",              // int/bool/string[] Response, int optional QueryParam
		"POST /open/v1/chats/send",        // BodyParams, access_token 인증
		"PUT /open/v1/chats/settings",     // int/bool BodyParams
		"GET /open/v1/streams/key",        // Response 없음 → error만 반환
	}

	var eps []Endpoint
	for _, key := range endpoints {
		ep, ok := FindEndpoint(key)
		if !ok {
			t.Fatalf("endpoint not found: %s", key)
		}
		eps = append(eps, ep)
	}

	if _, err := format.Source([]byte(apiClientGo(eps))); err != nil {
		t.Errorf("generated Go client code is not valid Go: %v", err)
	}
}

func TestApiClientGoFieldNameAcronyms(t *testing.T) {
	code := mustClientGoCode(t, "GET /open/v1/channels")

	if !strings.Contains(code, "ChannelID") {
		t.Errorf("expected ChannelID (not ChannelId) in generated Go code; got:\n%s", code)
	}
	if !strings.Contains(code, "ChannelImageURL") {
		t.Errorf("expected ChannelImageURL (not ChannelImageUrl) in generated Go code; got:\n%s", code)
	}
}

func TestApiClientGoLivesEndpoint(t *testing.T) {
	// GET /open/v1/lives: Response에 int/bool/string[], optional int QueryParam
	code := mustClientGoCode(t, "GET /open/v1/lives")

	if !strings.Contains(code, "bool") {
		t.Error("expected bool field type in generated Go code")
	}
	if !strings.Contains(code, "[]string") {
		t.Error("expected []string field type in generated Go code")
	}
	if !strings.Contains(code, "strconv.Itoa") {
		t.Error("expected strconv.Itoa for int QueryParam in generated Go code")
	}
}

func TestApiClientGoResponselessEndpoint(t *testing.T) {
	// Response 없는 엔드포인트 → 반환 타입 error, nil 반환 코드 경로
	code := mustClientGoCode(t, "GET /open/v1/streams/key")

	if strings.Contains(code, "Response struct") {
		t.Error("response-less endpoint should not generate a Response struct")
	}
	if !strings.Contains(code, "return nil") {
		t.Error("response-less endpoint should generate 'return nil' on success")
	}
}
