package chzzk

import (
	"fmt"
	"strings"
)

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

	sb.WriteString(fmt.Sprintf("// %s calls %s %s.\n", name, ep.Method, ep.Path))
	if ep.Description != "" {
		sb.WriteString(fmt.Sprintf("// %s\n", ep.Description))
	}

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
	var optParams []string
	for _, p := range ep.QueryParams {
		if !p.Required {
			optParams = append(optParams, fmt.Sprintf("%s %s", goParamName(p.Name), goType(p.Type)))
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

	if len(ep.Response) > 0 {
		sb.WriteString(fmt.Sprintf("  // %s\n", ep.Description))
		iName := goMethodName(ep) + "Response"
		sb.WriteString(fmt.Sprintf("  // Response type: %s\n", iName))
	}

	sb.WriteString(fmt.Sprintf("  /** %s\n   * %s %s\n   */\n", ep.Description, ep.Method, ep.Path))
	sb.WriteString(fmt.Sprintf("  async %s(", name))

	var params []string
	for _, p := range ep.QueryParams {
		if p.Required {
			params = append(params, fmt.Sprintf("%s: %s", goParamName(p.Name), tsType(p.Type)))
		}
	}
	for _, p := range ep.BodyParams {
		if p.Required {
			params = append(params, fmt.Sprintf("%s: %s", goParamName(p.Name), tsType(p.Type)))
		}
	}
	var optParams []string
	for _, p := range ep.QueryParams {
		if !p.Required {
			optParams = append(optParams, fmt.Sprintf("%s?: %s", goParamName(p.Name), tsType(p.Type)))
		}
	}
	params = append(params, optParams...)

	sb.WriteString(strings.Join(params, ", "))

	retType := "void"
	if len(ep.Response) > 0 {
		retType = goMethodName(ep) + "Response"
	}
	sb.WriteString(fmt.Sprintf("): Promise<%s> {\n", retType))

	if len(ep.QueryParams) > 0 {
		sb.WriteString("    const query: Record<string, string | number> = {};\n")
		for _, p := range ep.QueryParams {
			pn := goParamName(p.Name)
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
			pn := goParamName(p.Name)
			bodyFields = append(bodyFields, fmt.Sprintf("%s: %s", pn, pn))
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

func lowerFirst(s string) string { return strings.ToLower(s[:1]) + s[1:] }

func goFieldName(s string) string {
	s = strings.TrimSuffix(s, "[]")
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' || r == '.' })
	var sb strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		p = strings.ToUpper(p[:1]) + p[1:]
		switch {
		case strings.HasSuffix(p, "Id"):
			p = strings.TrimSuffix(p, "Id") + "ID"
		case strings.HasSuffix(p, "Url"):
			p = strings.TrimSuffix(p, "Url") + "URL"
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
	sb.WriteString(lowerFirst(parts[0]))
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
	return lowerFirst(name)
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
