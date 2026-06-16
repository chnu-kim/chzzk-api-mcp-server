package chzzk

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ─── chzzk_generate_auth_code ─────────────────────────────────────────────────

type GenerateAuthCodeInput struct {
	Language string `json:"language" jsonschema:"코드 생성 언어. go 또는 typescript"`
}

func handleGenerateAuthCode(_ context.Context, _ *mcp.CallToolRequest, input GenerateAuthCodeInput) (*mcp.CallToolResult, any, error) {
	switch strings.ToLower(input.Language) {
	case "go":
		return textResult(authCodeGo()), nil, nil
	case "typescript", "ts":
		return textResult(authCodeTypeScript()), nil, nil
	default:
		return errorResult(fmt.Sprintf("지원하지 않는 언어: %q. 지원 언어: go, typescript", input.Language)), nil, nil
	}
}

// ─── chzzk_generate_api_client ────────────────────────────────────────────────

type GenerateAPIClientInput struct {
	Language  string   `json:"language" jsonschema:"코드 생성 언어. go 또는 typescript"`
	Endpoints []string `json:"endpoints" jsonschema:"클라이언트를 생성할 엔드포인트 목록. 'METHOD /path' 형식. 예: ['GET /open/v1/lives', 'POST /open/v1/chats/send']"`
}

func handleGenerateAPIClient(_ context.Context, _ *mcp.CallToolRequest, input GenerateAPIClientInput) (*mcp.CallToolResult, any, error) {
	if len(input.Endpoints) == 0 {
		return errorResult("endpoints는 최소 한 개 이상이어야 합니다"), nil, nil
	}

	var eps []Endpoint
	var notFound []string
	for _, key := range input.Endpoints {
		ep, ok := FindEndpoint(strings.TrimSpace(key))
		if !ok {
			notFound = append(notFound, key)
		} else {
			eps = append(eps, ep)
		}
	}
	if len(notFound) > 0 {
		return errorResult(fmt.Sprintf("찾을 수 없는 엔드포인트: %s\nchzzk_list_apis 도구로 올바른 엔드포인트를 확인하세요.", strings.Join(notFound, ", "))), nil, nil
	}

	switch strings.ToLower(input.Language) {
	case "go":
		return textResult(apiClientGo(eps)), nil, nil
	case "typescript", "ts":
		return textResult(apiClientTypeScript(eps)), nil, nil
	default:
		return errorResult(fmt.Sprintf("지원하지 않는 언어: %q. 지원 언어: go, typescript", input.Language)), nil, nil
	}
}

// ─── chzzk_scaffold_project ───────────────────────────────────────────────────

type ScaffoldProjectInput struct {
	Language    string   `json:"language" jsonschema:"프로젝트 언어. go 또는 typescript"`
	ProjectName string   `json:"project_name" jsonschema:"프로젝트 이름 (예: my-chzzk-bot)"`
	Features    []string `json:"features" jsonschema:"포함할 기능 목록. auth, live, chat, channel, session 중 선택. 예: ['auth', 'live', 'chat']"`
}

func handleScaffoldProject(_ context.Context, _ *mcp.CallToolRequest, input ScaffoldProjectInput) (*mcp.CallToolResult, any, error) {
	if input.ProjectName == "" {
		input.ProjectName = "chzzk-app"
	}
	if len(input.Features) == 0 {
		input.Features = []string{"auth", "live", "chat"}
	}

	featureSet := make(map[string]bool)
	for _, f := range input.Features {
		featureSet[strings.ToLower(f)] = true
	}

	switch strings.ToLower(input.Language) {
	case "go":
		return textResult(scaffoldGo(input.ProjectName, featureSet)), nil, nil
	case "typescript", "ts":
		return textResult(scaffoldTypeScript(input.ProjectName, featureSet)), nil, nil
	default:
		return errorResult(fmt.Sprintf("지원하지 않는 언어: %q. 지원 언어: go, typescript", input.Language)), nil, nil
	}
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}

func errorResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}

// ─── Go auth code ─────────────────────────────────────────────────────────────

