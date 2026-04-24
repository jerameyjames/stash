# Task: Brain + MCP — Restructure into agent-first product

**Status:** Ready for Execution
**Date:** 2026-04-24

---

## 1. Context

- **Goal:** Replace the fragmented `internal/memory` + `internal/store` + HTTP server layers with a single `internal/brain` package exposed via MCP (Model Context Protocol) instead of REST.
- **Why:** The current codebase has multiple competing layers with no clear owner. CLI commands bypass memory and talk to store directly. HTTP handlers are inconsistent. There is no single authoritative agent-facing API. The product vision is: agents remember and recall — the brain handles everything else automatically. MCP is the natural protocol for agent tool integration (Claude Desktop, LLM frameworks, custom agents).

---

## 2. Boundaries

- **In Scope:**
  - Create `internal/brain/` with one file per operation (`memory_*.go`)
  - Move `internal/store/` → `internal/brain/store/` (import path change only, no logic change)
  - Create channel-driven background pipeline in `brain.Run()` (consolidation + relationship extraction, near-realtime via debounce)
  - Expose brain as MCP tools via `stash mcp serve` (SSE) and `stash mcp execute` (stdio)
  - Simplify CLI to 8 commands: remember, recall, forget, purge, reflect, contradict, env, mcp
  - Update bootstrap to expose only `bc.Brain` — no more `bc.Memory`, `bc.Store`, `bc.Embedder`, `bc.Reasoner`
  - Add `github.com/mark3labs/mcp-go` dependency
  - Remove `github.com/labstack/echo/v5` dependency

- **Non-Goals:**
  - No changes to `internal/embedder/` (stays as sibling)
  - No changes to `internal/reasoner/` (stays as sibling)
  - No changes to `internal/config/` (config vars unchanged)
  - No new config vars for pipeline (debounce window hardcoded at 10s)
  - No changes to store logic (PostgreSQL queries, migrations, schema)
  - No new features — restructuring only

- **Constraints:**
  - `internal/brain/store/` must not be imported by anything outside `internal/brain/`
  - CLI and MCP must never import `internal/brain/store`, `internal/embedder`, `internal/reasoner` directly
  - No goroutines spawned in `brain.New()` — only in `brain.Run()` (AGENTS.md rule 3.4)
  - No global state (AGENTS.md rule 3.3)
  - `go build ./cmd/cli/` must compile clean after every step

---

## 3. Approach & Review

