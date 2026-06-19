package chzzk

const BaseURL = "https://openapi.chzzk.naver.com"
const AuthURL = "https://chzzk.naver.com/account-interlock"
const TokenURL = BaseURL + "/auth/v1/token"

type AuthType string

const (
	AuthTypeNone              AuthType = "none"
	AuthTypeClientCredentials AuthType = "client_credentials"
	AuthTypeAccessToken       AuthType = "access_token"
)

type Category string

const (
	CategoryAuth        Category = "auth"
	CategoryUser        Category = "user"
	CategoryChannel     Category = "channel"
	CategoryCategory    Category = "category"
	CategoryLive        Category = "live"
	CategoryChat        Category = "chat"
	CategorySession     Category = "session"
	CategoryDrops       Category = "drops"
	CategoryRestriction Category = "restriction"
)

type Param struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     string `json:"default,omitempty"`
	Min         string `json:"min,omitempty"`
	Max         string `json:"max,omitempty"`
}

type ResponseField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type Endpoint struct {
	Method      string          `json:"method"`
	Path        string          `json:"path"`
	Category    Category        `json:"category"`
	Description string          `json:"description"`
	AuthType    AuthType        `json:"auth_type"`
	Scope       string          `json:"scope,omitempty"`
	QueryParams []Param         `json:"query_params,omitempty"`
	PathParams  []Param         `json:"path_params,omitempty"`
	BodyParams  []Param         `json:"body_params,omitempty"`
	Response    []ResponseField `json:"response,omitempty"`
}

// Key returns "METHOD /path" for use as a lookup key.
func (e Endpoint) Key() string {
	return e.Method + " " + e.Path
}