func authCodeGo() string {
	return `// Package auth provides Chzzk OAuth2 authentication helpers.
// Required environment variables:
//   CHZZK_CLIENT_ID     - 애플리케이션 클라이언트 ID
//   CHZZK_CLIENT_SECRET - 애플리케이션 클라이언트 시크릿
//   CHZZK_REDIRECT_URI  - 인가 코드를 전달받을 리다이렉트 URI
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	baseURL   = "https://openapi.chzzk.naver.com"
	authURL   = "https://chzzk.naver.com/account-interlock"
	tokenPath = "/auth/v1/token"
	revokePath = "/auth/v1/token/revoke"
)

// TokenResponse holds the tokens returned by the Chzzk API.
type TokenResponse struct {
	AccessToken  string ` + "`" + `json:"accessToken"` + "`" + `
	RefreshToken string ` + "`" + `json:"refreshToken"` + "`" + `
	TokenType    string ` + "`" + `json:"tokenType"` + "`" + `
	ExpiresIn    int    ` + "`" + `json:"expiresIn"` + "`" + `
	Scope        string ` + "`" + `json:"scope"` + "`" + `
}

type apiResponse[T any] struct {
	Code    int     ` + "`" + `json:"code"` + "`" + `
	Message *string ` + "`" + `json:"message"` + "`" + `
	Content T       ` + "`" + `json:"content"` + "`" + `
}

// Config holds auth configuration loaded from environment variables.
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// ConfigFromEnv loads auth config from environment variables.
func ConfigFromEnv() Config {
	return Config{
		ClientID:     os.Getenv("CHZZK_CLIENT_ID"),
		ClientSecret: os.Getenv("CHZZK_CLIENT_SECRET"),
		RedirectURI:  os.Getenv("CHZZK_REDIRECT_URI"),
	}
}

// AuthorizationURL builds the Chzzk account-interlock URL to redirect the user to.
// state should be a random string to prevent CSRF attacks.
func (c Config) AuthorizationURL(state string) string {
	params := url.Values{
		"clientId":    {c.ClientID},
		"redirectUri": {c.RedirectURI},
		"state":       {state},
	}
	return authURL + "?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for tokens.
func (c Config) ExchangeCode(ctx context.Context, code, state string) (*TokenResponse, error) {
	return postToken(ctx, url.Values{
		"grantType":    {"authorization_code"},
		"clientId":     {c.ClientID},
		"clientSecret": {c.ClientSecret},
		"code":         {code},
		"state":        {state},
	})
}

// RefreshToken uses a refresh token to obtain a new access token.
func (c Config) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	return postToken(ctx, url.Values{
		"grantType":    {"refresh_token"},
		"clientId":     {c.ClientID},
		"clientSecret": {c.ClientSecret},
		"refreshToken": {refreshToken},
	})
}

// RevokeToken invalidates the given token.
func (c Config) RevokeToken(ctx context.Context, token, tokenTypeHint string) error {
	data := url.Values{
		"clientId":     {c.ClientID},
		"clientSecret": {c.ClientSecret},
		"token":        {token},
	}
	if tokenTypeHint != "" {
		data.Set("tokenTypeHint", tokenTypeHint)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+revokePath,
		strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("revoke token: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func postToken(ctx context.Context, data url.Values) (*TokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+tokenPath,
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result apiResponse[TokenResponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}
	if result.Code != http.StatusOK {
		msg := "unknown error"
		if result.Message != nil {
			msg = *result.Message
		}
		return nil, fmt.Errorf("token error %d: %s", result.Code, msg)
	}
	return &result.Content, nil
}

// CallbackServer is a minimal HTTP server that captures the OAuth2 callback.
// Usage: start it, redirect the user to AuthorizationURL, then wait on Done.
type CallbackServer struct {
	Addr   string
	Done   chan CallbackResult
	server *http.Server
}

// CallbackResult holds the code/state or an error from the OAuth2 callback.
type CallbackResult struct {
	Code  string
	State string
	Error string
}

// NewCallbackServer creates a callback server on the given address (e.g. ":8080").
func NewCallbackServer(addr string) *CallbackServer {
	cs := &CallbackServer{Addr: addr, Done: make(chan CallbackResult, 1)}
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", cs.handleCallback)
	cs.server = &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	return cs
}

// Start begins listening in a goroutine.
func (cs *CallbackServer) Start() error {
	go func() { _ = cs.server.ListenAndServe() }()
	return nil
}

// Shutdown gracefully shuts down the server.
func (cs *CallbackServer) Shutdown(ctx context.Context) { _ = cs.server.Shutdown(ctx) }

func (cs *CallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if errMsg := q.Get("error"); errMsg != "" {
		cs.Done <- CallbackResult{Error: errMsg}
		fmt.Fprintln(w, "인증 실패: "+errMsg)
		return
	}
	cs.Done <- CallbackResult{Code: q.Get("code"), State: q.Get("state")}
	fmt.Fprintln(w, "인증 완료. 브라우저를 닫아도 됩니다.")
}
`
}

// ─── TypeScript auth code ─────────────────────────────────────────────────────

