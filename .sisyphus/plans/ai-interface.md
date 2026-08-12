# AI Interface — OpenAI SDK Wrapper for Self-Hosted Ollama (Core)

## TL;DR

> **Quick Summary**: Initialize Go module and build `internal/ai` — a thin, dumb-pipe wrapper around official `openai-go/v3` targeting a self-hosted Ollama endpoint (`OPENAI_BASE_URL`). Exposes `Chat` + `Embed`; configured chat/embedding models tracked as fields with getters. No model defaults — env required, fail loud. TDD, offline tests, no live API.
>
> **Deliverables**:
> - `go.mod` (module `github.com/Guuzzeji/teemysu`) + `openai-go/v3` dependency
> - `internal/ai` package: `New`, `NewWithModels`, `Chat`, `Embed`, `ChatModel()`, `EmbedModel()`
> - `internal/ai/*_test.go` — full offline test coverage (fake seam, no network)
> - Root `main.go` demo printing configured models (no API calls; exit 1 naming missing env var)
> - `.env.example` — 4-var env contract doc (no loader, docs only)
> - Verified `make build-local` / `make go-run`
>
> **Estimated Effort**: Short
> **Parallel Execution**: YES — 4 waves (Chat ∥ Embed in wave 3)
> **Critical Path**: Task 1 → Task 2 → Task 3/4 → Task 5 → F1-F4

---

## Context

### Original Request
"creating a simple ai interface that acts as a wrapper for my open-ai sdk, and allows me to keep track of what chat model i am using and embedding model I am using... only focus on creating the core features of that before anything else"

### Interview Summary
**Key Discussions** (grill-me session, ALL user-confirmed):
- SDK: official `github.com/openai/openai-go/v3`
- **Deployment target: self-hosted Ollama** at `http://fedora:11434/v1` — NOT OpenAI cloud. Chat model `gemma3:1b`, embedding model `embeddinggemma` (both env-configured, never baked into code)
- Goal: simple RAG-base chat + info organization (later plans); this wrapper is the foundation
- Model tracking: struct fields + getters only — no logging, no token stats
- Env config (4 vars, NO defaults, fail loud if missing): `OPENAI_BASE_URL`, `OPENAI_API_KEY` (dummy — Ollama ignores), `OPENAI_CHAT_MODEL`, `OPENAI_EMBED_MODEL`
- Model switching = redeploy with new env (restart-switch). NO setter method — live-switch deferred until a Discord command needs it (YAGNI)
- No streaming — full response at once. Deferred until bot exists and answers feel slow
- Dumb pipe — wrapper takes `[]Message` as-is. No system-prompt injection, no history management, no RAG context slotting (all caller-side, future bot layer)
- SDK defaults kept: built-in retries (2, backoff) ON; timeouts caller-owned via `ctx`; errors passthrough; Client immutable + goroutine-safe (no mutex needed)

**Research Findings** (context7, current docs verified):
- `openai.NewClient()` auto-reads env: `OPENAI_API_KEY`, `OPENAI_BASE_URL`, `OPENAI_ORG_ID`, `OPENAI_PROJECT_ID` (+others) — base URL support needs ZERO wrapper code
- `option.WithBaseURL("http://...")` available for explicit override if ever needed
- `client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{Messages, Model})` → `completion.Choices[0].Message.Content`
- Message helpers: `openai.SystemMessage(string)`, `openai.UserMessage(string)`, `openai.AssistantMessage(string)`
- `client.Embeddings.New(ctx, openai.EmbeddingNewParams{Input: openai.F[openai.EmbeddingNewParamsInputUnion](openai.EmbeddingNewParamsInputUnion{OfString: openai.String(text)}), Model})` → `embedding.Data[0].Embedding` (`[]float64`)
- Ollama serves OpenAI-compatible chat + embeddings under one `/v1` URL

### Metis Review + Grill-Me Resolutions
**Identified Gaps** (addressed):
- Module path unknown → resolved from git remote: `github.com/Guuzzeji/teemysu`
- Leaking SDK message types → own `Message{Role, Content}` struct, mapped internally
- `[]float32` vs `[]float64` → `[]float64` passthrough (SDK native, zero conversion)
- Live API in tests → forbidden; unexported interface seam + fake
- ~~Model defaults `gpt-4o-mini`~~ → **grill-me pivot**: self-hosted Ollama made OpenAI defaults meaningless → NO defaults at all; precedence: override > env > ERROR (fail loud at boot)
- Empty API key → no local validation; dummy key documented for Ollama (`OPENAI_API_KEY=ollama`)
- "On the fly" ambiguity → restart-switch confirmed, no setter
- Streaming re-check → excluded, confirmed against RAG-chat goal (1B local model = short fast answers)
- RAG assembly location → caller-side; wrapper is dumb pipe

---

## Work Objectives

### Core Objective
Thin, tested Go wrapper over openai-go v3 that always knows its configured chat + embedding models. Foundation for later Discord bot work — nothing more.

