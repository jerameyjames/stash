# Task: Metadata Filtering in Memory Layer

**Status:** Proposed  
**Date:** 2026-04-23

---

## 1. Context

**Goal:** Add predicate-based filtering to Memory.Recall, allowing callers to find events by both semantic similarity AND structured metadata (e.g., "find discussions of high-severity bugs").

**Why:** Events carry metadata (severity, component, source, etc.). Semantic search alone is imprecise. Combining vector similarity with structured filtering separates signal from noise and enables Phase 2 features (contradiction detection, consolidation).

**What this is:** A new `RecallWhere()` method in Memory that accepts a predicate filter alongside the semantic query. Exposed through the CLI with simple filter syntax.

**What this is NOT:**
- Full SQL-style query language
- Nested metadata path support (Phase 2)
- Query optimizer or planner
- Complex boolean logic (only AND chains for now)

---

## 2. Boundaries

**In scope:**
- New method `Memory.RecallWhere(ctx, namespaces, query, filter, limit)` that combines vector search + predicate
- CLI flag `--where` on `events:search` accepting simplified filter syntax (e.g., `severity=high,component=gateway`)
- Tests covering: single filter, multiple filters (AND), no filter (backward compat)
- Documentation: Memory interface clarity

**Not in scope:**
- Full predicate AST exposure in CLI (JSON syntax is future)
- Nested metadata paths (e.g., `metadata.nested.field=value`)
- OR predicates or complex boolean logic
- Optimization passes
- Migration of existing CLI calls to new method (existing `Recall()` unchanged)

---

## 3. Design

### 3.1 Memory layer

**New method signature:**

```go
// RecallWhere retrieves events matching both semantic similarity and structured metadata.
// Combines vector search with optional predicate filtering.
// Returns at most limit events ordered by relevance (score descending).
// If filter is nil, behaves like Recall().
// limit must be > 0.
func (m *Memory) RecallWhere(
    ctx context.Context,
    namespaces []string,
    query string,
    filter *store.Predicate,
    limit int,
) ([]Event, error)
```

**Implementation detail:**
- Embed the query string
- Build a compound predicate: `AND(type=event, ...user_filter...)`
- Call `store.Search()` with the combined predicate
- Return results as Events (same as Recall)
- Backward compatible: existing `Recall()` unchanged

### 3.2 CLI integration

**New flag on `events:search`:**

```
--where string    Metadata filter in format: field=value,field>=value,...
                  Operators: =, !=, <, >, <=, >=
                  Fields are top-level metadata keys (no nesting)
                  Multiple filters are AND-ed together
```

**Parser:** Simple, hand-rolled. Accept format:
```
severity=high
severity=high,component=gateway
level>=3,source=api
```

**Converted to store.Predicate:**
```go
// severity=high,component=gateway becomes:
Predicate{
    And: []Predicate{
        {Field: "metadata.severity", Op: OpEq, Value: "high"},
        {Field: "metadata.component", Op: OpEq, Value: "gateway"},
    },
}
```

**Error handling:** If parser fails, return error. No silent drops.

### 3.3 Tests

- Unit test: `RecallWhere()` with single filter
- Unit test: `RecallWhere()` with multiple filters (AND)
- Unit test: `RecallWhere()` with nil filter (same as Recall)
- Integration test: create events with different metadata, search+filter, verify subset returned
- CLI test: `./stash events search "query" --where "severity=high" --namespace ns1` returns only high-severity

---

## 4. Implementation Notes

**File changes:**
- `internal/memory/memory.go` — add `RecallWhere()` method
- `internal/memory/memory_test.go` — add unit tests
- `internal/actions/events.go` (or equivalent CLI handler) — parse `--where`, call `RecallWhere()`
- CLI command help text updated

**Dependencies:**
- No new imports
- Uses existing `store.Predicate`, `store.OpEq`, etc.

**Backward compatibility:**
- Existing `Recall()` unchanged
- Existing CLI calls work unchanged
- New flag is optional

---

## 5. Acceptance Criteria

- [ ] `Memory.RecallWhere()` method exists and returns []Event
- [ ] Filter is applied: calling with `field=high` returns only matching records
- [ ] Multiple filters are AND-ed: `sev=high,comp=api` matches both conditions
- [ ] Nil filter behaves as Recall: no filtering applied
- [ ] CLI parses `--where` flag and passes predicate to RecallWhere
- [ ] CLI help updated with `--where` documentation
- [ ] All existing tests pass
- [ ] New tests cover: single filter, multiple filters, nil filter, invalid syntax
- [ ] Parser handles operators: =, !=, <, >, <=, >=
- [ ] Parser rejects malformed input with clear error message
- [ ] `go vet` and `staticcheck` pass with no warnings

---

## 6. Explicit Assumptions

- Top-level metadata keys only (no nested `metadata.foo.bar`)
- Operators are matched exactly (=, !=, <, >, <=, >=)
- Values are treated as strings; caller responsible for type correctness
- Parser is case-sensitive
- Invalid syntax in `--where` aborts the command (no partial application)

---

## 7. Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Parser complexity creeps | Keep it simple: hand-rolled, ~50 lines, no external grammar |
| Filter pushdown doesn't work | Test with real Postgres; verify predicate translates correctly |
| Performance: filtering after search vs. before | Use store's filter (pre-filter); let Postgres optimize |
| User confusion: "where" not SQL-like | Doc clearly: "simple metadata filter, not SQL" |

---

## 8. Definition of Done

- Code compiles without warnings
- All tests pass
- No new dependencies added
- AGENTS.md rules satisfied (one-way dependencies, no global state, etc.)
- Backward compatible (existing code unchanged)
- Ready for review: clean diff, single logical change