func authCodeTypeScript() string {
	return `// Chzzk OAuth2 authentication helpers
// Required environment variables:
//   CHZZK_CLIENT_ID     - 애플리케이션 클라이언트 ID
//   CHZZK_CLIENT_SECRET - 애플리케이션 클라이언트 시크릿
//   CHZZK_REDIRECT_URI  - 인가 코드를 전달받을 리다이렉트 URI

const BASE_URL = "https://openapi.chzzk.naver.com";
const AUTH_URL = "https://chzzk.naver.com/account-interlock";

export interface ChzzkAuthConfig {
  clientId: string;
  clientSecret: string;
  redirectUri: string;
}

export interface TokenResponse {
  accessToken: string;
  refreshToken: string;
  tokenType: string;
  expiresIn: number;
  scope?: string;
}

interface ApiResponse<T> {
  code: number;
  message: string | null;
  content: T;
}

export function configFromEnv(): ChzzkAuthConfig {
  const clientId = process.env.CHZZK_CLIENT_ID;
  const clientSecret = process.env.CHZZK_CLIENT_SECRET;
  const redirectUri = process.env.CHZZK_REDIRECT_URI;

  if (!clientId || !clientSecret || !redirectUri) {
    throw new Error(
      "Missing required environment variables: CHZZK_CLIENT_ID, CHZZK_CLIENT_SECRET, CHZZK_REDIRECT_URI"
    );
  }
  return { clientId, clientSecret, redirectUri };
}

/** Builds the Chzzk account-interlock URL to redirect the user to. */
export function buildAuthorizationURL(config: ChzzkAuthConfig, state: string): string {
  const params = new URLSearchParams({
    clientId: config.clientId,
    redirectUri: config.redirectUri,
    state,
  });
  return ` + "`" + `${AUTH_URL}?${params}` + "`" + `;
}

/** Exchanges an authorization code for tokens. */
export async function exchangeCode(
  config: ChzzkAuthConfig,
  code: string,
  state: string
): Promise<TokenResponse> {
  return postToken(new URLSearchParams({
    grantType: "authorization_code",
    clientId: config.clientId,
    clientSecret: config.clientSecret,
    code,
    state,
  }));
}

/** Uses a refresh token to obtain a new access token. */
export async function refreshAccessToken(
  config: ChzzkAuthConfig,
  refreshToken: string
): Promise<TokenResponse> {
  return postToken(new URLSearchParams({
    grantType: "refresh_token",
    clientId: config.clientId,
    clientSecret: config.clientSecret,
    refreshToken,
  }));
}

/** Invalidates the given token. */
export async function revokeToken(
  config: ChzzkAuthConfig,
  token: string,
  tokenTypeHint?: "access_token" | "refresh_token"
): Promise<void> {
  const params = new URLSearchParams({
    clientId: config.clientId,
    clientSecret: config.clientSecret,
    token,
  });
  if (tokenTypeHint) params.set("tokenTypeHint", tokenTypeHint);

  const res = await fetch(` + "`" + `${BASE_URL}/auth/v1/token/revoke` + "`" + `, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: params,
  });
  if (!res.ok) throw new Error(` + "`" + `Revoke token failed: ${res.status}` + "`" + `);
}

async function postToken(params: URLSearchParams): Promise<TokenResponse> {
  const res = await fetch(` + "`" + `${BASE_URL}/auth/v1/token` + "`" + `, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: params,
  });
  const data: ApiResponse<TokenResponse> = await res.json();
  if (data.code !== 200) {
    throw new Error(` + "`" + `Token error ${data.code}: ${data.message}` + "`" + `);
  }
  return data.content;
}

// ── Token persistence example (Node.js) ─────────────────────────────────────
// import fs from "fs";
//
// const TOKEN_FILE = ".chzzk-token.json";
//
// export function saveTokens(tokens: TokenResponse): void {
//   fs.writeFileSync(TOKEN_FILE, JSON.stringify(tokens, null, 2));
// }
//
// export function loadTokens(): TokenResponse | null {
//   try {
//     return JSON.parse(fs.readFileSync(TOKEN_FILE, "utf-8"));
//   } catch {
//     return null;
//   }
// }
`
}

// ─── Go API client ────────────────────────────────────────────────────────────

func apiClientGo(eps []Endpoint) string {
	var sb strings.Builder

	sb.WriteString(`// Package chzzk provides a typed HTTP client for the Chzzk Open API.
package chzzk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const baseURL = "https://openapi.chzzk.naver.com"

// Client is a Chzzk API client.
type Client struct {
	http         *http.Client
	accessToken  string
	clientID     string
	clientSecret string
}

// NewClient creates a new Chzzk API client.
// accessToken is optional for endpoints that only require client credentials.
func NewClient(clientID, clientSecret, accessToken string) *Client {
	return &Client{
		http:         &http.Client{},
		accessToken:  accessToken,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

type apiResponse[T any] struct {
	Code    int     ` + "`" + `json:"code"` + "`" + `
	Message *string ` + "`" + `json:"message"` + "`" + `
	Content T       ` + "`" + `json:"content"` + "`" + `
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body string, useAccessToken bool) (*http.Response, error) {
	reqURL := baseURL + path
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}
	var bodyReader *strings.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	var req *http.Request
	var err error
	if bodyReader != nil {
		req, err = http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, reqURL, nil)
	}
	if err != nil {
		return nil, err
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if useAccessToken {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	return c.http.Do(req)
}

`)

	for _, ep := range eps {
		sb.WriteString(goMethodForEndpoint(ep))
		sb.WriteString("\n")
	}

	return sb.String()
}

