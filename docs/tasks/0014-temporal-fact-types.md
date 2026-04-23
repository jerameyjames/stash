# Task: Temporal Fact Types

**Status:** Completed  
**Date:** 2026-04-24

---

## 1. Context

**Goal:** Add semantic layer to Phase 2 memory. Facts now have temporal types that determine retrieval and update semantics.

**Why:** Raw facts are just "true/false at a time". Semantic facts need context: is this always true? is it current state? is it a snapshot? Different types require different retrieval and update strategies for Phase 3+ reasoning.

**What this is:**
- Three FactType constants: "atemporal", "state", "point-in-time"
- Fact.Type field to record temporal semantics
- Query methods per type: GetAtemporalFacts(), GetStateFactsFor(), GetPointInTimeFacts()
- ConsolidateRecent assigns type to synthesized facts
- CLI query command to filter facts by type

**What this is NOT:**
- Semantic consolidation or fact merging (Phase 3 Task 0016)
- Entity relationships or graph layer (Phase 3 Task 0015)
- Confidence-ranked retrieval (Phase 3 Task 0017)
- Automatic type detection (users/LLM infer types from fact content)

---

## 2. Boundaries

**In scope:**
- Add FactType constants (atemporal, state, point-in-time)
- Add Fact.Type field
- Update FactFromRecord to extract type from metadata
- Update ConsolidateRecent to set default type (state)
- Add query methods:
  - QueryFactsByType(ctx, namespace, factType) []Fact
  - GetAtemporalFacts(ctx, namespace) []Fact
  - GetStateFactsFor(ctx, namespace, entity) []Fact
  - GetPointInTimeFacts(ctx, namespace) []Fact
- CLI command: `stash facts query --namespace=X --type=state`
- Unit tests for all query methods
- Backward compatibility: facts without type default to "state"

**Not in scope:**
- Semantic consolidation or fact merging
- Entity relationships
- Ranked retrieval
- Type inference from LLM
- Time decay or expiration based on type

---

## 3. Approach & Review

**Temporal Types:**

1. **Atemporal** (`"atemporal"`)
   - Example: "Mohamed was born in Egypt"
   - ValidFrom = creation time
   - ValidUntil = nil forever (always true)
   - Never expires or changes
   - Retrieval: search across all facts, no time filtering

2. **State** (`"state"`)
   - Example: "Mohamed is working on Stash"
   - ValidFrom = creation time
   - ValidUntil = nil if ongoing, set when superseded
   - Current belief until contradicted
   - Retrieval: filter for ValidUntil=nil (only current state)

3. **Point-in-time** (`"point-in-time"`)
   - Example: "Mohamed deployed v0.1 on April 18, 2026"
   - ValidFrom = moment of event
   - ValidUntil = same as ValidFrom (snapshot)
   - Immutable historical record
   - Retrieval: search by timestamp

**Query Methods:**
- `QueryFactsByType`: Generic filter by type
- `GetAtemporalFacts`: All atemporal facts (always true)
- `GetStateFactsFor`: Current state facts for entity
- `GetPointInTimeFacts`: All snapshots

**Default Behavior:**
- ConsolidateRecent assigns `type="state"` to synthesized facts
- FactFromRecord defaults to "state" if type not in metadata
- Ensures backward compatibility with Phase 2 facts

**Design Decision:**
Temporal type is metadata, not a schema change. Stored in `_memory.fact_type` alongside existing fields. No store modifications.

---

## 4. Implementation Notes

**Files Modified:**
- `internal/memory/types.go`: FactType constants, Fact.Type field, FactFromRecord extraction
- `internal/memory/memory.go`: Query methods, constants for types, ConsolidateRecent
- `internal/memory/memory_test.go`: Unit tests for queries
- `cmd/cli/facts_query.go`: CLI query command (new)
- `cmd/cli/main.go`: Register facts query command

**Key Decisions:**
- Type is part of Fact struct (not separate type table)
- Stored as string in metadata (human-readable, queryable)
- Defaults to "state" for backward compat
- Query methods return filtered []Fact (client sorts if needed)
- No automatic type inference (explicit via metadata or LLM)

---

## 5. Acceptance Criteria

- [x] FactType constants defined (atemporal, state, point-in-time)
- [x] Fact.Type field added and populated from metadata
- [x] ConsolidateRecent sets fact_type=state by default
- [x] FactFromRecord defaults to state if type missing
- [x] QueryFactsByType implemented and tested
- [x] GetAtemporalFacts implemented and tested
- [x] GetStateFactsFor implemented and tested
- [x] GetPointInTimeFacts implemented and tested
- [x] CLI query command: `stash facts query --type=state`
- [x] 5 unit tests for fact type queries
- [x] All 130+ existing tests still pass
- [x] No schema changes to store
- [x] Full backward compatibility

---

## 6. Verification Plan

**Unit Tests:**
1. QueryFactsByType with valid types (state, atemporal, point-in-time)
2. QueryFactsByType rejects invalid types
3. Fact.Type defaults to "state" if not in metadata
4. GetAtemporalFacts returns only atemporal facts
5. GetStateFactsFor filters by entity and ValidUntil=nil

**Integration Test:**
1. Create facts of all 3 types via consolidation
2. Query each type separately
3. Verify no cross-contamination
4. Test CLI query command with different types

**Backward Compatibility:**
- Phase 2 facts (no type field) should query as state facts
- Existing ConsolidateRecent facts should work unchanged

---

## 7. Execution Log

- [2026-04-24 22:25] Added FactType constants to types.go
- [2026-04-24 22:25] Added Fact.Type field
- [2026-04-24 22:25] Updated FactFromRecord to extract and default type
- [2026-04-24 22:25] Updated ConsolidateRecent to set fact_type in metadata
- [2026-04-24 22:25] Implemented QueryFactsByType and related query methods
- [2026-04-24 22:25] Added 5 unit tests, all pass
- [2026-04-24 22:25] Created facts_query.go CLI command
- [2026-04-24 22:25] Registered query command in main.go
- [2026-04-24 22:25] All 130+ tests pass, build clean

---

## 8. Outcome

**Final Result:**

Task 0014 (Temporal Fact Types) is complete. Facts now have semantic temporal types that enable Phase 3 retrieval and update strategies.

**What Changed:**
- Fact struct: +1 field (Type string)
- types.go: +25 lines (constants, FactType enum)
- memory.go: +90 lines (query methods, type constants)
- CLI: +60 lines (facts_query command)
- Tests: +100 lines (5 new test functions)
- Total: ~275 lines

**What Was Verified:**
- QueryFactsByType works for all 3 types
- Type defaults to state for backward compat
- All query methods return correct facts
- CLI command executes and outputs JSON
- No schema changes
- All 130+ existing tests pass

**What Remains Open:**
- Task 0015: Entity Relationships (knowledge graph)
- Task 0016: Semantic Consolidation (merging, state updates)
- Task 0017: Confidence-Ranked Retrieval
- Integration tests with real PostgreSQL + OpenAI (pending Task 0017)

---

## 9. Next

Proceed to Task 0015 (Entity Relationships) to add knowledge graph layer.
