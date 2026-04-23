# Task: Event Relationships as Records

**Status:** Proposed  
**Date:** 2026-04-23

---

## 1. Context

**Goal:** Enable callers to express semantic relationships between events ("event X contradicts Y", "X caused Y") without extending the Store interface. Store relationships as first-class records.

**Why:** Phase 2 features (contradiction detection, consolidation, deduplication) depend on knowing how events relate. Relationships must be queryable, persistent, and invisible to normal event recall.

**What this is:** A new `Relation` type in Memory. Methods to create links between events and retrieve related events. All backed by store records with `_memory.type=relationship`.

**What this is NOT:**
- Graph traversal or transitive closure (Phase 2)
- Automatic contradiction detection (Phase 2)
- Bidirectional links (relationships are directional: from → to)
- Querying relationship chains (single-hop only)

---

## 2. Boundaries

**In scope:**
- New `Relation` type (from_id, to_id, relation_type, optional metadata)
- `Memory.LinkEvents(ctx, namespace, fromID, toID, relationType, metadata)` — create relationship
- `Memory.FindRelated(ctx, namespace, eventID, relationType)` — find all events related to an event
- Support 4 relation types: `contradicts`, `caused_by`, `similar_to`, `references` (extensible)
- Relationships stored as records, invisible to `Recall()` (filtered by type predicate)
- Tests: link, retrieve, verify namespace isolation

**Not in scope:**
- Reverse lookup (`FindRelatedTo` — Phase 2)
- Transitive queries (`FindRelatedVia` — Phase 2)
- Automatic link creation (Phase 2)
- Bidirectional relationships
- Weighted relationships (Phase 2)
- Relationship metadata indexed for filtering (Phase 2)

---

## 3. Design

### 3.1 Types

**New type in `internal/memory/types.go`:**

```go
// Relation represents a directed semantic link between two events.
// Stored as a store.Record with _memory.type = "relationship".
type Relation struct {
    ID             string                 // generated UUID
    Namespace      string                 // same namespace as linked events
    FromEventID    string                 // source event
    ToEventID      string                 // target event
    RelationType   string                 // e.g., "contradicts", "caused_by"
    Metadata       map[string]any         // optional caller metadata
    CreatedAt      time.Time
}

// Supported relation types (extensible)
const (
    RelationTypeContradicts = "contradicts"   // A contradicts B
    RelationTypeCausedBy    = "caused_by"     // A caused B
    RelationTypeSimilarTo   = "similar_to"    // A is similar to B
    RelationTypeReferences  = "references"    // A references B
)
```

### 3.2 Memory methods

**New methods in `internal/memory/memory.go`:**

```go
// LinkEvents creates a directed relationship from fromID to toID.
// Returns the relation ID.
// Both events must exist in the namespace (validated).
// relation_type must be one of the known types.
// metadata must not contain "_memory" keys.
func (m *Memory) LinkEvents(
    ctx context.Context,
    namespace string,
    fromID string,
    toID string,
    relationType string,
    metadata map[string]any,
) (string, error)

// FindRelated retrieves all events that are related to eventID by relationType.
// Returns events that satisfy: exists relation where from_event=eventID AND type=relationType.
// Returns empty slice if no relations found.
func (m *Memory) FindRelated(
    ctx context.Context,
    namespace string,
    eventID string,
    relationType string,
) ([]Event, error)
```

### 3.3 Storage format

**Relationships stored as records:**

```go
store.Record{
    ID:        "rel-uuid",
    Namespace: namespace,
    Content:   "",  // empty; relationships are metadata-only
    Metadata: {
        "_memory": {
            "type":            "relationship",
            "from_event_id":   "event-id-1",
            "to_event_id":     "event-id-2",
            "relation_type":   "contradicts",
            "created_at":      "2026-04-23T...",
        },
        // optional caller metadata
        "category": "detected_by_phase2",
    },
}
```

**Predicates to query:**
- Find all relationships: `_memory.type = "relationship"`
- Find contradictions: `AND(_memory.type = "relationship", _memory.relation_type = "contradicts", _memory.from_event_id = eventID)`

### 3.4 Validation

- Both fromID and toID must exist (call `store.Get()` to verify)
- relationType must be recognized or explicitly allowed
- metadata must not start with "_memory"
- fromID != toID (no self-links)

---

## 4. Implementation Notes

**File changes:**
- `internal/memory/types.go` — add `Relation` type, constants
- `internal/memory/memory.go` — add `LinkEvents()`, `FindRelated()` methods
- `internal/memory/memory_test.go` — tests for linking and retrieval

**Helper functions (internal):**
```go
func relationToRecord(rel Relation) store.Record
func recordToRelation(r store.Record) (Relation, error)
```

**Backward compatibility:**
- No changes to existing Memory methods
- Existing code unaffected

---

## 5. Acceptance Criteria

- [ ] `Relation` type exists with all required fields
- [ ] Relation type constants defined: `RelationTypeContradicts`, etc.
- [ ] `Memory.LinkEvents()` creates relationship and returns ID
- [ ] `Memory.FindRelated()` retrieves all events related by type
- [ ] `store.Get()` confirms both events exist before linking
- [ ] Relationships are stored as records with `_memory.type=relationship`
- [ ] Relationships invisible to `Recall()` (filtered out by type predicate)
- [ ] Namespace isolation verified: relations in ns1 don't appear in ns2
- [ ] Validation: rejects self-links, non-existent events, invalid metadata
- [ ] Tests: link creation, retrieval, namespace isolation, error cases
- [ ] `go vet` and `staticcheck` pass
- [ ] No new dependencies

---

## 6. Explicit Assumptions

- Relationships are directional: LinkEvents(A, B) ≠ LinkEvents(B, A)
- RelationType is case-sensitive and exact-match
- Metadata keys can be any string except those starting with "_memory"
- FindRelated returns Events, not Relation objects (simpler for callers)
- No automatic backlink creation

---

## 7. Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Orphaned relationships (event deleted, relation remains) | Document: cleanup is caller's responsibility (Phase 2) |
| Self-links allowed | Validate: reject fromID == toID |
| Relationship proliferation | Phase 2 adds cleanup/consolidation; ok for now |
| Query performance (many relationships per event) | Predicate indexed by _memory fields; relies on Postgres GIN |

---

## 8. Definition of Done

- Code compiles without warnings
- All tests pass
- Backward compatible (no changes to existing Memory API)
- Relationships properly scoped to namespace
- Error handling clear (ErrEventNotFound, ErrInvalidRelationType, etc.)
- Ready for review