func goMethodForEndpoint(ep Endpoint) string {
	var sb strings.Builder
	name := goMethodName(ep)
	useToken := ep.AuthType == AuthTypeAccessToken

	// struct definitions
	if len(ep.Response) > 0 {
		sb.WriteString(fmt.Sprintf("// %s is the response from %s %s.\n", name+"Response", ep.Method, ep.Path))
		sb.WriteString(fmt.Sprintf("type %sResponse struct {\n", name))
		for _, r := range ep.Response {
			fieldName := goFieldName(r.Name)
			fieldType := goType(r.Type)
			sb.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"` // %s\n", fieldName, fieldType, r.Name, r.Description))
		}
		sb.WriteString("}\n\n")
	}

	// function signature comment
	sb.WriteString(fmt.Sprintf("// %s calls %s %s.\n", name, ep.Method, ep.Path))
	if ep.Description != "" {
		sb.WriteString(fmt.Sprintf("// %s\n", ep.Description))
	}

	// params
	var params []string
	for _, p := range ep.QueryParams {
		if p.Required {
			params = append(params, fmt.Sprintf("%s %s", goParamName(p.Name), goType(p.Type)))
		}
	}
	for _, p := range ep.BodyParams {
		if p.Required {
			params = append(params, fmt.Sprintf("%s %s", goParamName(p.Name), goType(p.Type)))
		}
	}
	// optional query params as separate args
	var optParams []string
	for _, p := range ep.QueryParams {
		if !p.Required {
			optParams = append(optParams, fmt.Sprintf("%s %s", goParamName(p.Name), goOptType(p.Type)))
		}
	}
	params = append(params, optParams...)

	retType := "error"
	if len(ep.Response) > 0 {
		retType = fmt.Sprintf("(*%sResponse, error)", name)
	}

	paramStr := "ctx context.Context"
	if len(params) > 0 {
		paramStr += ", " + strings.Join(params, ", ")
	}

	sb.WriteString(fmt.Sprintf("func (c *Client) %s(%s) %s {\n", name, paramStr, retType))

	// build query
	hasQuery := len(ep.QueryParams) > 0
	if hasQuery {
		sb.WriteString("\tq := url.Values{}\n")
		for _, p := range ep.QueryParams {
			pn := goParamName(p.Name)
			cleanName := strings.TrimSuffix(p.Name, "[]")
			if p.Required {
				if p.Type == "int" {
					sb.WriteString(fmt.Sprintf("\tq.Set(%q, strconv.Itoa(%s))\n", cleanName, pn))
				} else {
					sb.WriteString(fmt.Sprintf("\tq.Set(%q, %s)\n", cleanName, pn))
				}
			} else {
				zero := goZero(p.Type)
				if p.Type == "int" {
					sb.WriteString(fmt.Sprintf("\tif %s != %s {\n\t\tq.Set(%q, strconv.Itoa(%s))\n\t}\n", pn, zero, cleanName, pn))
				} else {
					sb.WriteString(fmt.Sprintf("\tif %s != %s {\n\t\tq.Set(%q, %s)\n\t}\n", pn, zero, cleanName, pn))
				}
			}
		}
	}

	queryArg := "nil"
	if hasQuery {
		queryArg = "q"
	}

	if len(ep.Response) == 0 {
		sb.WriteString(fmt.Sprintf("\tresp, err := c.do(ctx, %q, %q, %s, \"\", %v)\n", ep.Method, ep.Path, queryArg, useToken))
		sb.WriteString("\tif err != nil {\n\t\treturn err\n\t}\n")
		sb.WriteString("\tdefer resp.Body.Close()\n")
		sb.WriteString("\tif resp.StatusCode != http.StatusOK {\n")
		sb.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"%s: unexpected status %%d\", resp.StatusCode)\n", name))
		sb.WriteString("\t}\n")
		sb.WriteString("\treturn nil\n")
	} else {
		sb.WriteString(fmt.Sprintf("\tresp, err := c.do(ctx, %q, %q, %s, \"\", %v)\n", ep.Method, ep.Path, queryArg, useToken))
		sb.WriteString("\tif err != nil {\n\t\treturn nil, err\n\t}\n")
		sb.WriteString("\tdefer resp.Body.Close()\n\n")
		sb.WriteString(fmt.Sprintf("\tvar result apiResponse[%sResponse]\n", name))
		sb.WriteString("\tif err := json.NewDecoder(resp.Body).Decode(&result); err != nil {\n\t\treturn nil, err\n\t}\n")
		sb.WriteString("\tif result.Code != http.StatusOK {\n")
		sb.WriteString("\t\tmsg := \"unknown error\"\n")
		sb.WriteString("\t\tif result.Message != nil { msg = *result.Message }\n")
		sb.WriteString(fmt.Sprintf("\t\treturn nil, fmt.Errorf(\"%s: API error %%d: %%s\", result.Code, msg)\n", name))
		sb.WriteString("\t}\n")
		sb.WriteString("\treturn &result.Content, nil\n")
	}
	sb.WriteString("}\n")

	return sb.String()
}

// ─── TypeScript API client ────────────────────────────────────────────────────