### Concrete Deliverables
- `go.mod` + `go.sum` with `github.com/openai/openai-go/v3`
- `internal/ai/ai.go` — wrapper (may split into `client.go`/`chat.go`/`embed.go` if cleaner; same package)
- `internal/ai/ai_test.go` (+ `chat_test.go`/`embed_test.go` if split)
- `main.go` — print-models demo
- `dist/bot` builds via existing makefile

### Definition of Done
- [ ] `go test ./...` → PASS, exit 0 (fully offline, no env needed)
- [ ] `make build-local` → `./dist/bot` exists and is executable
- [ ] Env set → `make go-run` prints both configured model names, exit 0
- [ ] Env unset → `make go-run` exit 1, stderr names the missing variable
- [ ] No dependency beyond `openai-go/v3` (+ transitive) in go.mod

### Must Have
- Module path `github.com/Guuzzeji/teemysu`
- Target: self-hosted Ollama via OpenAI-compatible endpoint — SDK reads `OPENAI_BASE_URL` natively (zero wrapper code for base URL)
- `New() (*Client, error)` — reads `OPENAI_CHAT_MODEL` / `OPENAI_EMBED_MODEL` env; **error naming missing var if unset/whitespace — NO defaults**
- `NewWithModels(chatModel, embedModel string) (*Client, error)` — args take precedence; empty/whitespace arg falls through to env; still errors if nothing resolves
- Precedence: constructor override > env var > **ERROR (fail loud)**
- `Chat(ctx, []Message) (string, error)` — own `Message{Role, Content}` type, no SDK types in signature; messages passed as-is (dumb pipe)
- `Embed(ctx, string) ([]float64, error)`
- `ChatModel() string`, `EmbedModel() string` — pure getters, no I/O
- Client immutable after construction — goroutine-safe for future concurrent Discord handlers
- TDD: failing tests written before implementation (CLAUDE.md mandate)
- Tests offline: unexported interface seam, fake injected in tests, zero network
- `.env.example` — documents the 4-var contract with commented example values (docs only, NOT loaded by code)
- Comments only on non-obvious logic (CLAUDE.md: complex logic only, no comment noise)

### Must NOT Have (Guardrails)
- Discord / discordgo, SQLite / go-sqlite3, knowledge tree, tagging, RAG assembly
- **Model name constants or defaults baked into code** (no `gpt-*`, no `gemma3:1b`, no `embeddinggemma` literals in .go files — env/docs only)
- **Setter/mutator methods** (live model-switch deferred until Discord command exists)
- **System-prompt injection, history management, RAG context slotting** (caller-side, future bot layer)
- **Streaming** (deferred until bot exists and answers feel slow)
- Token usage tracking, logging middleware, metrics, per-call logging
- Tool/function calling, custom retry/backoff wrappers (SDK built-in retries suffice), rate limiting
- Config files or loaders (yaml/json/viper/godotenv) — OS env only; `.env.example` is documentation, never parsed
- Public interface "for future consumers" — unexported test seam only
- Mock codegen frameworks (gomock/mockery) — hand-written fake, stdlib only
- Any dependency beyond `openai-go/v3` without asking human first
- Live API calls anywhere (tests, main demo)
- Makefile expansion (no lint/cover/docker targets; touch only if build breaks)
- `ai-shared-memory.md` entries unless a real bug is found and fixed

---

## Verification Strategy (MANDATORY)

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed.

### Test Decision
- **Infrastructure exists**: NO (greenfield) — but CLAUDE.md mandates basic go testing + TDD
- **Automated tests**: YES (TDD) — stdlib `testing` only
- **Framework**: `go test` (no external test deps)
- **TDD flow**: RED (failing test) → GREEN (minimal impl) → REFACTOR, per task

### QA Policy
Every task has agent-executed QA scenarios runnable via Bash. Evidence (terminal output) saved to `.sisyphus/evidence/task-{N}-{scenario}.{ext}`.

- **Go code**: Bash runs `go test` with exact `-run` patterns and env setups; assert exit code + output strings
- **Build/demo**: Bash runs `make` targets; assert exit code, file existence, stdout content

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (foundation):
└── Task 1: go mod init + openai-go/v3 dependency [quick]

Wave 2 (package core — blocks Chat/Embed):
└── Task 2: internal/ai core — Message type, Client struct, constructors, model resolution, getters (TDD) [quick]

Wave 3 (operations — MAX PARALLEL, both depend only on Task 2):
├── Task 3: Chat method + tests via fake seam [quick]
└── Task 4: Embed method + tests via fake seam [quick]

Wave 4 (wiring):
└── Task 5: main.go demo + makefile verification [quick]

Wave FINAL (4 parallel reviews, then user okay):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real manual QA (unspecified-high)
└── Task F4: Scope fidelity check (deep)

