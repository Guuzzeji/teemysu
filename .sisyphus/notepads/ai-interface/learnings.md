## 2026-08-04T00:33:12Z Task 1: go mod init
- goenv was pinning Go 1.21.0; openai-go v3 needs Go 1.23+ (uses iter package)
- Fixed with: goenv local system → uses /opt/homebrew Go 1.26.5
- openai-go/v3 v3.50.0 added; transitive deps: tidwall/gjson, tidwall/sjson, tidwall/match, tidwall/pretty
- go.mod shows all deps as '// indirect' until actual .go imports exist — normal, go mod tidy fixes later

## 2026-08-04 Task 2: internal/ai package (TDD)
- GOROOT mismatch: goenv sets GOROOT to 1.21.0 but binary is 1.26.5 — use `GOROOT=/opt/homebrew/Cellar/go/1.26.5/libexec` prefix for tests
- `param` package is at `github.com/openai/openai-go/v3/packages/param` (not `/param`) — prefer `openai.String(text)` helper instead, avoids import
- `openai.NewClient()` returns value (not pointer) — wrap with `&c` when storing in sdkClient
- `EmbeddingNewParamsInputUnion.OfString` takes `param.Opt[string]` — use `openai.String(text)` to create it
- `shared.ChatModel` and `EmbeddingModel` are both `= string` aliases — model params accept plain strings
- `sdkClient` wraps `*openai.Client`; test seam is unexported `api` interface — allows fakeAPI in tests without SDK calls
- resolveModel intentionally does NOT know env var names — caller constructs error with specific env var name

## 2026-08-04 Task 3: Client.Chat method (TDD)
- Client.Chat is a one-liner: `c.api.Chat(ctx, c.chatModel, msgs)` — delegates entirely to the api seam
- fakeAPI needed extension: chatErr/embedErr fields for error injection, embedCalls counter
- TestChatUnknownRole tests sdkClient directly (not through Client), nil client is safe because role validation happens before API call
- TestChatEmptyChoices: empty string from fake is valid — Client.Chat doesn't validate content, that's sdkClient's job
- gopls not available in Go 1.26.5 (only in 1.21.0) — build+tests are verification, no LSP diagnostics