func apiClientTypeScript(eps []Endpoint) string {
	var sb strings.Builder

	sb.WriteString(`// Chzzk Open API typed client
// Usage: const client = new ChzzkClient({ clientId, clientSecret, accessToken });

const BASE_URL = "https://openapi.chzzk.naver.com";

interface ApiResponse<T> {
  code: number;
  message: string | null;
  content: T;
}

export interface ChzzkClientConfig {
  clientId: string;
  clientSecret: string;
  accessToken?: string;
}

export class ChzzkClient {
  constructor(private config: ChzzkClientConfig) {}

  private async request<T>(
    method: string,
    path: string,
    options: { query?: Record<string, string | number>; body?: unknown; auth?: boolean } = {}
  ): Promise<T> {
    let url = ` + "`" + `${BASE_URL}${path}` + "`" + `;
    if (options.query) {
      const params = new URLSearchParams(
        Object.entries(options.query)
          .filter(([, v]) => v !== undefined && v !== "" && v !== 0)
          .map(([k, v]) => [k, String(v)])
      );
      if (params.size > 0) url += "?" + params;
    }
    const headers: Record<string, string> = { "Content-Type": "application/json" };
    if (options.auth && this.config.accessToken) {
      headers["Authorization"] = ` + "`" + `Bearer ${this.config.accessToken}` + "`" + `;
    }
    const res = await fetch(url, {
      method,
      headers,
      body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
    });
    const data: ApiResponse<T> = await res.json();
    if (data.code !== 200) {
      throw new Error(` + "`" + `Chzzk API error ${data.code}: ${data.message}` + "`" + `);
    }
    return data.content;
  }

`)

	for _, ep := range eps {
		sb.WriteString(tsMethodForEndpoint(ep))
		sb.WriteString("\n")
	}

	sb.WriteString("}\n")
	return sb.String()
}

func tsMethodForEndpoint(ep Endpoint) string {
	var sb strings.Builder
	name := tsMethodName(ep)
	useToken := ep.AuthType == AuthTypeAccessToken

	// interfaces
	if len(ep.Response) > 0 {
		sb.WriteString(fmt.Sprintf("  // %s\n", ep.Description))
		iName := tsInterfaceName(ep) + "Response"
		sb.WriteString(fmt.Sprintf("  // Response type: %s\n", iName))
	}

	sb.WriteString(fmt.Sprintf("  /** %s\n   * %s %s\n   */\n", ep.Description, ep.Method, ep.Path))
	sb.WriteString(fmt.Sprintf("  async %s(", name))

	var params []string
	for _, p := range ep.QueryParams {
		if p.Required {
			params = append(params, fmt.Sprintf("%s: %s", tsParamName(p.Name), tsType(p.Type)))
		}
	}
	for _, p := range ep.BodyParams {
		if p.Required {
			params = append(params, fmt.Sprintf("%s: %s", tsParamName(p.Name), tsType(p.Type)))
		}
	}
	var optParams []string
	for _, p := range ep.QueryParams {
		if !p.Required {
			optParams = append(optParams, fmt.Sprintf("%s?: %s", tsParamName(p.Name), tsType(p.Type)))
		}
	}
	params = append(params, optParams...)

	sb.WriteString(strings.Join(params, ", "))

	retType := "void"
	if len(ep.Response) > 0 {
		retType = tsInterfaceName(ep) + "Response"
	}
	sb.WriteString(fmt.Sprintf("): Promise<%s> {\n", retType))

	// build query obj
	if len(ep.QueryParams) > 0 {
		sb.WriteString("    const query: Record<string, string | number> = {};\n")
		for _, p := range ep.QueryParams {
			pn := tsParamName(p.Name)
			cleanName := strings.TrimSuffix(p.Name, "[]")
			if p.Required {
				sb.WriteString(fmt.Sprintf("    query[%q] = %s;\n", cleanName, pn))
			} else {
				sb.WriteString(fmt.Sprintf("    if (%s !== undefined) query[%q] = %s;\n", pn, cleanName, pn))
			}
		}
	}

	var bodyStr string
	if len(ep.BodyParams) > 0 {
		var bodyFields []string
		for _, p := range ep.BodyParams {
			pn := tsParamName(p.Name)
			bodyFields = append(bodyFields, fmt.Sprintf("%s: %s", tsParamName(p.Name), pn))
		}
		bodyStr = "{ " + strings.Join(bodyFields, ", ") + " }"
	}

	sb.WriteString(fmt.Sprintf("    return this.request<%s>(%q, %q, {\n", retType, ep.Method, ep.Path))
	if len(ep.QueryParams) > 0 {
		sb.WriteString("      query,\n")
	}
	if bodyStr != "" {
		sb.WriteString(fmt.Sprintf("      body: %s,\n", bodyStr))
	}
	if useToken {
		sb.WriteString("      auth: true,\n")
	}
	sb.WriteString("    });\n  }\n")

	return sb.String()
}

// ─── Go scaffold ──────────────────────────────────────────────────────────────

