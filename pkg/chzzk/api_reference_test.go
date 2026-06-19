package chzzk

import "testing"

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
	ep, ok := FindEndpoint("GET /open/v1/users/me")
	if !ok {
		t.Fatal("endpoint not found")
	}
	if ep.Scope == "" {
		t.Error("GET /open/v1/users/me: Scope should be '유저 정보 조회'")
	}
}

func TestAPIReference_StreamingRoles_HasResponseFields(t *testing.T) {
	ep, ok := FindEndpoint("GET /open/v1/channels/streaming-roles")
	if !ok {
		t.Fatal("endpoint not found")
	}
	required := []string{"managerChannelId", "managerChannelName", "userRole", "createdDate"}
	for _, field := range required {
		if _, ok := findResponseField(ep.Response, field); !ok {
			t.Errorf("GET /open/v1/channels/streaming-roles: response field %q missing", field)
		}
	}
}

func TestAPIReference_Lives_HasLiveCategoryValue(t *testing.T) {
	ep, ok := FindEndpoint("GET /open/v1/lives")
	if !ok {
		t.Fatal("endpoint not found")
	}
	if _, ok := findResponseField(ep.Response, "liveCategoryValue"); !ok {
		t.Error("GET /open/v1/lives: response field 'liveCategoryValue' missing")
	}
}

func TestAPIReference_Lives_SizeHasDefault(t *testing.T) {
	ep, ok := FindEndpoint("GET /open/v1/lives")
	if !ok {
		t.Fatal("endpoint not found")
	}
	p, ok := findParam(ep.QueryParams, "size")
	if !ok {
		t.Fatal("GET /open/v1/lives: query param 'size' not found")
	}
	if p.Default == "" {
		t.Error("GET /open/v1/lives: 'size' param should have default value '20'")
	}
}

func TestAPIReference_LivesSetting_HasResponseFields(t *testing.T) {
	ep, ok := FindEndpoint("GET /open/v1/lives/setting")
	if !ok {
		t.Fatal("endpoint not found")
	}
	if len(ep.Response) == 0 {
		t.Error("GET /open/v1/lives/setting: should have response fields")
	}
}

func TestAPIReference_ChatsSettings_HasScope(t *testing.T) {
	get, ok := FindEndpoint("GET /open/v1/chats/settings")
	if !ok {
		t.Fatal("GET endpoint not found")
	}
	if get.Scope == "" {
		t.Error("GET /open/v1/chats/settings: Scope should be '채팅 설정 조회'")
	}

	put, ok := FindEndpoint("PUT /open/v1/chats/settings")
	if !ok {
		t.Fatal("PUT endpoint not found")
	}
	if put.Scope == "" {
		t.Error("PUT /open/v1/chats/settings: Scope should be '채팅 설정 변경'")
	}
}
