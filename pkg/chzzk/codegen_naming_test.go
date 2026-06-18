package chzzk

import (
	"go/format"
	"strings"
	"testing"
)

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
		// 복수형
		{"channelIds[]", "ChannelIDs"},
		// 두문자어 아닌 필드 (변형 없음)
		{"expiresIn", "ExpiresIn"},
		{"accessToken", "AccessToken"},
		{"followerCount", "FollowerCount"},
		{"concurrentUserCount", "ConcurrentUserCount"},
		{"verifiedMark", "VerifiedMark"},
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
		{"channelId", "channelId"},                             // camelCase 보존
		{"channel_id", "channelId"},                           // snake_case → camelCase
		{"page", "page"},
		{"size", "size"},
		{"messageTime", "messageTime"},                        // camelCase 보존
		{"minFollowerMinute", "minFollowerMinute"},             // camelCase 보존
		{"allowSubscriberInFollowerMode", "allowSubscriberInFollowerMode"}, // camelCase 보존
		{"channelIds[]", "channelIds"},                        // [] suffix 제거
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
		{"", "string"}, // default
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
		{"", "string"}, // default
	}
	for _, tc := range cases {
		if got := tsType(tc.input); got != tc.want {
			t.Errorf("tsType(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ─── Go 코드 생성 결과 문법 유효성 (go/format 파서 사용) ─────────────────────

func TestApiClientGoSyntaxValid(t *testing.T) {
	// int, bool, string[], access_token 인증 등 다양한 타입을 포함하는 엔드포인트
	endpoints := []string{
		"GET /open/v1/channels",          // string[], string, int, bool Response 필드
		"GET /open/v1/channels/followers", // int optional QueryParams, access_token 인증
		"GET /open/v1/lives",             // int/bool/string[] Response, int optional QueryParam
		"POST /open/v1/chats/send",       // BodyParams, access_token 인증
		"PUT /open/v1/chats/settings",     // int/bool BodyParams
	}

	var eps []Endpoint
	for _, key := range endpoints {
		ep, ok := FindEndpoint(key)
		if !ok {
			t.Fatalf("endpoint not found: %s", key)
		}
		eps = append(eps, ep)
	}

	code := apiClientGo(eps)

	if _, err := format.Source([]byte(code)); err != nil {
		t.Errorf("generated Go client code is not valid Go:\n%v\n\ncode:\n%s", err, code)
	}
}

func TestApiClientGoFieldNameAcronyms(t *testing.T) {
	// channelId, channelImageUrl 포함 엔드포인트로 Go 컨벤션 준수 확인
	ep, ok := FindEndpoint("GET /open/v1/channels")
	if !ok {
		t.Fatal("endpoint not found: GET /open/v1/channels")
	}

	code := apiClientGo([]Endpoint{ep})

	if !strings.Contains(code, "ChannelID") {
		t.Errorf("expected ChannelID (not ChannelId) in generated Go code; snippet:\n%s", code)
	}
	if !strings.Contains(code, "ChannelImageURL") {
		t.Errorf("expected ChannelImageURL (not ChannelImageUrl) in generated Go code; snippet:\n%s", code)
	}
}

func TestApiClientGoBoolAndIntInResponse(t *testing.T) {
	// GET /open/v1/lives: Response에 int(concurrentUserCount), bool(adult), string[](tags)
	ep, ok := FindEndpoint("GET /open/v1/lives")
	if !ok {
		t.Fatal("endpoint not found: GET /open/v1/lives")
	}

	code := apiClientGo([]Endpoint{ep})

	if _, err := format.Source([]byte(code)); err != nil {
		t.Errorf("generated Go code is not valid Go: %v", err)
	}
	if !strings.Contains(code, "bool") {
		t.Error("expected bool field type in generated Go code")
	}
	if !strings.Contains(code, "[]string") {
		t.Error("expected []string field type in generated Go code")
	}
}

func TestApiClientGoIntQueryParamUsesStrconv(t *testing.T) {
	// int type required QueryParam → strconv.Itoa 생성 확인
	// POST /open/v1/chats/blind-message: BodyParams에 long(messageTime) 포함
	ep, ok := FindEndpoint("POST /open/v1/chats/blind-message")
	if !ok {
		t.Fatal("endpoint not found: POST /open/v1/chats/blind-message")
	}

	code := apiClientGo([]Endpoint{ep})

	if _, err := format.Source([]byte(code)); err != nil {
		t.Errorf("generated Go code is not valid Go: %v", err)
	}
	// long 타입 → int로 변환
	if !strings.Contains(code, "int") {
		t.Error("expected int type for long param in generated Go code")
	}
}