func scaffoldGo(projectName string, features map[string]bool) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`# Chzzk Go 프로젝트 스캐폴드: %s

## 디렉토리 구조

`+"```"+`
%s/
├── main.go
├── go.mod
├── internal/
`, projectName, projectName))

	if features["auth"] {
		sb.WriteString("│   ├── auth/\n│   │   └── auth.go\n")
	}
	if features["live"] {
		sb.WriteString("│   ├── live/\n│   │   └── live.go\n")
	}
	if features["chat"] {
		sb.WriteString("│   ├── chat/\n│   │   └── chat.go\n")
	}
	if features["channel"] {
		sb.WriteString("│   ├── channel/\n│   │   └── channel.go\n")
	}
	if features["session"] {
		sb.WriteString("│   ├── session/\n│   │   └── session.go\n")
	}
	sb.WriteString("│   └── client/\n│       └── client.go\n└── config.go\n```\n\n")

	sb.WriteString("## go.mod\n\n```\nmodule github.com/yourname/" + projectName + "\n\ngo 1.23\n```\n\n")
	sb.WriteString("## main.go\n\n```go\npackage main\n\nimport (\n\t\"context\"\n\t\"log\"\n\t\"os\"\n\n\t\"github.com/yourname/" + projectName + "/internal/client\"\n)\n\nfunc main() {\n\tc := client.NewFromEnv()\n\tctx := context.Background()\n\n\t_ = c\n\t_ = ctx\n\tlog.Println(\"Chzzk client initialized\")\n\n\t// TODO: 비즈니스 로직 작성\n}\n```\n\n")

	sb.WriteString("## config.go\n\n```go\npackage main\n\nimport \"os\"\n\ntype Config struct {\n\tClientID     string\n\tClientSecret string\n\tAccessToken  string\n\tRefreshToken string\n}\n\nfunc ConfigFromEnv() Config {\n\treturn Config{\n\t\tClientID:     os.Getenv(\"CHZZK_CLIENT_ID\"),\n\t\tClientSecret: os.Getenv(\"CHZZK_CLIENT_SECRET\"),\n\t\tAccessToken:  os.Getenv(\"CHZZK_ACCESS_TOKEN\"),\n\t\tRefreshToken: os.Getenv(\"CHZZK_REFRESH_TOKEN\"),\n\t}\n}\n```\n\n")

	sb.WriteString("## internal/client/client.go\n\n```go\npackage client\n\nimport (\n\t\"context\"\n\t\"encoding/json\"\n\t\"fmt\"\n\t\"net/http\"\n\t\"os\"\n)\n\nconst baseURL = \"https://openapi.chzzk.naver.com\"\n\ntype Client struct {\n\thttp         *http.Client\n\tClientID     string\n\tClientSecret string\n\tAccessToken  string\n}\n\nfunc NewFromEnv() *Client {\n\treturn &Client{\n\t\thttp:         &http.Client{},\n\t\tClientID:     os.Getenv(\"CHZZK_CLIENT_ID\"),\n\t\tClientSecret: os.Getenv(\"CHZZK_CLIENT_SECRET\"),\n\t\tAccessToken:  os.Getenv(\"CHZZK_ACCESS_TOKEN\"),\n\t}\n}\n\ntype apiResponse[T any] struct {\n\tCode    int     `json:\"code\"`\n\tMessage *string `json:\"message\"`\n\tContent T       `json:\"content\"`\n}\n\nfunc decode[T any](resp *http.Response) (T, error) {\n\tdefer resp.Body.Close()\n\tvar result apiResponse[T]\n\tif err := json.NewDecoder(resp.Body).Decode(&result); err != nil {\n\t\tvar zero T\n\t\treturn zero, err\n\t}\n\tif result.Code != http.StatusOK {\n\t\tvar zero T\n\t\tmsg := \"unknown error\"\n\t\tif result.Message != nil {\n\t\t\tmsg = *result.Message\n\t\t}\n\t\treturn zero, fmt.Errorf(\"API error %d: %s\", result.Code, msg)\n\t}\n\treturn result.Content, nil\n}\n\nfunc (c *Client) get(ctx context.Context, path string) (*http.Response, error) {\n\treq, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\treq.Header.Set(\"Authorization\", \"Bearer \"+c.AccessToken)\n\treturn c.http.Do(req)\n}\n```\n\n")

	sb.WriteString("## 환경 변수 설정\n\n```.env\nCHZZK_CLIENT_ID=your_client_id\nCHZZK_CLIENT_SECRET=your_client_secret\nCHZZK_ACCESS_TOKEN=your_access_token\nCHZZK_REFRESH_TOKEN=your_refresh_token\n```\n\n")
	sb.WriteString("## 실행\n\n```bash\ngo mod init github.com/yourname/" + projectName + "\ngo mod tidy\ngo run .\n```\n")

	return sb.String()
}

// ─── TypeScript scaffold ──────────────────────────────────────────────────────