- **Proposed Approach:**

  Build bottom-up: store move first, then brain package, then bootstrap, then CLI, then MCP. Each step compiles clean before moving to the next.

  **New directory structure:**
  ```
  internal/
  ├── brain/
  │   ├── brain.go                ← Brain struct, New(), Close()
  │   ├── pipeline.go             ← Run(), internal consolidation loop
  │   ├── memory_remember.go      ← Remember() + Memory type
  │   ├── memory_recall.go        ← Recall()
  │   ├── memory_forget.go        ← Forget() soft-delete
  │   ├── memory_purge.go         ← Purge() hard-delete
  │   ├── memory_reflect.go       ← Reflect() + Report type
  │   ├── memory_contradict.go    ← Contradict() + Contradiction type
  │   ├── errors.go
  │   └── store/                  ← moved from internal/store/
  │       ├── store.go
  │       └── postgres/
  ├── bootstrap/
  │   └── bootstrap.go            ← exposes only bc.Brain
  ├── embedder/                   ← unchanged
  └── reasoner/                   ← unchanged

  cmd/cli/
  ├── main.go                     ← 8 commands
  ├── cli_env.go                  ← unchanged
  ├── remember.go
  ├── recall.go
  ├── forget.go
  ├── purge.go
  ├── reflect.go
  ├── contradict.go
  └── mcp.go                      ← mcp serve + mcp execute
  ```

  **Brain public API:**
  ```go
  // brain.go
  type Brain struct { /* private fields */ }
  func New(s store.Store, e embedder.Embedder, r reasoner.Reasoner) (*Brain, error)
  func (b *Brain) Run(ctx context.Context) error
  func (b *Brain) Close() error

  // memory_remember.go
  type Memory struct {
      ID        string
      Namespace string
      Content   string
      Score     float32   // populated on Recall only
      CreatedAt time.Time
  }
  func (b *Brain) Remember(ctx context.Context, namespace, content string, metadata map[string]any) (string, error)

  // memory_recall.go
  func (b *Brain) Recall(ctx context.Context, namespace, query string, limit int) ([]Memory, error)

  // memory_forget.go
  // Forget finds the top memory matching the query and soft-deletes it.
  // No ID required — designed for natural language "forget about X" flows.
  func (b *Brain) Forget(ctx context.Context, namespace, query string) error

  // memory_purge.go
  func (b *Brain) Purge(ctx context.Context, id string) error

  // memory_reflect.go
  type Report struct { /* ... */ }
  func (b *Brain) Reflect(ctx context.Context, namespace string) (*Report, error)

  // memory_contradict.go
  type Contradiction struct { /* ... */ }
  func (b *Brain) Contradict(ctx context.Context, namespace string) ([]Contradiction, error)
  ```

  **Pipeline (internal, unexported):**
  ```go
  // pipeline.go
  func (b *Brain) Run(ctx context.Context) error
  // channel-driven debounce per namespace:
  //   Remember() → push namespace to pipelineCh (non-blocking)
  //   worker: collect namespaces, wait 10s idle, then:
  //     consolidate events → facts
  //     extract relationships from new facts
  //   context cancellation stops the worker cleanly
  ```

  **MCP tools (both transports):**

  Three tools, all with human-conversational descriptions so the LLM knows when to invoke them.
  `context` replaces `namespace` as the user-facing vocabulary — the LLM tracks the active context
  across turns and passes it automatically. The human never sees the word "namespace".

  ```
  remember
    description: Use when the user says things like:
                   "remember that...", "note that...", "keep in mind...",
                   "don't forget...", "I want you to know...",
                   or corrects something you previously knew.
    input:
      content:   string   (what to remember, required)
      context?:  string   (which context to store in, e.g. "work", "personal")
                          defaults to "general" if omitted
      metadata?: object   (optional structured data)
    output: { id: string }

  recall
    description: Use when the user asks things like:
                   "do you remember...", "what do you know about...",
                   "what have I told you about...", "remind me about...",
                   "have I mentioned...", or when you need context before answering.
    input:
      query:    string   (what to search for, in natural language, required)
      context?: string   (which context to search — omit to search across all)
      limit?:   number   (max results, default 10)
    output: Memory[]

  forget
    description: Use when the user says things like:
                   "forget that...", "that's no longer true",
                   "discard what I said about...", "ignore what I told you about...",
                   "remove the memory about..."
    input:
      about:    string   (what to forget, in natural language, required)
                         brain recalls top match then soft-deletes it — no ID needed
      context?: string   (which context to search in)
    output: { ok: bool }
  ```

  **Context switching — no extra tool needed:**
  When the user says "switch context to work" or "let's talk about personal stuff",
  the LLM simply starts passing `context="work"` or `context="personal"` to subsequent
  tool calls. The brain stores and retrieves from that namespace transparently.
  The human never manages namespaces — they just say what context they're in.

  **Bootstrap:**
  ```go
  type Context struct {
      Config *config.Config
      Brain  *brain.Brain
      Logger *slog.Logger
  }
  ```

- **Self-Critique:**
  - Moving `internal/store/` changes import paths. Manageable because we rewrite the memory layer entirely — no partial migration.
  - Pipeline debounce is new code with concurrency. Must handle context cancellation carefully. Keep it simple: one goroutine, one channel, one ticker.
  - `mcp-go` SSE transport API needs verification before writing `mcp.go`. Check library examples first.
  - Report and Contradiction types can be lifted directly from `internal/memory/types.go` — no reimplementation.

- **Decision:** Build bottom-up. Compile after every step. Do not proceed to the next step if the build is broken.

- **Explicit Assumptions:**
  - `github.com/mark3labs/mcp-go` supports both SSE and stdio transports.
  - The existing PostgreSQL store logic (`internal/store/postgres/`) requires no changes — only the import path changes.
  - Consolidation and relationship extraction logic from `internal/memory/memory.go` moves into `internal/brain/pipeline.go` unchanged.
  - `internal/bootstrap/bootstrap.go` drops `bc.Store`, `bc.Embedder`, `bc.Reasoner` as public fields — only `bc.Brain` is exposed.
  - `stash mcp serve` uses `--host` and `--port` flags (same as old `stash server`).
  - `stash mcp execute` uses no flags — reads from stdin, writes to stdout.
  - `context` in MCP tool parameters maps 1:1 to `namespace` in the brain API. The translation happens in `mcp.go` — brain never sees the word "context".
  - `forget` on the brain does: Recall(namespace, query, 1) → take top result → Store.Delete(id). If nothing found, return a sentinel error (not a hard failure).
  - MCP tool descriptions must be written for the LLM, not the developer. They explain *when* to call the tool, not *what* it does technically.