Critical Path: 1 → 2 → 3/4 → 5 → F1-F4 → user okay
Max Concurrent: 2 (Wave 3)
Note: small plan — waves below 3-task target because scope is deliberately minimal (ponytail). Splitting further would be artificial.
```

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|------|-----------|--------|------|
| 1 | — | 2 | 1 |
| 2 | 1 | 3, 4, 5 | 2 |
| 3 | 2 | 5 | 3 |
| 4 | 2 | 5 | 3 |
| 5 | 2, 3, 4 | — | 4 |

### Agent Dispatch Summary

- **Wave 1**: T1 → `quick`
- **Wave 2**: T2 → `quick` + `test-driven-development`
- **Wave 3**: T3, T4 → `quick` + `test-driven-development` (parallel)
- **Wave 4**: T5 → `quick`
- **FINAL**: F1 → `oracle`, F2 → `unspecified-high`, F3 → `unspecified-high`, F4 → `deep`

---

## TODOs

> Implementation + Test = ONE Task. Never separate.

- [x] 1. Initialize Go module + openai-go dependency

  **What to do**:
  - `go mod init github.com/Guuzzeji/teemysu`
  - `go get github.com/openai/openai-go/v3@latest`
  - Verify module compiles empty: `go build ./...`
  - No test files needed (nothing to test yet) — QA via commands below

  **Must NOT do**:
  - Add any other dependency
  - Create any .go source files (Task 2+ owns those)
  - Run `go mod tidy` against nonexistent imports

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: two commands + verification, trivial scope
  - **Skills**: []
    - No skills needed — pure command execution
  - **Skills Evaluated but Omitted**:
    - `test-driven-development`: no logic exists to test yet

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 1 (alone)
  - **Blocks**: Task 2
  - **Blocked By**: None

  **References**:
  - `makefile:1-9` — existing targets expect `main.go` at root eventually; this task only ensures module builds
  - External: `https://pkg.go.dev/github.com/openai/openai-go/v3` — confirm latest v3.x version tag for `go get`
  - **WHY**: module path must be `github.com/Guuzzeji/teemysu` (git remote origin) — wrong path = import pain forever

  **Acceptance Criteria**:
  - [ ] `test -f go.mod && head -1 go.mod` → `module github.com/Guuzzeji/teemysu`
  - [ ] `grep 'openai/openai-go/v3' go.mod` → match found
  - [ ] `go build ./...` → exit 0
  - [ ] `grep -c '	require' go.mod` → exactly 1 direct require block entry (openai-go only; indirect may exist)

  **QA Scenarios**:

  ```
  Scenario: Module initializes and builds clean
    Tool: Bash
    Preconditions: repo root, branch gabe/core-ai-features, no go.mod present
    Steps:
      1. Run: go mod init github.com/Guuzzeji/teemysu && go get github.com/openai/openai-go/v3@latest
      2. Run: go build ./...
      3. Assert: exit code 0 from both
      4. Run: cat go.mod
      5. Assert: first line is "module github.com/Guuzzeji/teemysu" and file contains "openai-go/v3"
    Expected Result: go.mod + go.sum exist, build exits 0, single direct dependency
    Failure Indicators: wrong module path, extra direct deps, build errors, network failure during go get
    Evidence: .sisyphus/evidence/task-1-module-init.txt

  Scenario: No stray dependencies
    Tool: Bash
    Preconditions: task 1 complete
    Steps:
      1. Run: go list -m all
      2. Assert: only teemysu + openai-go/v3 + its transitive deps listed
      3. Run: grep -E 'discordgo|sqlite|godotenv|viper|gomock' go.mod || echo "CLEAN"
      4. Assert: output is "CLEAN"
    Expected Result: dependency graph minimal, no forbidden libraries
    Failure Indicators: any forbidden dep in go.mod or module graph
    Evidence: .sisyphus/evidence/task-1-deps-clean.txt
  ```

  **Evidence to Capture**:
  - [ ] `task-1-module-init.txt` — full terminal output of init + get + build + cat go.mod
  - [ ] `task-1-deps-clean.txt` — output of `go list -m all` + grep check

  **Commit**: YES
  - Message: `chore(build): init go module with openai-go v3`
  - Files: `go.mod`, `go.sum`
  - Pre-commit: `go build ./...`

