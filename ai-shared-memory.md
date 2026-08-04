## Example Usage

- **2025-06-14:sum.test.ts sum function bug** Issue found in test case file `sum.test.ts`, line 10. Confusion between `sum` and `sum2` functions, leading to incorrect results. FIX: removing `sum2` function, and replacing all `sum` calls with `sum2` calls. Combine both functions into one function, and having a parameter to determine which function to call.

- **2026-08-04:ai-interface core (openai-go v3 wrapper)** Built the core AI interface for the knowledge-tree bot on branch `gabe/core-ai-features`: a thin wrapper around the official `github.com/openai/openai-go/v3` SDK targeting a self-hosted Ollama endpoint (`OPENAI_BASE_URL`). Issues resolved and decisions made:
  - **Go toolchain too old**: goenv was pinning Go 1.21.0, but openai-go v3 requires Go 1.23+ (uses stdlib `iter` package). FIX: `goenv local system` → uses Homebrew Go 1.26.5. If `go test` fails with "package iter is not in std", GOROOT is stale — prefix with `GOROOT=/opt/homebrew/Cellar/go/1.26.5/libexec`.
  - **SDK client is a value, not a pointer**: `openai.NewClient()` returns `openai.Client` (value) — wrap with `&c` when storing in a wrapper struct.
  - **param.Opt[string] helper**: embedding input takes `openai.String(text)` for `EmbeddingNewParamsInputUnion.OfString` — avoids importing the `param` package.
  - **Base URL needs zero wrapper code**: SDK auto-reads `OPENAI_BASE_URL` / `OPENAI_API_KEY` from env; Ollama's `/v1` endpoint works with no options passed.
  - **Design decisions**: own `Message{Role, Content}` type (no SDK types leak into signatures); unexported `api` interface seam so tests run fully offline via a hand-written fake (stdlib only, no gomock, zero network); model precedence = constructor override > env var > fail-loud error naming the missing var (NO baked model literals — no `gpt-*`/`gemma3:1b`/`embeddinggemma` in .go files); `resolveModel` trims whitespace; `Chat()` maps roles to `SystemMessage`/`UserMessage`/`AssistantMessage` and errors on unknown role and empty `Choices`; `Embed()` returns `[]float64` passthrough and errors on empty `Data`; `Client` is immutable + goroutine-safe.
  - **Deliverables**: `go.mod`/`go.sum` (module `github.com/Guuzzeji/teemysu`, direct dep `openai-go/v3 v3.50.0` — promoted out of `// indirect`), the ai package, root `main.go` demo that prints configured models and exits 1 naming the missing var (no API calls), `.env.example` documenting the 4-var contract (docs only, never parsed). Verified via `make build-local` and `make go-run`.
  - **Open / in-flight**: working tree has `internal/ai` deleted and the code moved to `src/ai` (untracked), but `main.go` still imports `internal/ai` → tree does not build until the import is updated or the location is reverted.

## Notes

<!-- Add any additional notes here -->
