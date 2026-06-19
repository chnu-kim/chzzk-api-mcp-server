package chzzk

import "testing"

func mustFindEndpoint(t *testing.T, key string) Endpoint {
	t.Helper()
	ep, ok := FindEndpoint(key)
	if !ok {
		t.Fatalf("endpoint not found: %s", key)
	}
	return ep
}

func findParam(params []Param, name string) (Param, bool) {
	for _, p := range params {
		if p.Name == name {
			return p, true
		}
	}
	return Param{}, false
}

func findResponseField(fields []ResponseField, name string) (ResponseField, bool) {
	for _, f := range fields {
		if f.Name == name {
			return f, true
		}
	}
	return ResponseField{}, false
}

func TestAPIReference_UsersMe_HasScope(t *testing.T) {
	ep := mustFindEndpoint(t, "GET /open/v1/users/me")
	if ep.Scope == "" {
		t.Error("Scope should be '유저 정보 조회'")
	}
}

func TestAPIReference_StreamingRoles_HasResponseFields(t *testing.T) {
	ep := mustFindEndpoint(t, "GET /open/v1/channels/streaming-roles")
	for _, field := range []string{"managerChannelId", "managerChannelName", "userRole", "createdDate"} {
		if _, ok := findResponseField(ep.Response, field); !ok {
			t.Errorf("response field %q missing", field)
		}
	}
}

func TestAPIReference_Lives_HasLiveCategoryValue(t *testing.T) {
	ep := mustFindEndpoint(t, "GET /open/v1/lives")
	if _, ok := findResponseField(ep.Response, "liveCategoryValue"); !ok {
		t.Error("response field 'liveCategoryValue' missing")
	}
}

func TestAPIReference_Lives_SizeHasDefault(t *testing.T) {
	ep := mustFindEndpoint(t, "GET /open/v1/lives")
	p, ok := findParam(ep.QueryParams, "size")
	if !ok {
		t.Fatal("query param 'size' not found")
	}
	if p.Default == "" {
		t.Error("'size' param should have default value '20'")
	}
}

func TestAPIReference_LivesSetting_HasResponseFields(t *testing.T) {
	ep := mustFindEndpoint(t, "GET /open/v1/lives/setting")
	if len(ep.Response) == 0 {
		t.Error("should have response fields")
	}
}

func TestAPIReference_ChatsSettings_HasScope(t *testing.T) {
	for _, key := range []string{"GET /open/v1/chats/settings", "PUT /open/v1/chats/settings"} {
		t.Run(key, func(t *testing.T) {
			ep := mustFindEndpoint(t, key)
			if ep.Scope == "" {
				t.Errorf("Scope is empty")
			}
		})
	}
}

func assertResponseFields(t *testing.T, ep Endpoint, want []string) {
	t.Helper()
	wantSet := make(map[string]bool, len(want))
	for _, n := range want {
		wantSet[n] = true
	}
	gotSet := make(map[string]bool, len(ep.Response))
	for _, f := range ep.Response {
		gotSet[f.Name] = true
	}
	for _, n := range want {
		if !gotSet[n] {
			t.Errorf("response field %q missing", n)
		}
	}
	for _, f := range ep.Response {
		if !wantSet[f.Name] {
			t.Errorf("unexpected response field %q", f.Name)
		}
	}
}

func TestAPIReference_ResponseFieldsComplete(t *testing.T) {
	cases := []struct {
		endpoint string
		fields   []string
	}{
		{
			"GET /open/v1/users/me",
			[]string{"channelId", "channelName"},
		},
		{
			"GET /open/v1/channels",
			[]string{"channelId", "channelName", "channelImageUrl", "followerCount", "verifiedMark"},
		},
		{
			"GET /open/v1/channels/streaming-roles",
			[]string{"managerChannelId", "managerChannelName", "userRole", "createdDate"},
		},
		{
			"GET /open/v1/channels/followers",
			[]string{"channelId", "channelName", "createdDate"},
		},
		{
			"GET /open/v1/channels/subscribers",
			[]string{"channelId", "channelName", "month", "tierNo", "createdDate"},
		},
		{
			"GET /open/v1/categories/search",
			[]string{"categoryType", "categoryId", "categoryValue", "posterImageUrl"},
		},
		{
			"GET /open/v1/lives",
			[]string{
				"liveId", "liveTitle", "liveThumbnailImageUrl", "concurrentUserCount",
				"openDate", "adult", "tags", "categoryType", "liveCategory", "liveCategoryValue",
				"channelId", "channelName", "channelImageUrl",
			},
		},
		{
			"GET /open/v1/lives/setting",
			[]string{"defaultLiveTitle", "tags"},
		},
		{
			"GET /open/v1/chats/settings",
			[]string{
				"chatAvailableCondition", "chatAvailableGroup", "minFollowerMinute",
				"allowSubscriberInFollowerMode", "chatSlowModeSec", "chatEmojiMode",
			},
		},
		{
			"GET /open/v1/sessions/auth/client",
			[]string{"webSocketUrl"},
		},
		{
			"GET /open/v1/sessions/auth",
			[]string{"webSocketUrl"},
		},
		{
			"GET /open/v1/sessions/client",
			[]string{"sessionKey", "connectedDate", "disconnectedDate", "subscribedEvents"},
		},
		{
			"GET /open/v1/drops/reward-claims",
			[]string{
				"claimId", "campaignId", "rewardId", "categoryId", "categoryName",
				"channelId", "fulfillmentState", "claimedDate", "updatedDate",
			},
		},
		{
			"GET /open/v1/restrict-channels",
			[]string{"restrictedChannelId", "restrictedChannelName", "createdDate", "releaseDate"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.endpoint, func(t *testing.T) {
			ep := mustFindEndpoint(t, tc.endpoint)
			assertResponseFields(t, ep, tc.fields)
		})
	}
}