- [x] 2. internal/ai core — types, constructors, model resolution, getters (TDD)

  **What to do**:
  - **RED first**: write `internal/ai/ai_test.go` with failing tests for everything below, watch them fail
  - `Message` struct: `{Role, Content string}`; role constants `RoleSystem`, `RoleUser`, `RoleAssistant` (typed string)
  - `Client` struct: holds SDK client (via unexported seam — see below), `chatModel`, `embedModel` strings
  - Unexported test seam: `type api interface { Chat(ctx context.Context, model string, msgs []Message) (string, error); Embed(ctx context.Context, model, text string) ([]float64, error) }` — real SDK impl comes in Tasks 3/4; tests inject fake
  - `New() (*Client, error)` — reads `OPENAI_CHAT_MODEL` / `OPENAI_EMBED_MODEL` env; constructs SDK client via `openai.NewClient()` (SDK natively reads `OPENAI_API_KEY` + `OPENAI_BASE_URL` — pass no options)
  - `NewWithModels(chatModel, embedModel string) (*Client, error)` — same but args take precedence
  - Resolution helper: `resolveModel(override, envVal string) (string, error)` — trim whitespace; override if non-empty, else envVal if non-empty, else **error naming which model/env var is missing — NO default fallback**
  - `ChatModel() string`, `EmbedModel() string` — pure getters
  - **GREEN**: minimal implementation until tests pass
  - **REFACTOR**: comment the precedence rule (non-obvious: fail-loud instead of defaults); no other comments

  **Must NOT do**:
  - Implement Chat/Embed methods (Tasks 3/4)
  - **Any model default, constant, or literal** (`gpt-*`, `gemma3:1b`, `embeddinggemma` must NOT appear in .go files — test files MAY use them as arbitrary example strings)
  - Validate API key or base URL presence (SDK/Ollama handle at call time)
  - Pass `option.WithBaseURL` explicitly (SDK reads `OPENAI_BASE_URL` env natively — zero code)
  - Export the `api` interface or any mock
  - Read `.env` files — OS env only
  - Validate model names against any catalog (trust configured string)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: single package, ~100 lines, well-specified types and resolution rules
  - **Skills**: [`test-driven-development`]
    - `test-driven-development`: CLAUDE.md mandates TDD; this task is RED-GREEN-REFACTOR by design
  - **Skills Evaluated but Omitted**:
    - `writing-plans`: plan already exists, this is execution
    - `systematic-debugging`: no bug involved

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 2 (alone)
  - **Blocks**: Tasks 3, 4, 5
  - **Blocked By**: Task 1

  **References**:
  - `.sisyphus/drafts/ai-interface.md` — full decision log (precedence rules, seam design, defaults)
  - `CLAUDE.md` — TDD mandate, minimal deps, comment discipline (complex logic only)
  - External: `https://github.com/openai/openai-go` client-initialization docs — `openai.NewClient()` auto-reads `OPENAI_API_KEY`; no key validation at construction
  - **WHY the seam is unexported**: Metis guardrail — no public interface "for future consumers" (YAGNI); seam exists ONLY so Tasks 3/4 tests avoid network

  **Acceptance Criteria**:
  - [ ] `go test ./internal/ai/ -run TestNewMissingEnv -v` → PASS: unset env → error naming the missing variable
  - [ ] `go test ./internal/ai/ -run TestNewEnv -v` → PASS: env vars populate both models
  - [ ] `go test ./internal/ai/ -run TestNewWithModels -v` → PASS: args override env; empty arg falls through to env; both empty → error
  - [ ] `go test ./internal/ai/ -run TestResolveModel -v` → PASS: table-driven — whitespace trim, empty-string, precedence chain, error case
  - [ ] `go test ./internal/ai/ -run TestGetters -v` → PASS: getters return configured strings
  - [ ] All above pass with NO env set at all — zero network
  - [ ] `grep -rn 'gpt-\|gemma3\|embeddinggemma\|nomic-embed' internal/ai/*.go --include='!*_test.go'` → no matches (no baked model literals)
  - [ ] `grep -n 'openai\.' internal/ai/ai.go` → SDK types NOT in any exported signature (Message, New, NewWithModels, getters)

  **QA Scenarios**:

  ```
  Scenario: Model resolution precedence chain works
    Tool: Bash
    Preconditions: task 2 code exists, repo root
    Steps:
      1. Run: env -u OPENAI_CHAT_MODEL -u OPENAI_EMBED_MODEL -u OPENAI_API_KEY go test ./internal/ai/ -run 'TestNewMissingEnv' -v
      2. Assert: PASS; test asserts error message contains "OPENAI_CHAT_MODEL"
      3. Run: OPENAI_CHAT_MODEL=env-chat OPENAI_EMBED_MODEL=env-embed go test ./internal/ai/ -run 'TestNewEnv' -v
      4. Assert: PASS
      5. Run: go test ./internal/ai/ -run 'TestNewWithModels|TestResolveModel|TestGetters' -v
      6. Assert: PASS, all table cases green
    Expected Result: go test exits 0 for all runs; precedence override > env > ERROR proven by tests
    Failure Indicators: any FAIL line; error message not naming the missing var; a silent default appearing anywhere
    Evidence: .sisyphus/evidence/task-2-model-resolution.txt

  Scenario: Tests are fully offline
    Tool: Bash
    Preconditions: task 2 complete
    Steps:
      1. Run: env -u OPENAI_API_KEY -u OPENAI_BASE_URL HTTPS_PROXY=http://127.0.0.1:1 go test ./... -count=1
      2. Assert: exit 0 (proxy blackhole would kill any network call)
      3. Run: grep -rn 'http\.\|httptest' internal/ai/*_test.go || echo "NO NETWORK REFS"
      4. Assert: no real HTTP usage in tests
    Expected Result: tests pass even with network blackholed
    Failure Indicators: test failure/timeout under broken proxy = network dependency leaked in
    Evidence: .sisyphus/evidence/task-2-offline-proof.txt
  ```

  **Evidence to Capture**:
  - [ ] `task-2-model-resolution.txt` — all four go test runs with -v output
  - [ ] `task-2-offline-proof.txt` — blackhole-proxy test run output

  **Commit**: YES
  - Message: `feat(ai): add client wrapper with model config resolution`
  - Files: `internal/ai/ai.go`, `internal/ai/ai_test.go`
  - Pre-commit: `go test ./internal/ai/`