---

## 4. Next-Step Handoff

- **Implementation Notes:**
  - Start with step 1 (store move) and verify `go build` before continuing.
  - `internal/brain/pipeline.go` contains `Run()` plus all unexported helpers for consolidation. These are moved verbatim from `internal/memory/memory.go` (`ConsolidateRecent`, `ConsolidateRelationships`, `clusterRecordsBySimilarity`, `queryRecentEventRecords`).
  - `memory_reflect.go` and `memory_contradict.go` lift types + logic verbatim from `internal/memory/types.go` and `internal/memory/memory.go`.
  - For `mcp.go`: read `mcp-go` README before writing. Verify `server.NewSSEServer()` and `server.NewStdioServer()` API signatures.
  - `bootstrap.go` must call `brain.Run(ctx)` as a goroutine before returning — or the caller (`serverCmd`, `mcpCmd`) must call it explicitly. Prefer caller-explicit to keep bootstrap simple (no hidden goroutines).

- **Files / Areas Likely Affected:**
  - Created: `internal/brain/` (all files)
  - Moved: `internal/store/` → `internal/brain/store/`
  - Updated: `internal/bootstrap/bootstrap.go`
  - Updated: `cmd/cli/main.go`
  - Rewritten: `cmd/cli/remember.go`, `cmd/cli/recall.go`, `cmd/cli/purge.go`
  - Created: `cmd/cli/forget.go`, `cmd/cli/reflect.go`, `cmd/cli/contradict.go`, `cmd/cli/mcp.go`
  - Deleted: `internal/memory/`, `internal/handlers/`, `cmd/cli/server.go`, `cmd/cli/context.go`, `cmd/cli/delete.go`, `cmd/cli/list.go`, `cmd/cli/filter.go`, `cmd/cli/facts_*.go`
  - Updated: `go.mod`, `go.sum`

- **Risks / Watchouts:**
  - Import path `internal/store` → `internal/brain/store` must be updated in every file that references it. Grep before and after.
  - Pipeline goroutine must stop cleanly when context is cancelled. Leaking goroutines will cause issues in tests and graceful shutdown.
  - `mcp-go` SSE transport requires an HTTP listener. Verify it doesn't conflict with any existing port usage.
  - `bootstrap.go` currently exposes `bc.Store` which is used by `cmd/cli/list.go`, `cmd/cli/delete.go`, `cmd/cli/purge.go`. These files are being deleted/rewritten — verify no dangling references before removing `bc.Store`.

- **Verification Plan:**
  1. `go build ./cmd/cli/` — must compile clean
  2. `go vet ./...` — must pass
  3. `stash env` — must print config without error
  4. `stash remember "Alice works at TechCorp"` — must return `{"id": "..."}` 
  5. `stash recall "Alice"` — must return relevant memories
  6. `stash forget "Alice works at TechCorp"` — must find top match and soft-delete without error
  7. `stash purge <id>` — must hard-delete without error
  8. `stash reflect` — must return report without error
  9. `stash contradict` — must return without error
  10. `stash mcp execute` — must start, list `remember`, `recall`, and `forget` as available tools
  11. `stash mcp serve --port 8080` — must start SSE server, respond to connections
  12. `grep -r "internal/memory" --include="*.go"` — must return empty
  13. `grep -r "internal/store" --include="*.go"` — must return only `internal/brain/store/` references
  14. `grep -r "labstack/echo" --include="*.go"` — must return empty

- **Acceptance Criteria:**
  - `internal/memory/` does not exist
  - `internal/store/` does not exist (moved to `internal/brain/store/`)
  - `internal/handlers/` does not exist
  - `bc.Brain` is the only brain-related field on bootstrap.Context
  - `cmd/cli/` contains exactly 8 command files + `main.go` + `cli_env.go`
  - `stash mcp execute` lists exactly 3 tools: `remember`, `recall`, `forget`
  - `stash mcp serve` starts without error
  - `go build ./cmd/cli/` compiles with zero errors and zero warnings
  - `go vet ./...` passes clean

