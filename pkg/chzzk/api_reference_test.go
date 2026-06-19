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