- [x] 3. Chat method + tests via fake seam (TDD)

  **What to do**:
  - **RED first**: write `internal/ai/chat_test.go` — fake `api` returning canned response/error; failing tests for cases below
  - `Chat(ctx context.Context, msgs []Message) (string, error)` on `*Client` — delegates to seam with `c.chatModel`
  - SDK impl of seam (`sdkAPI` or method on unexported struct wrapping `openai.Client`): maps `Message` → `openai.SystemMessage/UserMessage/AssistantMessage` by Role; unknown role → error before any SDK call
  - Call `client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{Messages: mapped, Model: c.chatModel})`
  - Empty `completion.Choices` → return descriptive error
  - Return `Choices[0].Message.Content, nil` (content may be empty string — pass through)
  - Empty `msgs` → pass through to seam (SDK/API errors surface; no local guard)
  - **GREEN** → **REFACTOR**

  **Must NOT do**:
  - Streaming, n>1 choices, tool calls, temperature/options params (defaults only)
  - Retry on error, wrap errors in custom types
  - Message history storage — stateless per call
  - Leak `openai.ChatCompletionMessageParamUnion` into exported signatures

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: one method + param mapping + fake-based tests, ~80 lines
  - **Skills**: [`test-driven-development`]
    - `test-driven-development`: CLAUDE.md TDD mandate; RED-GREEN-REFACTOR specified in task
  - **Skills Evaluated but Omitted**:
    - `systematic-debugging`: no existing broken behavior

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Task 4)
  - **Blocks**: Task 5
  - **Blocked By**: Task 2

  **References**:
  - External: `https://github.com/openai/openai-go/blob/main/_autodocs/chat-completions-api.md` — verified call pattern: `client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{Messages: []openai.ChatCompletionMessageParamUnion{openai.SystemMessage(...), openai.UserMessage(...)}, Model: model})` → `completion.Choices[0].Message.Content`
  - `internal/ai/ai.go` (Task 2 output) — `api` seam signature Chat must satisfy; role constants to map
  - **WHY own Message type**: keeps whole app decoupled from openai-go; if SDK swapped later, only this mapping changes

  **Acceptance Criteria**:
  - [ ] `go test ./internal/ai/ -run TestChat -v` → PASS: happy path returns canned content via fake
  - [ ] `go test ./internal/ai/ -run TestChatEmptyChoices -v` → PASS: non-nil error when fake returns empty choices
  - [ ] `go test ./internal/ai/ -run TestChatError -v` → PASS: seam error propagates to caller
  - [ ] `go test ./internal/ai/ -run TestChatUnknownRole -v` → PASS: mapping rejects unknown role with error, no SDK call made (fake asserts not invoked)
  - [ ] `go test ./internal/ai/ -run TestChatUsesConfiguredModel -v` → PASS: fake records model arg == client's chatModel
  - [ ] All pass with `OPENAI_API_KEY=` and blackholed proxy

  **QA Scenarios**:

  ```
  Scenario: Chat happy path through fake seam
    Tool: Bash
    Preconditions: tasks 1-2 complete, task 3 implemented
    Steps:
      1. Run: OPENAI_API_KEY= go test ./internal/ai/ -run 'TestChat' -v -count=1
      2. Assert: exit 0, all TestChat* cases PASS
      3. Run: OPENAI_API_KEY= HTTPS_PROXY=http://127.0.0.1:1 go test ./internal/ai/ -run 'TestChat' -count=1
      4. Assert: exit 0 (offline proven)
    Expected Result: chat logic verified end-to-end against fake — content returned, model passed, errors propagate
    Failure Indicators: FAIL lines; network attempt under blackhole proxy
    Evidence: .sisyphus/evidence/task-3-chat-tests.txt

  Scenario: Unknown role rejected before any API call
    Tool: Bash
    Preconditions: task 3 implemented
    Steps:
      1. Run: OPENAI_API_KEY= go test ./internal/ai/ -run 'TestChatUnknownRole' -v
      2. Assert: PASS — test asserts error non-nil AND fake's called-count == 0
    Expected Result: invalid input fails fast locally with clear error
    Failure Indicators: fake invoked despite bad role; nil error returned
    Evidence: .sisyphus/evidence/task-3-role-guard.txt
  ```

  **Evidence to Capture**:
  - [ ] `task-3-chat-tests.txt` — full -v output incl. blackhole run
  - [ ] `task-3-role-guard.txt` — unknown-role test output

  **Commit**: YES
  - Message: `feat(ai): add chat completion method`
  - Files: `internal/ai/chat.go`, `internal/ai/chat_test.go` (or merged into ai.go/ai_test.go if simpler)
  - Pre-commit: `go test ./internal/ai/`