---

## 5. Execution Steps

- [x] 1. Move `internal/store/` → `internal/brain/store/`, fix all import paths, verify `go build`
- [x] 2. Create `internal/brain/errors.go`
- [x] 3. Create `internal/brain/brain.go` (Brain struct + New() + Close())
- [x] 4. Create `internal/brain/memory_remember.go` (Remember() + Memory type)
- [x] 5. Create `internal/brain/memory_recall.go` (Recall())
- [x] 6. Create `internal/brain/memory_forget.go` (Forget() soft-delete)
- [x] 7. Create `internal/brain/memory_purge.go` (Purge() hard-delete)
- [x] 8. Create `internal/brain/memory_reflect.go` (Reflect() + Report type)
- [x] 9. Create `internal/brain/memory_contradict.go` (Contradict() + Contradiction type)
- [x] 10. Create `internal/brain/pipeline.go` (Run() + internal consolidation + relationship extraction)
- [x] 11. Update `internal/bootstrap/bootstrap.go` (wire brain.New(), expose only bc.Brain)
- [x] 12. Delete `internal/memory/`
- [x] 13. Verify `go build` clean
- [x] 14. Rewrite `cmd/cli/main.go` (new 8-command tree)
- [x] 15. Rewrite `cmd/cli/remember.go` → calls bc.Brain.Remember()
- [x] 16. Rewrite `cmd/cli/recall.go` → calls bc.Brain.Recall()
- [x] 17. Create `cmd/cli/forget.go` → calls bc.Brain.Forget(namespace, query) — takes natural language, not ID
- [x] 18. Rewrite `cmd/cli/purge.go` → calls bc.Brain.Purge()
- [x] 19. Create `cmd/cli/reflect.go` → calls bc.Brain.Reflect()
- [x] 20. Create `cmd/cli/contradict.go` → calls bc.Brain.Contradict()
- [x] 21. Delete old CLI files (server.go, context.go, delete.go, list.go, filter.go, facts_*.go)
- [x] 22. Delete `internal/handlers/`
- [x] 23. Verify `go build` clean
- [x] 24. Add `github.com/mark3labs/mcp-go`, remove `github.com/labstack/echo/v5` from go.mod
- [x] 25. Create `cmd/cli/mcp.go` (mcp serve + mcp execute, both using bc.Brain)
- [x] 26. `go mod tidy`
- [x] 27. Final `go build` + `go vet` — must be clean
- [x] 28. Run verification plan (steps 1-14 above)
- [x] 29. Commit

---

## 6. Progress Notes

- [2026-04-24] Task created. Ready for execution.
- [2026-04-24] Updated MCP tools: added `forget` as 3rd tool, renamed `namespace` → `context` in tool parameters, added human-conversational descriptions, clarified forget-by-query semantics (no ID needed).
- [2026-04-24] All 29 steps completed. Build compiles clean. go vet passes.

---

## 7. Outcome

- **Final Result:** Completed
- **Commit:** 9a5f390
- **Net change:** +448 lines, -3241 lines (~2800 lines deleted)
- **Build status:** `go build ./cmd/cli/` compiles clean, `go vet ./...` passes

**What was built:**
- `internal/brain/` — 9 files, single entry point for all memory operations
- `internal/brain/store/` — moved from `internal/store/`
- `cmd/cli/mcp.go` — MCP server with 3 tools (remember, recall, forget)
- Simplified CLI — 8 commands instead of 16

**What was deleted:**
- `internal/memory/` — 1900+ lines
- `internal/handlers/` — HTTP handlers
- `internal/actions/` — action wrappers
- `cmd/cli/server.go` — Echo HTTP server
- `cmd/cli/context.go`, `delete.go`, `filter.go`, `list.go` — old commands
- `cmd/cli/facts_*.go` — 8 fact-related commands
- `github.com/labstack/echo/v5` dependency

**Verification:**
- `grep -r "internal/memory" --include="*.go"` → empty
- `grep -r '"github.com/alash3al/stash/internal/store"' --include="*.go"` → empty
- `grep -r 'labstack/echo' --include="*.go"` → empty
- `stash --help` shows 8 commands + mcp subcommands