func scaffoldTypeScript(projectName string, features map[string]bool) string {
	bt := "`" // backtick helper to avoid raw-string nesting

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Chzzk TypeScript 프로젝트 스캐폴드: %s\n\n", projectName))

	// Directory structure
	sb.WriteString("## 디렉토리 구조\n\n```\n")
	sb.WriteString(projectName + "/\n├── src/\n")
	if features["auth"] {
		sb.WriteString("│   ├── auth.ts\n")
	}
	if features["live"] {
		sb.WriteString("│   ├── live.ts\n")
	}
	if features["chat"] {
		sb.WriteString("│   ├── chat.ts\n")
	}
	if features["channel"] {
		sb.WriteString("│   ├── channel.ts\n")
	}
	sb.WriteString("│   ├── client.ts\n│   └── index.ts\n├── package.json\n├── tsconfig.json\n└── .env\n```\n\n")

	// package.json
	sb.WriteString("## package.json\n\n```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"name\": \"" + projectName + "\",\n")
	sb.WriteString("  \"version\": \"1.0.0\",\n")
	sb.WriteString("  \"type\": \"module\",\n")
	sb.WriteString("  \"scripts\": {\n")
	sb.WriteString("    \"dev\": \"tsx src/index.ts\",\n")
	sb.WriteString("    \"build\": \"tsc\"\n")
	sb.WriteString("  },\n")
	sb.WriteString("  \"devDependencies\": {\n")
	sb.WriteString("    \"typescript\": \"^5.0.0\",\n")
	sb.WriteString("    \"tsx\": \"^4.0.0\"\n")
	sb.WriteString("  }\n}\n```\n\n")

	// tsconfig.json
	sb.WriteString("## tsconfig.json\n\n```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"compilerOptions\": {\n")
	sb.WriteString("    \"target\": \"ES2022\",\n")
	sb.WriteString("    \"module\": \"NodeNext\",\n")
	sb.WriteString("    \"moduleResolution\": \"NodeNext\",\n")
	sb.WriteString("    \"strict\": true,\n")
	sb.WriteString("    \"outDir\": \"dist\"\n")
	sb.WriteString("  },\n")
	sb.WriteString("  \"include\": [\"src\"]\n}\n```\n\n")

	// src/client.ts  — template literals expressed via bt variable
	sb.WriteString("## src/client.ts\n\n```typescript\n")
	sb.WriteString("const BASE_URL = \"https://openapi.chzzk.naver.com\";\n\n")
	sb.WriteString("interface ApiResponse<T> {\n")
	sb.WriteString("  code: number;\n")
	sb.WriteString("  message: string | null;\n")
	sb.WriteString("  content: T;\n}\n\n")
	sb.WriteString("export class ChzzkClient {\n")
	sb.WriteString("  private accessToken: string;\n\n")
	sb.WriteString("  constructor(\n")
	sb.WriteString("    private clientId = process.env.CHZZK_CLIENT_ID ?? \"\",\n")
	sb.WriteString("    private clientSecret = process.env.CHZZK_CLIENT_SECRET ?? \"\",\n")
	sb.WriteString("    accessToken = process.env.CHZZK_ACCESS_TOKEN ?? \"\"\n")
	sb.WriteString("  ) {\n    this.accessToken = accessToken;\n  }\n\n")
	sb.WriteString("  async request<T>(method: string, path: string, options: {\n")
	sb.WriteString("    query?: Record<string, string | number>;\n")
	sb.WriteString("    body?: unknown;\n")
	sb.WriteString("    auth?: boolean;\n")
	sb.WriteString("  } = {}): Promise<T> {\n")
	sb.WriteString("    let url = " + bt + "${BASE_URL}${path}" + bt + ";\n")
	sb.WriteString("    if (options.query) {\n")
	sb.WriteString("      const p = new URLSearchParams(\n")
	sb.WriteString("        Object.entries(options.query)\n")
	sb.WriteString("          .filter(([, v]) => v !== undefined)\n")
	sb.WriteString("          .map(([k, v]) => [k, String(v)])\n")
	sb.WriteString("      );\n")
	sb.WriteString("      if (p.size) url += \"?\" + p;\n")
	sb.WriteString("    }\n")
	sb.WriteString("    const headers: Record<string, string> = {};\n")
	sb.WriteString("    if (options.body) headers[\"Content-Type\"] = \"application/json\";\n")
	sb.WriteString("    if (options.auth) headers[\"Authorization\"] = " + bt + "Bearer ${this.accessToken}" + bt + ";\n")
	sb.WriteString("    const res = await fetch(url, {\n")
	sb.WriteString("      method, headers,\n")
	sb.WriteString("      body: options.body ? JSON.stringify(options.body) : undefined,\n")
	sb.WriteString("    });\n")
	sb.WriteString("    const data: ApiResponse<T> = await res.json();\n")
	sb.WriteString("    if (data.code !== 200)\n")
	sb.WriteString("      throw new Error(" + bt + "API error ${data.code}: ${data.message}" + bt + ");\n")
	sb.WriteString("    return data.content;\n")
	sb.WriteString("  }\n}\n```\n\n")

	// .env
	sb.WriteString("## .env\n\n```\n")
	sb.WriteString("CHZZK_CLIENT_ID=your_client_id\n")
	sb.WriteString("CHZZK_CLIENT_SECRET=your_client_secret\n")
	sb.WriteString("CHZZK_ACCESS_TOKEN=your_access_token\n")
	sb.WriteString("CHZZK_REFRESH_TOKEN=your_refresh_token\n")
	sb.WriteString("```\n\n")

	// run
	sb.WriteString("## 실행\n\n```bash\nnpm install\nnpm run dev\n```\n")

	_ = bt // used above
	return sb.String()
}

