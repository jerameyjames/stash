# Task: Event Decay / TTL

**Status:** Proposed  
**Date:** 2026-04-23

---

## 1. Context

**Goal:** Events expire and fade from memory unless actively reinforced. Add optional time-to-live (TTL) to events so memory behaves like actual memory — impermanent by default.

**Why:** Real memory decays. A conversation from 6 months ago shouldn't resurface in every semantic search. TTL adds realism and prevents event bloat. Cleanup can be batched, no background goroutines.

**What this is:** Optional `expires_at` field on events. New `Remember()` variant with TTL. Automatic filtering of expired events on read. Explicit cleanup method for hard-delete.

**What this is NOT:**
- Background cleanup goroutine (caller runs cleanup explicitly)
- Decay functions or time-weighted relevance (Phase 2)
- Automatic re-embedding on expiry
- Event touch/refresh (Phase 2)

---

## 2. Boundaries

**In scope:**
- New `RememberWithTTL(ctx, namespace, content, ttl, metadata)` method
- Auto-filter expired events in `Recall()` and `RecallWhere()`
- New `PurgeExpired(ctx, namespace)` method for cleanup
- Events without TTL live forever (backward compatible)
- Tests: expiry, filtering, cleanup, backward compat

**Not in scope:**
- Background cleanup tasks
- Decay weight on relevance
- Refresh/touch mechanism
- Global default TTL (always explicit)
- Partial expiry (events either expire or don't)

---

## 3. Design

### 3.1 Event type extension

**Update `internal/memory/types.go` `Event` struct:**

```go
type Event struct {
    ID        string
    Namespace string
    Content   string
    Timestamp time.Time
    ExpiresAt *time.Time        // NEW: nil = forever, non-nil = expiration
    Metadata  map[string]any
    Score     float32
}
```

In storage, ExpiresAt lives in `_memory.expires_at` (RFC3339 string or nil).

### 3.2 Memory methods

**New method:**

```go
// RememberWithTTL stores an event that expires after ttl duration.
// Generates UUID and embedding. Returns event ID.
// ttl must be > 0.
// metadata must not start with "_memory".
func (m *Memory) RememberWithTTL(
    ctx context.Context,
    namespace string,
    content string,
    ttl time.Duration,
    metadata map[string]any,
) (string, error)
```

**Updated methods (backward compatible):**

```go
// Recall now filters out expired events automatically.
// If an event has passed its ExpiresAt, it's not returned.
// Non-expiring events (ExpiresAt = nil) are always included.
func (m *Memory) Recall(
    ctx context.Context,
    namespaces []string,
    query string,
    limit int,
) ([]Event, error)  // UNCHANGED signature, behavior updated

// RecallWhere also filters expired (already done by Recall).
// No changes needed.
```

**New cleanup method:**

```go
// PurgeExpired hard-deletes all expired events in the given namespaces.
// Returns count of deleted records.
// Non-expiring events are never touched.
// Safe to call frequently; idempotent.
func (m *Memory) PurgeExpired(
    ctx context.Context,
    namespaces []string,
) (int64, error)
```

### 3.3 Storage format

**Event with TTL:**

```go
store.Record{
    ID:        "event-uuid",
    Namespace: namespace,
    Content:   content,
    Vectors:   {...},
    Metadata: {
        "_memory": {
            "type":       "event",
            "content":    content,
            "timestamp":  "2026-04-23T20:31:33Z",
            "expires_at": "2026-04-30T20:31:33Z",  // NEW
        },
        ...callerMetadata...
    },
}
```

**Event without TTL (backward compatible):**

```go
"_memory": {
    "type":       "event",
    "content":    content,
    "timestamp":  "2026-04-23T20:31:33Z",
    // no expires_at field
}
```

### 3.4 Filtering logic

**In `Recall()` and `RecallWhere()`:**

After retrieving results from store, filter out any event where:
```
ExpiresAt != nil AND ExpiresAt < now.UTC()
```

This is done in-memory (not in the store predicate) for simplicity.

**In `PurgeExpired()`:**

Build a predicate:
```go
Predicate{
    And: []Predicate{
        {Field: "metadata._memory.type", Op: OpEq, Value: "event"},
        {Field: "metadata._memory.expires_at", Op: OpExists, Value: true},
    },
}
```

Then iterate results, check `expires_at < now`, and call `store.Delete()` for each.

---

## 4. Implementation Notes

**File changes:**
- `internal/memory/types.go` — add `ExpiresAt *time.Time` to Event
- `internal/memory/memory.go` — add `RememberWithTTL()`, `PurgeExpired()`; update `Recall()`
- `internal/memory/memory_test.go` — tests

**Helper functions:**
```go
// Extract expires_at from _memory metadata, parse, return *time.Time or nil
func extractExpiresAt(memMeta map[string]any) *time.Time

// Check if an event is expired
func isExpired(expiresAt *time.Time) bool
```

**Backward compatibility:**
- `Remember()` unchanged; events have no TTL
- `Recall()` always filters expired (transparent to caller)
- Existing tests still pass
- Existing code unaffected

---

## 5. Acceptance Criteria

- [ ] `Event.ExpiresAt *time.Time` field added
- [ ] `Memory.RememberWithTTL()` exists and returns event ID
- [ ] Events created with TTL store `expires_at` in metadata
- [ ] Events created without TTL (via `Remember()`) have no `expires_at`
- [ ] `Recall()` automatically filters expired events
- [ ] `Recall()` still returns non-expiring events
- [ ] Expired events are invisible to search (not returned)
- [ ] `PurgeExpired()` hard-deletes all expired events in namespace
- [ ] `PurgeExpired()` returns count of deleted
- [ ] Non-expiring events are never auto-deleted
- [ ] Backward compatible: existing `Remember()` works unchanged
- [ ] Tests: create with TTL, verify hidden after expiry, cleanup removes
- [ ] Tests: TTL=0 or negative rejected (ErrInvalidTTL)
- [ ] Tests: multiple namespaces, verify isolation
- [ ] `go vet` and `staticcheck` pass

---

## 6. Explicit Assumptions

- Time comparisons use UTC
- TTL is always > 0 (no permanent events via TTL parameter)
- ExpiresAt is stored as RFC3339 string in metadata
- Cleanup is explicitly requested (no background goroutines)
- Expired events are soft-deleted, not purged (caller uses `PurgeExpired()` for hard delete)
- Single namespace per `PurgeExpired()` call (or accept []string)

---

## 7. Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Expired events accumulate (soft-deleted, not hard-deleted) | Document: call `PurgeExpired()` periodically |
| Clock skew (server time vs. client creation time) | Use server time (now.UTC()) for all expiry calculations |
| Partial expiry (some replicas expired, others not) | Not applicable: single Postgres instance |
| Performance: filtering in-memory after retrieve | Acceptable for typical recall limits (10-100 results); optimize later |

---

## 8. Definition of Done

- Code compiles without warnings
- All tests pass
- Backward compatible (existing code unchanged)
- TTL filtering transparent to caller
- Cleanup method works and is documented
- Ready for review