// AllEndpoints contains every Chzzk Open API endpoint.
var AllEndpoints = []Endpoint{
	// ── Auth ────────────────────────────────────────────────────────────────
	{
		Method:      "GET",
		Path:        "/account-interlock",
		Category:    CategoryAuth,
		Description: "인가 코드 요청 (브라우저 리다이렉트). Base URL: https://chzzk.naver.com",
		AuthType:    AuthTypeNone,
		QueryParams: []Param{
			{Name: "clientId", Type: "string", Required: true, Description: "애플리케이션 클라이언트 ID"},
			{Name: "redirectUri", Type: "string", Required: true, Description: "인가 코드를 전달받을 리다이렉트 URI"},
			{Name: "state", Type: "string", Required: true, Description: "CSRF 방지용 임의 문자열"},
		},
		Response: []ResponseField{
			{Name: "code", Type: "string", Description: "인가 코드"},
			{Name: "state", Type: "string", Description: "요청 시 전달한 state 값"},
		},
	},
	{
		Method:      "POST",
		Path:        "/auth/v1/token",
		Category:    CategoryAuth,
		Description: "Access Token 발급 (Authorization Code Grant)",
		AuthType:    AuthTypeNone,
		BodyParams: []Param{
			{Name: "grantType", Type: "string", Required: true, Description: "authorization_code 고정값"},
			{Name: "clientId", Type: "string", Required: true, Description: "클라이언트 ID"},
			{Name: "clientSecret", Type: "string", Required: true, Description: "클라이언트 시크릿"},
			{Name: "code", Type: "string", Required: true, Description: "인가 코드"},
			{Name: "state", Type: "string", Required: true, Description: "인가 요청 시 사용한 state 값"},
		},
		Response: []ResponseField{
			{Name: "accessToken", Type: "string", Description: "API 호출에 사용할 액세스 토큰"},
			{Name: "refreshToken", Type: "string", Description: "액세스 토큰 갱신에 사용할 리프레시 토큰"},
			{Name: "tokenType", Type: "string", Description: "토큰 타입 (Bearer)"},
			{Name: "expiresIn", Type: "int", Description: "액세스 토큰 만료 시간(초)"},
		},
	},
	{
		Method:      "POST",
		Path:        "/auth/v1/token#refresh",
		Category:    CategoryAuth,
		Description: "Access Token 갱신 (Refresh Token Grant)",
		AuthType:    AuthTypeNone,
		BodyParams: []Param{
			{Name: "grantType", Type: "string", Required: true, Description: "refresh_token 고정값"},
			{Name: "clientId", Type: "string", Required: true, Description: "클라이언트 ID"},
			{Name: "clientSecret", Type: "string", Required: true, Description: "클라이언트 시크릿"},
			{Name: "refreshToken", Type: "string", Required: true, Description: "갱신에 사용할 리프레시 토큰"},
		},
		Response: []ResponseField{
			{Name: "accessToken", Type: "string", Description: "새 액세스 토큰"},
			{Name: "refreshToken", Type: "string", Description: "새 리프레시 토큰"},
			{Name: "tokenType", Type: "string", Description: "Bearer"},
			{Name: "expiresIn", Type: "int", Description: "만료 시간(초)"},
			{Name: "scope", Type: "string", Description: "토큰 스코프"},
		},
	},
	{
		Method:      "POST",
		Path:        "/auth/v1/token/revoke",
		Category:    CategoryAuth,
		Description: "Token 폐기",
		AuthType:    AuthTypeNone,
		BodyParams: []Param{
			{Name: "clientId", Type: "string", Required: true, Description: "클라이언트 ID"},
			{Name: "clientSecret", Type: "string", Required: true, Description: "클라이언트 시크릿"},
			{Name: "token", Type: "string", Required: true, Description: "폐기할 토큰"},
			{Name: "tokenTypeHint", Type: "string", Required: false, Description: "토큰 타입 힌트 (access_token | refresh_token)"},
		},
	},

	// ── User ────────────────────────────────────────────────────────────────
	{
		Method:      "GET",
		Path:        "/open/v1/users/me",
		Category:    CategoryUser,
		Description: "로그인한 사용자의 채널 정보 조회",
		AuthType:    AuthTypeAccessToken,
		Scope:       "유저 정보 조회",
		Response: []ResponseField{
			{Name: "channelId", Type: "string", Description: "채널 ID"},
			{Name: "channelName", Type: "string", Description: "채널 이름"},
		},
	},

	// ── Channel ─────────────────────────────────────────────────────────────
	{
		Method:      "GET",
		Path:        "/open/v1/channels",
		Category:    CategoryChannel,
		Description: "채널 정보 조회 (최대 20개 동시 조회)",
		AuthType:    AuthTypeClientCredentials,
		QueryParams: []Param{
			{Name: "channelIds[]", Type: "string[]", Required: true, Description: "조회할 채널 ID 목록 (최대 20개)", Max: "20"},
		},
		Response: []ResponseField{
			{Name: "channelId", Type: "string", Description: "채널 ID"},
			{Name: "channelName", Type: "string", Description: "채널 이름"},
			{Name: "channelImageUrl", Type: "string", Description: "채널 이미지 URL"},
			{Name: "followerCount", Type: "int", Description: "팔로워 수"},
			{Name: "verifiedMark", Type: "bool", Description: "인증 마크 여부"},
		},
	},
	{
		Method:      "GET",
		Path:        "/open/v1/channels/streaming-roles",
		Category:    CategoryChannel,
		Description: "채널 방송 관리자 목록 조회",
		AuthType:    AuthTypeAccessToken,
		Scope:       "채널 관리자 조회",
		Response: []ResponseField{
			{Name: "managerChannelId", Type: "string", Description: "관리자 채널 ID"},
			{Name: "managerChannelName", Type: "string", Description: "관리자 채널 이름"},
			{Name: "userRole", Type: "string", Description: "역할 (소유자 | 관리자 | 운영자 | 정산관리자)"},
			{Name: "createdDate", Type: "string", Description: "등록 날짜"},
		},
	},
	{
		Method:      "GET",
		Path:        "/open/v1/channels/followers",
		Category:    CategoryChannel,
		Description: "채널 팔로워 목록 조회 (페이지네이션)",
		AuthType:    AuthTypeAccessToken,
		Scope:       "채널 팔로워 조회",
		QueryParams: []Param{
			{Name: "page", Type: "int", Required: false, Description: "페이지 번호", Default: "0"},
			{Name: "size", Type: "int", Required: false, Description: "페이지 당 항목 수", Default: "30", Min: "1", Max: "50"},
		},
		Response: []ResponseField{
			{Name: "channelId", Type: "string", Description: "팔로워 채널 ID"},
			{Name: "channelName", Type: "string", Description: "팔로워 채널 이름"},
			{Name: "createdDate", Type: "string", Description: "팔로우 날짜"},
		},
	},
	{
		Method:      "GET",
		Path:        "/open/v1/channels/subscribers",
		Category:    CategoryChannel,
		Description: "채널 구독자 목록 조회",
		AuthType:    AuthTypeAccessToken,
		Scope:       "채널 구독자 조회",
		QueryParams: []Param{
			{Name: "page", Type: "int", Required: false, Description: "페이지 번호", Default: "0"},
			{Name: "size", Type: "int", Required: false, Description: "페이지 당 항목 수", Default: "30", Min: "1", Max: "50"},
			{Name: "sort", Type: "string", Required: false, Description: "정렬 기준 (RECENT | LONGER)", Default: "RECENT"},
		},
		Response: []ResponseField{
			{Name: "channelId", Type: "string", Description: "구독자 채널 ID"},
			{Name: "channelName", Type: "string", Description: "구독자 채널 이름"},
			{Name: "month", Type: "int", Description: "구독 개월 수"},
			{Name: "tierNo", Type: "int", Description: "구독 티어"},
			{Name: "createdDate", Type: "string", Description: "구독 날짜"},
		},
	},

	// ── Category ─────────────────────────────────────────────────────────────
	{
		Method:      "GET",
		Path:        "/open/v1/categories/search",
		Category:    CategoryCategory,
		Description: "카테고리 검색",
		AuthType:    AuthTypeClientCredentials,
		QueryParams: []Param{
			{Name: "query", Type: "string", Required: true, Description: "검색 키워드"},
			{Name: "size", Type: "int", Required: false, Description: "반환 개수", Default: "20", Min: "1", Max: "50"},
		},
		Response: []ResponseField{
			{Name: "categoryType", Type: "string", Description: "카테고리 타입 (GAME | SPORTS | ETC)"},
			{Name: "categoryId", Type: "string", Description: "카테고리 ID"},
			{Name: "categoryValue", Type: "string", Description: "카테고리 이름"},
			{Name: "posterImageUrl", Type: "string", Description: "카테고리 포스터 이미지 URL"},
		},
	},

	// ── Live ─────────────────────────────────────────────────────────────────
	{
		Method:      "GET",
		Path:        "/open/v1/lives",
		Category:    CategoryLive,
		Description: "현재 라이브 목록 조회 (커서 페이지네이션)",
		AuthType:    AuthTypeClientCredentials,
		QueryParams: []Param{
			{Name: "size", Type: "int", Required: false, Description: "반환 개수", Default: "20", Min: "1", Max: "20"},
			{Name: "next", Type: "string", Required: false, Description: "다음 페이지 커서"},
		},
		Response: []ResponseField{
			{Name: "liveId", Type: "string", Description: "라이브 ID"},
			{Name: "liveTitle", Type: "string", Description: "방송 제목"},
			{Name: "liveThumbnailImageUrl", Type: "string", Description: "썸네일 URL"},
			{Name: "concurrentUserCount", Type: "int", Description: "동시 시청자 수"},
			{Name: "openDate", Type: "string", Description: "방송 시작 시각"},
			{Name: "adult", Type: "bool", Description: "성인 방송 여부"},
			{Name: "tags", Type: "string[]", Description: "태그 목록"},
			{Name: "categoryType", Type: "string", Description: "카테고리 타입"},
			{Name: "liveCategory", Type: "string", Description: "라이브 카테고리 ID"},
			{Name: "liveCategoryValue", Type: "string", Description: "라이브 카테고리 이름"},
			{Name: "channelId", Type: "string", Description: "채널 ID"},
			{Name: "channelName", Type: "string", Description: "채널 이름"},
			{Name: "channelImageUrl", Type: "string", Description: "채널 이미지 URL"},
		},
	},
	{
		Method:      "GET",
		Path:        "/open/v1/streams/key",
		Category:    CategoryLive,
		Description: "방송 스트림 키 조회",
		AuthType:    AuthTypeAccessToken,
		Scope:       "방송 스트림키 조회",
	},
	{
		Method:      "GET",
		Path:        "/open/v1/lives/setting",
		Category:    CategoryLive,
		Description: "방송 설정 조회",
		AuthType:    AuthTypeAccessToken,
		Scope:       "방송 설정 조회",
		Response: []ResponseField{
			{Name: "defaultLiveTitle", Type: "string", Description: "기본 방송 제목"},
			{Name: "category.categoryType", Type: "string", Description: "카테고리 타입 (GAME | SPORTS | ETC)"},
			{Name: "category.categoryId", Type: "string", Description: "카테고리 ID"},
			{Name: "category.categoryValue", Type: "string", Description: "카테고리 이름"},
			{Name: "category.posterImageUrl", Type: "string", Description: "카테고리 포스터 이미지 URL"},
			{Name: "tags", Type: "string[]", Description: "태그 목록"},
		},
	},
	{
		Method:      "PATCH",
		Path:        "/open/v1/lives/setting",
		Category:    CategoryLive,
		Description: "방송 설정 변경",
		AuthType:    AuthTypeAccessToken,
		Scope:       "방송 설정 변경",
		BodyParams: []Param{
			{Name: "defaultLiveTitle", Type: "string", Required: false, Description: "기본 방송 제목"},
			{Name: "categoryType", Type: "string", Required: false, Description: "카테고리 타입"},
			{Name: "categoryId", Type: "string", Required: false, Description: "카테고리 ID"},
			{Name: "tags[]", Type: "string[]", Required: false, Description: "태그 목록"},
		},
	},

	// ── Chat ─────────────────────────────────────────────────────────────────
	{
		Method:      "POST",
		Path:        "/open/v1/chats/send",
		Category:    CategoryChat,
		Description: "채팅 메시지 전송 (최대 100자)",
		AuthType:    AuthTypeAccessToken,
		Scope:       "채팅 메시지 쓰기",
		BodyParams: []Param{
			{Name: "message", Type: "string", Required: true, Description: "전송할 메시지 (최대 100자)", Max: "100"},
		},
		Response: []ResponseField{
			{Name: "messageId", Type: "string", Description: "전송된 메시지 ID"},
		},
	},
	{
		Method:      "POST",
		Path:        "/open/v1/chats/notice",
		Category:    CategoryChat,
		Description: "채팅 공지 등록",
		AuthType:    AuthTypeAccessToken,
		Scope:       "채팅 공지 쓰기",
		BodyParams: []Param{
			{Name: "message", Type: "string", Required: false, Description: "공지 메시지 내용 (message 또는 messageId 중 하나 필수)"},
			{Name: "messageId", Type: "string", Required: false, Description: "기존 메시지 ID (message 또는 messageId 중 하나 필수)"},
		},
	},
	{
		Method:      "GET",
		Path:        "/open/v1/chats/settings",
		Category:    CategoryChat,
		Description: "채팅 설정 조회",
		AuthType:    AuthTypeAccessToken,
		Scope:       "채팅 설정 조회",
		Response: []ResponseField{
			{Name: "chatAvailableCondition", Type: "string", Description: "채팅 가능 조건"},
			{Name: "chatAvailableGroup", Type: "string", Description: "채팅 가능 그룹"},
			{Name: "minFollowerMinute", Type: "int", Description: "최소 팔로우 시간(분)"},
			{Name: "allowSubscriberInFollowerMode", Type: "bool", Description: "팔로워 모드에서 구독자 허용 여부"},
			{Name: "chatSlowModeSec", Type: "int", Description: "슬로우 모드 초"},
			{Name: "chatEmojiMode", Type: "bool", Description: "이모지 전용 모드 여부"},
		},
	},
	{
		Method:      "PUT",
		Path:        "/open/v1/chats/settings",
		Category:    CategoryChat,
		Description: "채팅 설정 변경",
		AuthType:    AuthTypeAccessToken,
		Scope:       "채팅 설정 변경",
		BodyParams: []Param{
			{Name: "chatAvailableCondition", Type: "string", Required: false, Description: "채팅 가능 조건"},
			{Name: "chatAvailableGroup", Type: "string", Required: false, Description: "채팅 가능 그룹"},
			{Name: "minFollowerMinute", Type: "int", Required: false, Description: "최소 팔로우 시간(분)"},
			{Name: "allowSubscriberInFollowerMode", Type: "bool", Required: false, Description: "팔로워 모드에서 구독자 허용 여부"},
			{Name: "chatSlowModeSec", Type: "int", Required: false, Description: "슬로우 모드 초"},
			{Name: "chatEmojiMode", Type: "bool", Required: false, Description: "이모지 전용 모드 여부"},
		},
	},
	{
		Method:      "POST",
		Path:        "/open/v1/chats/blind-message",
		Category:    CategoryChat,
		Description: "채팅 메시지 블라인드 처리",
		AuthType:    AuthTypeAccessToken,
		Scope:       "채팅 메시지 쓰기",
		BodyParams: []Param{
			{Name: "chatChannelId", Type: "string", Required: true, Description: "채팅 채널 ID"},
			{Name: "messageTime", Type: "long", Required: true, Description: "메시지 전송 시각 (Unix timestamp milliseconds)"},
			{Name: "senderChannelId", Type: "string", Required: true, Description: "메시지 발신 채널 ID"},
		},
	},

	// ── Session ───────────────────────────────────────────────────────────────
	{
		Method:      "GET",
		Path:        "/open/v1/sessions/auth/client",
		Category:    CategorySession,
		Description: "WebSocket 연결 URL 발급 (클라이언트 인증, 최대 10 연결)",
		AuthType:    AuthTypeClientCredentials,
		Response: []ResponseField{
			{Name: "webSocketUrl", Type: "string", Description: "WebSocket 연결 URL"},
		},
	},
	{
		Method:      "GET",
		Path:        "/open/v1/sessions/auth",
		Category:    CategorySession,
		Description: "WebSocket 연결 URL 발급 (Access Token 인증, 사용자당 최대 3 연결)",
		AuthType:    AuthTypeAccessToken,
		Response: []ResponseField{
			{Name: "webSocketUrl", Type: "string", Description: "WebSocket 연결 URL"},
		},
	},
	{
		Method:      "GET",
		Path:        "/open/v1/sessions/client",
		Category:    CategorySession,
		Description: "세션 목록 조회 (클라이언트 인증)",
		AuthType:    AuthTypeClientCredentials,
		QueryParams: []Param{
			{Name: "page", Type: "int", Required: false, Description: "페이지 번호"},
			{Name: "size", Type: "int", Required: false, Description: "페이지 당 항목 수", Min: "1", Max: "50"},
		},
		Response: []ResponseField{
			{Name: "sessionKey", Type: "string", Description: "세션 키"},
			{Name: "connectedDate", Type: "string", Description: "연결 시각"},
			{Name: "disconnectedDate", Type: "string", Description: "연결 해제 시각"},
			{Name: "subscribedEvents", Type: "string[]", Description: "구독 중인 이벤트 목록"},
		},
	},
	{
		Method:      "GET",
		Path:        "/open/v1/sessions",
		Category:    CategorySession,
		Description: "세션 목록 조회 (Access Token 인증)",
		AuthType:    AuthTypeAccessToken,
		QueryParams: []Param{
			{Name: "page", Type: "int", Required: false, Description: "페이지 번호"},
			{Name: "size", Type: "int", Required: false, Description: "페이지 당 항목 수", Min: "1", Max: "50"},
		},
	},
	{
		Method:      "POST",
		Path:        "/open/v1/sessions/events/subscribe/chat",
		Category:    CategorySession,
		Description: "채팅 이벤트 구독",
		AuthType:    AuthTypeAccessToken,
		QueryParams: []Param{
			{Name: "sessionKey", Type: "string", Required: true, Description: "세션 키"},
		},
	},
	{
		Method:      "POST",
		Path:        "/open/v1/sessions/events/unsubscribe/chat",
		Category:    CategorySession,
		Description: "채팅 이벤트 구독 해제",
		AuthType:    AuthTypeAccessToken,
		QueryParams: []Param{
			{Name: "sessionKey", Type: "string", Required: true, Description: "세션 키"},
		},
	},
	{
		Method:      "POST",
		Path:        "/open/v1/sessions/events/subscribe/donation",
		Category:    CategorySession,
		Description: "후원 이벤트 구독",
		AuthType:    AuthTypeAccessToken,
		QueryParams: []Param{
			{Name: "sessionKey", Type: "string", Required: true, Description: "세션 키"},
		},
	},
	{
		Method:      "POST",
		Path:        "/open/v1/sessions/events/unsubscribe/donation",
		Category:    CategorySession,
		Description: "후원 이벤트 구독 해제",
		AuthType:    AuthTypeAccessToken,
		QueryParams: []Param{
			{Name: "sessionKey", Type: "string", Required: true, Description: "세션 키"},
		},
	},
	{
		Method:      "POST",
		Path:        "/open/v1/sessions/events/subscribe/subscription",
		Category:    CategorySession,
		Description: "구독 이벤트 구독",
		AuthType:    AuthTypeAccessToken,
		QueryParams: []Param{
			{Name: "sessionKey", Type: "string", Required: true, Description: "세션 키"},
		},
	},
	{
		Method:      "POST",
		Path:        "/open/v1/sessions/events/unsubscribe/subscription",
		Category:    CategorySession,
		Description: "구독 이벤트 구독 해제",
		AuthType:    AuthTypeAccessToken,
		QueryParams: []Param{
			{Name: "sessionKey", Type: "string", Required: true, Description: "세션 키"},
		},
	},

	// ── Drops ─────────────────────────────────────────────────────────────────
	{
		Method:      "GET",
		Path:        "/open/v1/drops/reward-claims",
		Category:    CategoryDrops,
		Description: "Drops 보상 청구 목록 조회 (기업 인증 필요)",
		AuthType:    AuthTypeClientCredentials,
		Scope:       "Drops API Scope",
		QueryParams: []Param{
			{Name: "page.from", Type: "string", Required: false, Description: "페이지 시작 커서"},
			{Name: "page.size", Type: "int", Required: false, Description: "페이지 크기"},
			{Name: "claimId", Type: "string", Required: false, Description: "청구 ID 필터"},
			{Name: "channelId", Type: "string", Required: false, Description: "채널 ID 필터"},
			{Name: "campaignId", Type: "string", Required: false, Description: "캠페인 ID 필터"},
			{Name: "categoryId", Type: "string", Required: false, Description: "카테고리 ID 필터"},
			{Name: "fulfillmentState", Type: "string", Required: false, Description: "처리 상태 (CLAIMED | FULFILLED)"},
		},
		Response: []ResponseField{
			{Name: "claimId", Type: "string", Description: "청구 ID"},
			{Name: "campaignId", Type: "string", Description: "캠페인 ID"},
			{Name: "rewardId", Type: "string", Description: "보상 ID"},
			{Name: "categoryId", Type: "string", Description: "카테고리 ID"},
			{Name: "categoryName", Type: "string", Description: "카테고리 이름"},
			{Name: "channelId", Type: "string", Description: "채널 ID"},
			{Name: "fulfillmentState", Type: "string", Description: "처리 상태"},
			{Name: "claimedDate", Type: "string", Description: "청구 날짜"},
			{Name: "updatedDate", Type: "string", Description: "업데이트 날짜"},
		},
	},
	{
		Method:      "PUT",
		Path:        "/open/v1/drops/reward-claims",
		Category:    CategoryDrops,
		Description: "Drops 보상 처리 상태 변경 (기업 인증 필요)",
		AuthType:    AuthTypeClientCredentials,
		Scope:       "Drops API Scope",
		BodyParams: []Param{
			{Name: "claimIds[]", Type: "string[]", Required: true, Description: "처리할 청구 ID 목록"},
			{Name: "fulfillmentState", Type: "string", Required: true, Description: "변경할 처리 상태 (CLAIMED | FULFILLED)"},
		},
	},

	// ── Restriction ───────────────────────────────────────────────────────────
	{
		Method:      "POST",
		Path:        "/open/v1/restrict-channels",
		Category:    CategoryRestriction,
		Description: "채널 활동 제한 추가",
		AuthType:    AuthTypeAccessToken,
		Scope:       "활동제한 쓰기",
		BodyParams: []Param{
			{Name: "targetChannelId", Type: "string", Required: true, Description: "제한할 채널 ID"},
		},
	},
	{
		Method:      "DELETE",
		Path:        "/open/v1/restrict-channels",
		Category:    CategoryRestriction,
		Description: "채널 활동 제한 해제",
		AuthType:    AuthTypeAccessToken,
		Scope:       "활동제한 쓰기",
		BodyParams: []Param{
			{Name: "targetChannelId", Type: "string", Required: true, Description: "제한 해제할 채널 ID"},
		},
	},
	{
		Method:      "GET",
		Path:        "/open/v1/restrict-channels",
		Category:    CategoryRestriction,
		Description: "채널 활동 제한 목록 조회 (커서 페이지네이션)",
		AuthType:    AuthTypeAccessToken,
		Scope:       "활동제한 조회",
		QueryParams: []Param{
			{Name: "size", Type: "int", Required: false, Description: "반환 개수", Max: "30"},
			{Name: "next", Type: "string", Required: false, Description: "다음 페이지 커서"},
		},
		Response: []ResponseField{
			{Name: "restrictedChannelId", Type: "string", Description: "제한된 채널 ID"},
			{Name: "restrictedChannelName", Type: "string", Description: "제한된 채널 이름"},
			{Name: "createdDate", Type: "string", Description: "제한 날짜"},
			{Name: "releaseDate", Type: "string", Description: "제한 해제 날짜"},
		},
	},
	{
		Method:      "POST",
		Path:        "/open/v1/temporary-restrict-channels",
		Category:    CategoryRestriction,
		Description: "채널 임시 활동 제한 추가",
		AuthType:    AuthTypeAccessToken,
		Scope:       "활동제한 쓰기",
		BodyParams: []Param{
			{Name: "targetChannelId", Type: "string", Required: true, Description: "임시 제한할 채널 ID"},
			{Name: "chatChannelId", Type: "string", Required: true, Description: "채팅 채널 ID"},
		},
	},
	{
		Method:      "DELETE",
		Path:        "/open/v1/temporary-restrict-channels",
		Category:    CategoryRestriction,
		Description: "채널 임시 활동 제한 해제",
		AuthType:    AuthTypeAccessToken,
		Scope:       "활동제한 쓰기",
		BodyParams: []Param{
			{Name: "targetChannelId", Type: "string", Required: true, Description: "임시 제한 해제할 채널 ID"},
			{Name: "chatChannelId", Type: "string", Required: true, Description: "채팅 채널 ID"},
		},
	},
}

var endpointIndex map[string]Endpoint

func init() {
	endpointIndex = make(map[string]Endpoint, len(AllEndpoints))
	for _, e := range AllEndpoints {
		endpointIndex[e.Key()] = e
	}
}

// FindEndpoint returns the endpoint matching "METHOD /path", or false if not found.
func FindEndpoint(key string) (Endpoint, bool) {
	e, ok := endpointIndex[key]
	return e, ok
}

// EndpointsByCategory returns all endpoints in the given category.
func EndpointsByCategory(cat Category) []Endpoint {
	var result []Endpoint
	for _, e := range AllEndpoints {
		if e.Category == cat {
			result = append(result, e)
		}
	}
	return result
}

// Categories returns all distinct categories in order.
var Categories = []Category{
	CategoryAuth,
	CategoryUser,
	CategoryChannel,
	CategoryCategory,
	CategoryLive,
	CategoryChat,
	CategorySession,
	CategoryDrops,
	CategoryRestriction,
}