- [x] 4. Embed method + tests via fake seam (TDD)

  **What to do**:
  - **RED first**: write `internal/ai/embed_test.go` — fake `api` returning canned vector/error; failing tests
  - `Embed(ctx context.Context, text string) ([]float64, error)` on `*Client` — delegates to seam with `c.embedModel`
  - SDK impl: `client.Embeddings.New(ctx, openai.EmbeddingNewParams{Input: openai.F[openai.EmbeddingNewParamsInputUnion](openai.EmbeddingNewParamsInputUnion{OfString: openai.String(text)}), Model: c.embedModel})`
  - Empty `embedding.Data` → descriptive error
  - Return `embedding.Data[0].Embedding, nil` — `[]float64` passthrough, no conversion
  - Empty `text` → pass through (API errors surface; no local guard)
  - **GREEN** → **REFACTOR**

  **Must NOT do**:
  - Batch embedding ([]string input) — single string only
  - Dimensions/encoding_format params — defaults only
  - `[]float32` conversion — `[]float64` passthrough
  - Vector normalization, caching, persistence

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: one method + union-input construction + fake tests, ~60 lines
  - **Skills**: [`test-driven-development`]
    - `test-driven-development`: CLAUDE.md TDD mandate
  - **Skills Evaluated but Omitted**:
    - `systematic-debugging`: no bug

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Task 3)
  - **Blocks**: Task 5
  - **Blocked By**: Task 2

  **References**:
  - External: `https://github.com/openai/openai-go/blob/main/_autodocs/core-apis.md` — verified embedding pattern incl. `openai.F[openai.EmbeddingNewParamsInputUnion]` union wrapping and `embedding.Data[0].Embedding` access; constants `EmbeddingModel3Small/3Large` exist but wrapper uses plain strings
  - `internal/ai/ai.go` (Task 2) — seam Embed signature to satisfy
  - **WHY []float64**: SDK returns float64 natively; converting to float32 is precision loss for zero benefit (Metis directive)

  **Acceptance Criteria**:
  - [ ] `go test ./internal/ai/ -run TestEmbed -v` → PASS: happy path returns canned vector via fake
  - [ ] `go test ./internal/ai/ -run TestEmbedEmptyData -v` → PASS: non-nil error when fake returns empty Data
  - [ ] `go test ./internal/ai/ -run TestEmbedError -v` → PASS: seam error propagates
  - [ ] `go test ./internal/ai/ -run TestEmbedUsesConfiguredModel -v` → PASS: fake records model arg == embedModel
  - [ ] All pass with `OPENAI_API_KEY=` and blackholed proxy

  **QA Scenarios**:

  ```
  Scenario: Embed happy path through fake seam
    Tool: Bash
    Preconditions: tasks 1-2 complete, task 4 implemented
    Steps:
      1. Run: OPENAI_API_KEY= go test ./internal/ai/ -run 'TestEmbed' -v -count=1
      2. Assert: exit 0, all TestEmbed* cases PASS
      3. Run: OPENAI_API_KEY= HTTPS_PROXY=http://127.0.0.1:1 go test ./internal/ai/ -run 'TestEmbed' -count=1
      4. Assert: exit 0 (offline proven)
    Expected Result: embed logic verified against fake — vector returned ([]float64, len>0), model passed, errors propagate
    Failure Indicators: FAIL lines; network under blackhole; wrong vector type
    Evidence: .sisyphus/evidence/task-4-embed-tests.txt

  Scenario: Empty Data handled with error
    Tool: Bash
    Preconditions: task 4 implemented
    Steps:
      1. Run: OPENAI_API_KEY= go test ./internal/ai/ -run 'TestEmbedEmptyData' -v
      2. Assert: PASS — error non-nil, no index-out-of-range panic
    Expected Result: graceful error, no panic
    Failure Indicators: panic in test output; nil error on empty Data
    Evidence: .sisyphus/evidence/task-4-empty-data.txt
  ```

  **Evidence to Capture**:
  - [ ] `task-4-embed-tests.txt` — full -v output incl. blackhole run
  - [ ] `task-4-empty-data.txt` — empty-Data test output

  **Commit**: YES
  - Message: `feat(ai): add embedding method`
  - Files: `internal/ai/embed.go`, `internal/ai/embed_test.go` (or merged)
  - Pre-commit: `go test ./internal/ai/`

