# Task: Confidence-Ranked Retrieval

**Status:** In Execution  
**Date:** 2026-04-24

---

## 1. Context

**Goal:** Rank search results by combining semantic relevance with confidence scores. Higher confidence facts rank higher, relationships with higher confidence are preferred.

**Why:** Raw semantic search treats all results equally. In a memory system, confidence matters: a fact observed 10 times (0.9 confidence) is more reliable than one observed once (0.5 confidence). Ranking enables the model to prioritize well-established beliefs over uncertain ones.

**What this is:**
- New method `Memory.RecallFactsRanked(query, limit)` — semantic search + confidence ranking
- Confidence propagation: fact confidence + relationship confidence in graph traversal
- Ranking formula: `score = (relevance * 0.6) + (confidence * 0.4)`
- Support for relationship-weighted traversal in graph
- CLI command to test confidence-ranked retrieval

**What this is NOT:**
- Reranking all existing search methods (only new method)
- Changing store-level search behavior
- Temporal reasoning or time-decay
- Automatic confidence updates
- Distributed/federated retrieval

---

## 2. Boundaries

**In scope:**
- `Memory.RecallFactsRanked(ctx, namespace, query, limit)` method
  - Searches facts using embedder
  - Returns facts ranked by (relevance * 0.6) + (confidence * 0.4)
  - Confidence comes from fact.Confidence field
- Optional: `Memory.TraverseGraphRanked(ctx, namespace, entity, depth)` 
  - Traverses graph with confidence-weighted edges
  - Returns edges ranked by confidence
- CLI command: `stash facts recall --namespace=<ns> --query=<q> [--ranked]`
  - If `--ranked` flag, use confidence-ranked retrieval
  - Show confidence + relevance scores in JSON output
- Unit tests (3–5 tests)
- Integration test with real facts + relationships

**Not in scope:**
- Changing existing Recall/RecallWhere methods (backward compatibility)
- Multiple ranking formulas or tuning parameters
- Dynamic confidence adjustment
- Real-time reranking across billions of facts
- Ranking for events (only facts)

---

## 3. Approach & Review

**Ranking Formula:**

```
final_score = (relevance_score * 0.6) + (confidence * 0.4)

Where:
  relevance_score = similarity score from embedder (0-1)
  confidence = fact.Confidence (0-1)
```

**Implementation Steps:**

1. Get search results from store (sorted by relevance)
2. For each result, extract as Fact to get confidence
3. Recompute score: `(relevance * 0.6) + (confidence * 0.4)`
4. Sort by new score descending
5. Return top `limit` results

**Optional Graph Ranking:**

- In `TraverseGraph`, edges with higher confidence are traversed first (BFS with priority)
- Useful for "who/what matters most in this entity's network"

**Design Decisions:**

- **Weighting**: 60% relevance, 40% confidence (relevance is primary, confidence is tiebreaker)
- **Simplicity**: No complex ML models, just formula-based
- **Backward compat**: New method only, existing search unchanged
- **Extraction**: Each result converted to Fact to access confidence (no schema change)

---

## 4. Implementation Notes

**Files to Modify:**
- `internal/memory/memory.go`: Add `RecallFactsRanked` method
- Optional: `internal/memory/memory.go`: Add `TraverseGraphRanked` method
- `cmd/cli/recall.go`: Add `--ranked` flag to recall command
- `cmd/cli/main.go`: Register flag if new command needed
- `internal/memory/memory_test.go`: 3–5 tests

**RecallFactsRanked Logic:**

```go
func (m *Memory) RecallFactsRanked(ctx context.Context, namespace, query string, limit int) ([]Fact, error) {
    // 1. Get vector from embedder
    // 2. Search store (get results with similarity scores)
    // 3. For each result, parse as Fact
    // 4. Compute: final_score = (relevance * 0.6) + (confidence * 0.4)
    // 5. Sort by final_score descending
    // 6. Return top limit
}
```

---

## 5. Acceptance Criteria

- [ ] `RecallFactsRanked` method exists and returns Facts ranked by confidence
- [ ] Ranking formula applies correctly: (relevance * 0.6) + (confidence * 0.4)
- [ ] High-confidence facts rank higher when relevance is similar
- [ ] Low-relevance facts are not boosted by confidence alone
- [ ] Results sorted by final score descending
- [ ] Limit parameter works correctly
- [ ] Empty namespace handled gracefully
- [ ] CLI `--ranked` flag works (or new method callable)
- [ ] 3+ unit tests covering: basic ranking, confidence weighting, edge cases
- [ ] Integration test creates facts with different confidences, verifies ranking
- [ ] All existing tests still pass
- [ ] Full backward compatibility

---

## 6. Verification Plan

**Unit Tests:**
1. RecallFactsRanked ranks high-confidence facts higher
2. Relevance + confidence balance (60/40 split)
3. Low-confidence facts ranked lower despite high relevance
4. Empty results handled
5. Limit parameter works correctly

**Integration Test:**
1. Create facts with varied confidence (0.5, 0.7, 0.9)
2. Query semantically similar text
3. Verify results ranked by combined score
4. Verify no regressions to existing Recall method

**Compatibility:**
- Existing Recall/RecallWhere methods unchanged
- No schema changes
- Phase 2 facts work unchanged

---

## 7. Execution Steps

- [ ] Add RecallFactsRanked method to Memory
- [ ] Implement ranking formula
- [ ] Add CLI integration (--ranked flag or new command)
- [ ] Write 3–5 unit tests
- [ ] Run integration test
- [ ] Verify all tests pass
- [ ] Commit with conventional message

---

## 8. Progress Notes

- [2026-04-24 starting] Reading existing search methods

---

## 9. Outcome

(To be filled after completion)
