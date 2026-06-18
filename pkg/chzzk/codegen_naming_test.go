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
		// 두문자어 아닌 필드 — suffix에 id/url을 포함하지 않으면 변환 없음
		{"expiresIn", "ExpiresIn"},
		{"accessToken", "AccessToken"},
		{"followerCount", "FollowerCount"},
		{"concurrentUserCount", "ConcurrentUserCount"},
		{"verifiedMark", "VerifiedMark"},
		// "id/url"을 부분 포함하지만 suffix가 아닌 경우 → 변환 없음 (false positive 방지)
		{"studio", "Studio"},     // "udio"로 끝남, "id" suffix 아님
		{"valid", "Valid"},        // "alid"로 끝남, "Id" suffix 아님
		{"periods", "Periods"},    // "iods"로 끝남, "Ids" suffix 아님
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
		{"channelId", "channelId"},                                            // camelCase 보존
		{"channel_id", "channelId"},                                           // snake_case → camelCase
		{"page", "page"},
		{"size", "size"},
		{"messageTime", "messageTime"},                                        // camelCase 보존
		{"minFollowerMinute", "minFollowerMinute"},                             // camelCase 보존
		{"allowSubscriberInFollowerMode", "allowSubscriberInFollowerMode"},     // camelCase 보존
		{"channelIds[]", "channelIds"},                                        // [] suffix 제거
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

func TestApiClientGoSyntaxValid(t *testing.T) {
	endpoints := []string{
		"GET /open/v1/channels",           // string[], string, int, bool Response 필드
		"GET /open/v1/channels/followers",  // int optional QueryParams, access_token 인증
		"GET /open/v1/lives",              // int/bool/string[] Response, int optional QueryParam
		"POST /open/v1/chats/send",        // BodyParams, access_token 인증
		"PUT /open/v1/chats/settings",     // int/bool BodyParams
		"GET /open/v1/streams/key",        // Response 없음 → void(error만 반환) 코드 경로
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
		t.Errorf("expected ChannelID (not ChannelId) in generated Go code; got:\n%s", code)
	}
	if !strings.Contains(code, "ChannelImageURL") {
		t.Errorf("expected ChannelImageURL (not ChannelImageUrl) in generated Go code; got:\n%s", code)
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
	// optional int QueryParam → strconv.Itoa 조건부 생성 확인
	// GET /open/v1/lives: optional int QueryParam "size"
	ep, ok := FindEndpoint("GET /open/v1/lives")
	if !ok {
		t.Fatal("endpoint not found: GET /open/v1/lives")
	}

	code := apiClientGo([]Endpoint{ep})

	if _, err := format.Source([]byte(code)); err != nil {
		t.Errorf("generated Go code is not valid Go: %v", err)
	}
	if !strings.Contains(code, "strconv.Itoa") {
		t.Error("expected strconv.Itoa for int QueryParam in generated Go code")
	}
}

func TestApiClientGoResponselessEndpoint(t *testing.T) {
	// Response 없는 엔드포인트 → 반환 타입 error, nil 반환 코드 경로
	ep, ok := FindEndpoint("GET /open/v1/streams/key")
	if !ok {
		t.Fatal("endpoint not found: GET /open/v1/streams/key")
	}

	code := apiClientGo([]Endpoint{ep})

	if _, err := format.Source([]byte(code)); err != nil {
		t.Errorf("generated Go code for response-less endpoint is not valid Go: %v\n\ncode:\n%s", err, code)
	}
	// Response 없는 엔드포인트는 반환 타입이 error만이어야 함
	if strings.Contains(code, "Response struct") {
		t.Error("response-less endpoint should not generate a Response struct")
	}
	if !strings.Contains(code, "return nil") {
		t.Error("response-less endpoint should generate 'return nil' on success")
	}
}