- [x] 5. main.go demo + .env.example + makefile verification

  **What to do**:
  - Root `main.go`: `package main` — call `ai.New()`, on error print error to stderr + exit 1 (error already names missing var); on success print configured models to stdout, e.g. `fmt.Printf("chat model: %s\nembed model: %s\n", c.ChatModel(), c.EmbedModel())`
  - `.env.example` at repo root — 4 lines + comments: `OPENAI_BASE_URL=http://fedora:11434/v1`, `OPENAI_API_KEY=ollama`, `OPENAI_CHAT_MODEL=gemma3:1b`, `OPENAI_EMBED_MODEL=embeddinggemma`, each commented with purpose. Documentation only — code never reads this file
  - NO live API calls in demo — construction + getters only
  - Run `make build-local` → verify `./dist/bot` exists and is executable
  - Run with example env → verify stdout shows both model names, exit 0
  - Run with env unset → verify exit 1 + stderr names missing var
  - Only touch makefile if a target is actually broken (expected: no changes)

  **Must NOT do**:
  - Call Chat/Embed from main (no network in demo)
  - Load `.env.example` or any file from code — it is documentation for humans only
  - Add CLI flags, arg parsing, usage text
  - Rename makefile binary `bot` (zero churn; future Discord work owns that)
  - Add new makefile targets

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: ~15-line main.go + make verification runs
  - **Skills**: [`verification-before-completion`]
    - `verification-before-completion`: task's whole job is proving make targets work — run commands, confirm output, no success claims without evidence
  - **Skills Evaluated but Omitted**:
    - `test-driven-development`: trivial print-only main; YAGNI on tests per ponytail (getters already tested in Task 2)

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 4 (alone)
  - **Blocks**: None (final implementation task)
  - **Blocked By**: Tasks 2, 3, 4

  **References**:
  - `makefile:1-9` — targets under test: `build-local` (mkdir dist, go build -o ./dist/bot .), `go-run` (go run main.go)
  - `internal/ai/ai.go` (Task 2) — `New()` + getters used by demo
  - **WHY print-only demo**: proves wiring end-to-end without API key; gives `make go-run` a purpose until Discord bot exists

  **Acceptance Criteria**:
  - [ ] `make build-local` → exit 0; `test -x ./dist/bot` → exit 0
  - [ ] `OPENAI_BASE_URL=http://fedora:11434/v1 OPENAI_API_KEY=ollama OPENAI_CHAT_MODEL=gemma3:1b OPENAI_EMBED_MODEL=embeddinggemma make go-run` → exit 0; stdout contains `gemma3:1b` AND `embeddinggemma`
  - [ ] `env -u OPENAI_CHAT_MODEL -u OPENAI_EMBED_MODEL make go-run` → exit 1; stderr contains `OPENAI_CHAT_MODEL`
  - [ ] `grep -c 'Chat(\|Embed(' main.go` → 0 (no API calls in demo)
  - [ ] `test -f .env.example && grep -c 'OPENAI_' .env.example` → 4 vars documented
  - [ ] `git diff makefile` → empty (unless target was broken — document if changed)

  **QA Scenarios**:

  ```
  Scenario: Full build + run cycle with env configured
    Tool: Bash
    Preconditions: tasks 1-4 complete, repo root
    Steps:
      1. Run: make build-local
      2. Assert: exit 0
      3. Run: test -x ./dist/bot && echo "BINARY OK"
      4. Assert: "BINARY OK" printed
      5. Run: OPENAI_BASE_URL=http://fedora:11434/v1 OPENAI_API_KEY=ollama OPENAI_CHAT_MODEL=gemma3:1b OPENAI_EMBED_MODEL=embeddinggemma ./dist/bot
      6. Assert: exit 0, stdout contains "gemma3:1b" and "embeddinggemma"
      7. Run: OPENAI_BASE_URL=http://fedora:11434/v1 OPENAI_API_KEY=ollama OPENAI_CHAT_MODEL=gemma3:1b OPENAI_EMBED_MODEL=embeddinggemma make go-run
      8. Assert: exit 0, same model names in stdout
    Expected Result: binary + go run both print configured models, zero network calls made (construction only)
    Failure Indicators: build errors; missing binary; wrong/missing model names; non-zero exit; network attempt (would fail in sandbox without fedora reachable — demo must not call API)
    Evidence: .sisyphus/evidence/task-5-build-run.txt

  Scenario: Missing env fails loud with named variable
    Tool: Bash
    Preconditions: task 5 implemented
    Steps:
      1. Run: env -u OPENAI_CHAT_MODEL -u OPENAI_EMBED_MODEL -u OPENAI_API_KEY -u OPENAI_BASE_URL make go-run; echo "EXIT=$?"
      2. Assert: EXIT=1, stderr contains "OPENAI_CHAT_MODEL"
      3. Run: env OPENAI_CHAT_MODEL=only-chat -u OPENAI_EMBED_MODEL make go-run; echo "EXIT=$?"
      4. Assert: EXIT=1, stderr contains "OPENAI_EMBED_MODEL"
    Expected Result: each missing var produces exit 1 + error naming THAT var
    Failure Indicators: exit 0 with empty model; panic; generic error not naming the var
    Evidence: .sisyphus/evidence/task-5-fail-loud.txt
  ```

  **Evidence to Capture**:
  - [ ] `task-5-build-run.txt` — make build-local, binary check, ./dist/bot run, make go-run outputs
  - [ ] `task-5-fail-loud.txt` — both missing-env runs with EXIT codes

  **Commit**: YES
  - Message: `feat(main): add model config demo entrypoint`
  - Files: `main.go`, `.env.example`
  - Pre-commit: `make build-local && OPENAI_CHAT_MODEL=gemma3:1b OPENAI_EMBED_MODEL=embeddinggemma make go-run`

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.
>
> **Do NOT auto-proceed after verification. Wait for user's explicit approval.**
> Rejection or user feedback → fix → re-run → present again → wait for okay.

