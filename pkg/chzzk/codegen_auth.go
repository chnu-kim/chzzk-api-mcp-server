package chzzk

func authCodeGo() string {
	return `// Package auth provides Chzzk OAuth2 authentication helpers.
// Required environment variables:
//   CHZZK_CLIENT_ID     - 애플리케이션 클라이언트 ID
//   CHZZK_CLIENT_SECRET - 애플리케이션 클라이언트 시크릿
//   CHZZK_REDIRECT_URI  - 인가 코드를 전달받을 리다이렉트 URI
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	baseURL    = "https://openapi.chzzk.naver.com"
	authURL    = "https://chzzk.naver.com/account-interlock"
	tokenPath  = "/auth/v1/token"
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
// state must echo the value sent in AuthorizationURL (Chzzk API requirement).
func (c Config) ExchangeCode(ctx context.Context, code, state string) (*TokenResponse, error) {
	return postToken(ctx, map[string]string{
		"grantType":    "authorization_code",
		"clientId":     c.ClientID,
		"clientSecret": c.ClientSecret,
		"code":         code,
		"state":        state,
	})
}

// RefreshToken uses a refresh token to obtain a new access token.
func (c Config) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	return postToken(ctx, map[string]string{
		"grantType":    "refresh_token",
		"clientId":     c.ClientID,
		"clientSecret": c.ClientSecret,
		"refreshToken": refreshToken,
	})
}

// RevokeToken invalidates the given token.
func (c Config) RevokeToken(ctx context.Context, token, tokenTypeHint string) error {
	data := map[string]string{
		"clientId":     c.ClientID,
		"clientSecret": c.ClientSecret,
		"token":        token,
	}
	if tokenTypeHint != "" {
		data["tokenTypeHint"] = tokenTypeHint
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+revokePath,
		bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
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

func postToken(ctx context.Context, data map[string]string) (*TokenResponse, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+tokenPath,
		bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

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

func authCodeTypeScript() string {
	bt := "`"
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
  scopes: string[];
}

interface RawTokenContent {
  accessToken: string;
  refreshToken: string;
  tokenType: string;
  expiresIn: number | string;
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
  return ` + bt + `${AUTH_URL}?${params}` + bt + `;
}

/** Exchanges an authorization code for tokens.
 *  state must echo the value sent in buildAuthorizationURL (Chzzk API requirement). */
export async function exchangeCode(
  config: ChzzkAuthConfig,
  code: string,
  state: string
): Promise<TokenResponse> {
  return postToken({
    grantType: "authorization_code",
    clientId: config.clientId,
    clientSecret: config.clientSecret,
    code,
    state,
  });
}

/** Uses a refresh token to obtain a new access token. */
export async function refreshAccessToken(
  config: ChzzkAuthConfig,
  refreshToken: string
): Promise<TokenResponse> {
  return postToken({
    grantType: "refresh_token",
    clientId: config.clientId,
    clientSecret: config.clientSecret,
    refreshToken,
  });
}

/** Invalidates the given token. */
export async function revokeToken(
  config: ChzzkAuthConfig,
  token: string,
  tokenTypeHint?: "access_token" | "refresh_token"
): Promise<void> {
  const body: Record<string, string> = {
    clientId: config.clientId,
    clientSecret: config.clientSecret,
    token,
  };
  if (tokenTypeHint) body.tokenTypeHint = tokenTypeHint;

  const res = await fetch(` + bt + `${BASE_URL}/auth/v1/token/revoke` + bt + `, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const data: ApiResponse<unknown> = await res.json();
  if (data.code !== 200) {
    throw new Error(` + bt + `Revoke token failed ${data.code}: ${data.message}` + bt + `);
  }
}

async function postToken(params: Record<string, string>): Promise<TokenResponse> {
  const res = await fetch(` + bt + `${BASE_URL}/auth/v1/token` + bt + `, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(params),
  });
  const data: ApiResponse<RawTokenContent> = await res.json();
  if (data.code !== 200) {
    throw new Error(` + bt + `Token error ${data.code}: ${data.message}` + bt + `);
  }
  const { accessToken, refreshToken, tokenType, expiresIn, scope } = data.content;
  return {
    accessToken,
    refreshToken,
    tokenType,
    expiresIn: Number(expiresIn),
    scopes: scope ? scope.split(" ") : [],
  };
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
