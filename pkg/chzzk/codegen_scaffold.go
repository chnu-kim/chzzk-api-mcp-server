package chzzk

import (
	"fmt"
	"strings"
)

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

	sb.WriteString("## .env\n\n```\n")
	sb.WriteString("CHZZK_CLIENT_ID=your_client_id\n")
	sb.WriteString("CHZZK_CLIENT_SECRET=your_client_secret\n")
	sb.WriteString("CHZZK_ACCESS_TOKEN=your_access_token\n")
	sb.WriteString("CHZZK_REFRESH_TOKEN=your_refresh_token\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## 실행\n\n```bash\nnpm install\nnpm run dev\n```\n")

	return sb.String()
}