- [ ] F1. **Plan Compliance Audit** — `oracle`
  Read plan end-to-end. For each "Must Have": verify implementation exists (read file, run command). For each "Must NOT Have": grep codebase for forbidden patterns (discordgo, sqlite, streaming, retry, token stats, godotenv, extra deps in go.mod) — reject with file:line if found. Check evidence files exist in `.sisyphus/evidence/`.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [ ] F2. **Code Quality Review** — `unspecified-high`
  Run `go vet ./...` + `gofmt -l .` + `go test ./...`. Review all Go files for: unused imports, dead code, comment noise (comments on obvious lines — CLAUDE.md violation), exported symbols without need, generic names (data/result/temp). Confirm no AI-slop: no speculative abstractions, no over-validation beyond spec.
  Output: `Vet [PASS/FAIL] | Fmt [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [ ] F3. **Real Manual QA** — `unspecified-high`
  From clean state (`go clean -cache` optional): execute EVERY QA scenario from EVERY task — exact commands, exact env setups, capture all terminal output to `.sisyphus/evidence/final-qa/`. Verify: `go test ./...` passes fully offline with no env; `make build-local` produces executable `dist/bot`; example env (`gemma3:1b`/`embeddinggemma`) → demo prints both; missing env → exit 1 naming the var; `.env.example` documents all 4 vars.
  Output: `Scenarios [N/N pass] | Build [PASS/FAIL] | Fail-loud [PASS/FAIL] | VERDICT`

- [ ] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff (`git diff` vs branch start). Verify 1:1 — everything in spec built (no missing), nothing beyond spec built (no creep). Check every "Must NOT do" item. Flag unaccounted files (anything new not in plan's deliverables).
  Output: `Tasks [N/N compliant] | Creep [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

| Task | Message | Files | Pre-commit |
|------|---------|-------|-----------|
| 1 | `chore(build): init go module with openai-go v3` | `go.mod`, `go.sum` | `go build ./...` |
| 2 | `feat(ai): add client wrapper with model config resolution` | `internal/ai/*.go` | `go test ./internal/ai/` |
| 3 | `feat(ai): add chat completion method` | `internal/ai/*.go` | `go test ./internal/ai/` |
| 4 | `feat(ai): add embedding method` | `internal/ai/*.go` | `go test ./internal/ai/` |
| 5 | `feat(main): add model config demo entrypoint` | `main.go`, `.env.example` | `make build-local && OPENAI_CHAT_MODEL=gemma3:1b OPENAI_EMBED_MODEL=embeddinggemma make go-run` |

Branch: `gabe/core-ai-features` (current). Conventional commits per repo history style.

---

## Success Criteria

### Verification Commands
```bash
go test ./...
# Expected: PASS, ok github.com/Guuzzeji/teemysu/internal/ai, exit 0 — no network, no env needed

make build-local
# Expected: dist/bot created, executable, exit 0

OPENAI_BASE_URL=http://fedora:11434/v1 OPENAI_API_KEY=ollama OPENAI_CHAT_MODEL=gemma3:1b OPENAI_EMBED_MODEL=embeddinggemma make go-run
# Expected: stdout contains "gemma3:1b" and "embeddinggemma", exit 0

env -u OPENAI_CHAT_MODEL -u OPENAI_EMBED_MODEL make go-run
# Expected: exit 1, stderr names OPENAI_CHAT_MODEL

grep -rn 'gpt-\|gemma3\|embeddinggemma' internal/ai/*.go --include='!*_test.go'
# Expected: no matches — zero baked model literals

grep 'openai/openai-go/v3' go.mod
# Expected: match — only direct dependency
```

### Final Checklist
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] All tests pass offline
- [ ] F1–F4 all APPROVE
- [ ] User explicit okay received