// ─── naming helpers ───────────────────────────────────────────────────────────

func goMethodName(ep Endpoint) string {
	path := ep.Path
	path = strings.TrimPrefix(path, "/open/v1/")
	path = strings.TrimPrefix(path, "/auth/v1/")
	path = strings.TrimPrefix(path, "/open/")
	parts := strings.FieldsFunc(path, func(r rune) bool { return r == '/' || r == '-' || r == '#' || r == '_' })
	var name strings.Builder
	switch ep.Method {
	case "GET":
		name.WriteString("Get")
	case "POST":
		name.WriteString("Create")
	case "PUT", "PATCH":
		name.WriteString("Update")
	case "DELETE":
		name.WriteString("Delete")
	}
	for _, p := range parts {
		if p == "" {
			continue
		}
		name.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return name.String()
}

func goFieldName(s string) string {
	s = strings.TrimSuffix(s, "[]")
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' || r == '.' })
	var sb strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		p = strings.ToUpper(p[:1]) + p[1:]
		// common Go acronyms
		switch strings.ToUpper(p) {
		case "Id":
			p = "ID"
		case "Url":
			p = "URL"
		}
		sb.WriteString(p)
	}
	return sb.String()
}

func goParamName(s string) string {
	s = strings.TrimSuffix(s, "[]")
	s = strings.TrimSuffix(s, ".")
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' || r == '.' })
	if len(parts) == 0 {
		return s
	}
	var sb strings.Builder
	sb.WriteString(strings.ToLower(parts[0]))
	for _, p := range parts[1:] {
		if p == "" {
			continue
		}
		sb.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return sb.String()
}

func goType(t string) string {
	switch t {
	case "int", "long":
		return "int"
	case "bool":
		return "bool"
	case "string[]":
		return "[]string"
	default:
		return "string"
	}
}

func goOptType(t string) string {
	switch t {
	case "int", "long":
		return "int"
	case "bool":
		return "bool"
	case "string[]":
		return "[]string"
	default:
		return "string"
	}
}

func goZero(t string) string {
	switch t {
	case "int", "long":
		return "0"
	case "bool":
		return "false"
	default:
		return `""`
	}
}

func tsMethodName(ep Endpoint) string {
	name := goMethodName(ep)
	if name == "" {
		return "call"
	}
	return strings.ToLower(name[:1]) + name[1:]
}

func tsInterfaceName(ep Endpoint) string {
	return goMethodName(ep)
}

func tsParamName(s string) string {
	return goParamName(s)
}

func tsType(t string) string {
	switch t {
	case "int", "long":
		return "number"
	case "bool":
		return "boolean"
	case "string[]":
		return "string[]"
	default:
		return "string"
	}
}

// RegisterCodegenTools adds code generation tools to the MCP server.
func RegisterCodegenTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name: "chzzk_generate_auth_code",
		Description: "치지직 OAuth2 인증 플로우 완성 코드를 생성합니다. " +
			"인가 코드 요청 URL 생성, 토큰 발급/갱신/폐기 함수, 콜백 서버(Go)를 포함합니다. " +
			"환경 변수: CHZZK_CLIENT_ID, CHZZK_CLIENT_SECRET, CHZZK_REDIRECT_URI. " +
			"지원 언어: go, typescript",
	}, handleGenerateAuthCode)

	mcp.AddTool(s, &mcp.Tool{
		Name: "chzzk_generate_api_client",
		Description: "지정한 치지직 API 엔드포인트에 대한 타입 안전 HTTP 클라이언트 코드를 생성합니다. " +
			"endpoints는 'METHOD /path' 형식으로 전달하세요 (예: 'GET /open/v1/lives'). " +
			"사용 가능한 엔드포인트는 chzzk_list_apis 또는 chzzk_get_api_spec으로 확인하세요. " +
			"지원 언어: go, typescript",
	}, handleGenerateAPIClient)

	mcp.AddTool(s, &mcp.Tool{
		Name: "chzzk_scaffold_project",
		Description: "치지직 API를 연동하는 서비스의 프로젝트 보일러플레이트를 생성합니다. " +
			"선택한 features(auth, live, chat, channel, session)에 맞는 디렉토리 구조, " +
			"기본 클라이언트, 설정 파일, 환경 변수 예시를 포함합니다. " +
			"지원 언어: go, typescript",
	}, handleScaffoldProject)
}
