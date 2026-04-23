# Task: Batch Operations

**Status:** Proposed  
**Date:** 2026-04-23

---

## 1. Context

**Goal:** Support bulk create/delete for real workflows. CLI-driven batch imports from files (JSONL format). Programmatic batch operations for high-volume use cases.

**Why:** CLI remeber/recall one-by-one doesn't scale. Need to import conversations, backfill history, bulk-delete by criteria. Store already supports `PutMany()` — wire it up.

**What this is:** New `RememberMany()` in Memory. CLI `events:import` command reading JSONL. Atomic batch operations.

**What this is NOT:**
- CSV parsing or format conversion
- Streaming from databases or APIs
- Batch update/patch operations
- Progress reporting or cancellation mid-batch

---

## 2. Boundaries

**In scope:**
- `Memory.RememberMany(ctx, namespace, events []BulkRemember)` using store.PutMany
- CLI command `events:import` reading JSONL from stdin/file
- JSONL format: one JSON object per line with `content`, optional `metadata`, optional `ttl`
- Hard cap: 10,000 events per batch
- Atomic: all-or-nothing semantics via store.PutMany
- Tests: import, verify stored, searchable, errors on malformed

**Not in scope:**
- CSV, TSV, or other formats (JSONL only)
- Streaming ingestion (all in-memory)
- Progress bars or status updates
- Batch update or delete operations
- Format conversion tools (caller's responsibility)
- Dedupe or merging (Phase 2)

---

## 3. Design

### 3.1 Types

**New type in `internal/memory/types.go`:**

```go
// BulkRemember represents a single event for batch import.
// Minimal structure: just content, optional metadata and TTL.
type BulkRemember struct {
    Content  string         // required, non-empty
    Metadata map[string]any // optional caller metadata
    TTL      *time.Duration // optional; nil = no expiry
}
```

### 3.2 Memory method

**New method in `internal/memory/memory.go`:**

```go
// RememberMany stores multiple events atomically using store.PutMany.
// Generates UUIDs and embeddings for each.
// Returns count of stored events.
// Errors if any event is invalid (empty content, bad metadata).
// Errors if count > 10,000.
// All-or-nothing: if any embedding fails, entire batch is rolled back.
func (m *Memory) RememberMany(
    ctx context.Context,
    namespace string,
    events []BulkRemember,
) (count int, err error)
```

**Implementation:**
1. Validate all events (non-empty content, metadata ok, count <= 10k)
2. Embed all contents in parallel (fan out, fan in)
3. Build store.Record slice
4. Call store.PutMany()
5. Return count

### 3.3 CLI command

**New command: `events:import`**

```bash
./stash events:import --namespace proj-1 < events.jsonl
# or
./stash events:import --namespace proj-1 --file events.jsonl

# Output:
# {
#   "success": true,
#   "count": 1000,
#   "errors": 0,
#   "namespace": "proj-1"
# }
```

**JSONL format (one JSON object per line):**

```jsonl
{"content": "met Alice at KubeCon"}
{"content": "fixed auth bug", "metadata": {"severity": "high", "component": "api"}}
{"content": "investigated latency", "ttl": "24h"}
{"content": "conversation notes", "metadata": {"source": "slack"}, "ttl": "7d"}
```

**Parser:**
- Read line by line
- Unmarshal each as JSON
- Validate (content non-empty, metadata keys ok)
- Collect into []BulkRemember
- Pass to RememberMany()
- Report success/failure with counts

**Error handling:**
- Invalid JSON: report line number and error, abort
- Missing content: report line number, abort
- Bad TTL format: report line number, abort
- Batch exceeds 10k: report limit, suggest chunking

### 3.4 TTL string parsing

**Accept human-readable durations:**

```go
// Parse CLI ttl string: "1h", "24h", "7d", "30m", "1h30m"
// Use Go's time.ParseDuration()
// Return *time.Duration or error
```

Examples:
- `"1h"` → 1 hour
- `"24h"` → 24 hours
- `"7d"` → invalid (ParseDuration doesn't support "d"), so handle with custom logic or document as hours only

**Decision:** Accept `time.ParseDuration` format only (ns, us, ms, s, m, h). If callers want "7d", they write "168h". Simplicity.

---

## 4. Implementation Notes

**File changes:**
- `internal/memory/types.go` — add `BulkRemember` type
- `internal/memory/memory.go` — add `RememberMany()` method
- `internal/memory/memory_test.go` — tests
- `internal/actions/events.go` (or CLI handler) — add `import` command

**Embedding parallelism:**
```go
// Embed all contents in parallel
// Use sync.Errgroup or simple goroutines with buffered channel
// If any embed fails, cancel context and return error
```

**Validation:**
```go
func validateBulkRemember(br BulkRemember) error
```

**Backward compatibility:**
- No changes to existing Memory methods
- New method only
- Existing code unaffected

---

## 5. Acceptance Criteria

- [ ] `BulkRemember` type defined with Content, Metadata, TTL
- [ ] `Memory.RememberMany()` exists and returns (int, error)
- [ ] All events embedded (vectorized) before storage
- [ ] store.PutMany() used for atomic batch insert
- [ ] Hard cap: rejects > 10,000 events with ErrBatchTooLarge
- [ ] CLI command `events:import` exists
- [ ] Reads JSONL from stdin or --file flag
- [ ] Each line parsed as JSON, validated, collected
- [ ] Success response: {success, count, errors, namespace}
- [ ] Error handling: malformed JSON reports line number and aborts
- [ ] TTL parsing accepts time.ParseDuration format (h, m, s, etc.)
- [ ] Atomicity: if any embed fails, entire batch rejected (no partial writes)
- [ ] Tests: import 100 events, verify all stored, searchable
- [ ] Tests: 10,001 events rejected with error
- [ ] Tests: malformed JSON aborts with line number
- [ ] Tests: TTL parsed correctly and events expire
- [ ] `go vet` and `staticcheck` pass
- [ ] No new dependencies

---

## 6. Explicit Assumptions

- All events in a batch have the same namespace (single batch per namespace call)
- TTL is optional per event (nil = no expiry)
- Metadata keys are caller-validated (no "_memory" enforcement in this task)
- Embedding API is stateless and idempotent (safe to retry on failure)
- Events are stored in order (no shuffling)
- Batch size limit of 10k is reasonable for typical workflows

---

## 7. Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Embedding API timeout on large batches | Parallelize embeds; caller sets context timeout |
| Memory exhaustion (10k vectors in-memory) | 10k limit is reasonable; optimize later if needed |
| Partial failure mid-batch | Use errgroup + early return on first error (all-or-nothing) |
| JSONL format fragility | Document expected format clearly; strict parsing with line numbers |
| TTL parsing confusion | Document supported format (Go duration syntax); examples in help |

---

## 8. Definition of Done

- Code compiles without warnings
- All tests pass
- Batch operations atomic (all-or-nothing)
- CLI reads JSONL, reports clear errors
- Hard cap enforced and documented
- Backward compatible (no changes to existing API)
- Ready for review

