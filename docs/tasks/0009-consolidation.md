# Task: Consolidation & Reasoner

**Status:** Completed  
**Date:** 2026-04-23
**Completed:** 2026-04-23

---

## 1. Context

**Goal:** Build the Phase 2 cognitive process for consolidation. Raw events → durable facts via LLM synthesis. Introduce Reasoner abstraction (parallel to Embedder) for structured reasoning.

**Why:** Memory without maintenance is a junk drawer. Consolidation distills raw observations into durable knowledge, enabling contradiction detection, reinforcement, and reflection (Phase 2 later). Without it, memory is just a searchable event log, not intelligence.

**What this is:** 
- New `internal/reasoner/` package with interface + OpenAI implementation
- `Memory.ConsolidateRecent()` method to synthesize events into facts
- Config integration for `STASH_REASONER_DRIVER` and `STASH_REASONER_MODEL`
- Facts stored as Records with `_memory.type=fact` metadata
- Simple conflict detection on synthesis (warn, don't block)

**What this is NOT:**
- Contradiction resolution or merging
- Background consolidation or scheduled tasks
- Auto-tuning of consolidation parameters
- Phase 2 cognitive processes beyond synthesis (reinforcement, reflection come in 0010+)
- Semantic facts or entity relationships (Phase 3)

---

## 2. Boundaries

**In scope:**
- `internal/reasoner/` package with `Reasoner` interface
- `reasoner.OpenAI` implementation (calls `chat.completions` with structured prompt)
- `reasoner.Fake` implementation (deterministic for tests)
- Config: add `STASH_REASONER_DRIVER` and `STASH_REASONER_MODEL` env vars
- Bootstrap: wire Reasoner into context alongside Embedder
- `Memory.ConsolidateRecent(ctx, timeWindow, limit)` method
- Facts stored as Records with `_memory.type=fact`, `_memory.synthesized_from=[event_ids]`
- Simple conflict check: if new fact conflicts with existing, log warning (do not block)
- Tests: unit tests with Fake reasoner, user tests with real synthesis

**Not in scope:**
- Explicit `Memory.FindContradictions()` (comes in 0010)
- Auto-resolution of conflicts (Phase 2)
- Consolidation triggers (manual user call only)
- Background goroutines or scheduled consolidation
- Temporal reasoning or versioning of facts
- Confidence scoring or probabilistic conflict detection
- LLM prompt optimization or tuning

---

## 3. Design

### 3.1 Reasoner abstraction

**New interface in `internal/reasoner/reasoner.go`:**

```go
package reasoner

import "context"

// Reasoner synthesizes structured reasoning over text.
// Implementation is driver-specific (OpenAI, etc).
type Reasoner interface {
	// Reason takes a list of raw texts and returns synthesized reasoning (e.g., a fact).
	// Implementation determines how to group, query LLM, format result.
	Reason(ctx context.Context, texts []string) (string, error)

	// Model returns the model identifier used by this reasoner.
	Model() string

	// Driver returns the driver name (e.g., "openai").
	Driver() string
}
```

**OpenAI implementation in `internal/reasoner/openai.go`:**

```go
type OpenAI struct {
	client *openai.Client
	model  string
	driver string
}

// NewOpenAI constructs an OpenAI reasoner.
// driver: "openai" or other OpenAI-compatible endpoint identifier
// model: e.g., "gpt-4o-mini"
func NewOpenAI(apiKey, driver, model string) (*OpenAI, error)

func (o *OpenAI) Reason(ctx context.Context, texts []string) (string, error) {
	// Build prompt: "Synthesize these events into a single durable fact"
	// Call chat.completions API
	// Return synthesized text (the fact)
}

func (o *OpenAI) Model() string { return o.model }
func (o *OpenAI) Driver() string { return o.driver }
```

**Prompt design (internal, no need to expose):**

```
You are a memory synthesis engine. Given raw observations (events), distill them into a single durable fact.

Events:
- <event 1>
- <event 2>
- ...

Output a single, declarative fact statement (1–2 sentences). Focus on what is true, not when or how often.
Example: "Mohamed prefers Go for systems programming" (not "Mohamed mentioned Go three times").
```

**Fake implementation in `internal/reasoner/fake.go`:**

```go
type Fake struct {
	model  string
	driver string
}

func NewFake(driver, model string) *Fake

func (f *Fake) Reason(ctx context.Context, texts []string) (string, error) {
	// Deterministic: hash input, return consistent output
	// E.g., "Synthesized fact from <len(texts)> events"
	// For tests: verify consolidation logic, not LLM quality
}

func (f *Fake) Model() string { return f.model }
func (f *Fake) Driver() string { return f.driver }
```

### 3.2 Config & bootstrap

**New fields in `internal/config/config.go`:**

```go
type Config struct {
	// ... existing fields ...

	ReasonerDriver string // e.g., "openai"
	ReasonerModel  string // e.g., "gpt-4o-mini"
}
```

**Load from env in `internal/config/load.go`:**

```go
cfg.ReasonerDriver = os.Getenv("STASH_REASONER_DRIVER")
cfg.ReasonerModel = os.Getenv("STASH_REASONER_MODEL")

// Validation: if either set, both must be set
if (cfg.ReasonerDriver != "") != (cfg.ReasonerModel != "") {
	return fmt.Errorf("STASH_REASONER_DRIVER and STASH_REASONER_MODEL must both be set")
}
```

**Wire in `internal/bootstrap/context.go`:**

```go
// In NewBootstrapContext or similar:
var reasoner reasoner.Reasoner
if cfg.ReasonerDriver != "" {
	reasoner, err = reasoner.NewOpenAI(apiKey, cfg.ReasonerDriver, cfg.ReasonerModel)
	if err != nil {
		return nil, err
	}
} else {
	// Default to Fake for tests/local
	reasoner = reasoner.NewFake(cfg.ReasonerDriver, cfg.ReasonerModel)
}

ctx.Reasoner = reasoner
```

### 3.3 Consolidation method

**New method in `internal/memory/memory.go`:**

```go
// ConsolidateRecent groups recent events by semantic similarity and synthesizes each group into a fact.
// timeWindow: how far back to look (e.g., 7*24*time.Hour for last week)
// limit: max number of facts to synthesize in this pass (e.g., 10)
// Returns IDs of newly created facts.
// Errors if Reasoner is nil or Reason call fails.
func (m *Memory) ConsolidateRecent(
	ctx context.Context,
	namespace string,
	timeWindow time.Duration,
	limit int,
) ([]string, error)
```

**Implementation:**

1. Query recent events within `timeWindow` (use store predicates)
2. If < 2 events, return empty (nothing to consolidate)
3. Cluster by semantic similarity (embeddings):
   - Compute pairwise distances between event vectors
   - Group events with distance < threshold (e.g., 0.1) into clusters
   - Keep only top `limit` clusters (by size or recency)
4. For each cluster:
   - Extract event texts
   - Call `m.reasoner.Reason(ctx, texts)` → synthesized fact
   - Check conflict: does fact conflict with existing facts in namespace?
   - If conflict: log warning, include in metadata `_memory.conflict_with=[fact_ids]`
   - Store fact as Record: `{ID: uuid, Text: synthesized, Metadata: {_memory.type: "fact", _memory.synthesized_from: [event_ids]}}`
5. Return created fact IDs

**Conflict detection (simple):**

```go
// Simple heuristic: if fact mentions same entity/property as existing fact
// and values differ, flag conflict.
// For now: check Metadata for `entity` and `property` fields.
// Example: fact A says "Mohamed speaks French", fact B says "Mohamed speaks Spanish"
// Same entity (Mohamed) + property (speaks) + different value → conflict.
// Implementation: scan existing facts in namespace, extract entity+property, compare.
```

**Clustering via embeddings:**

Hand-rolled cosine similarity (no DB optimization needed yet):
- Compute pairwise cosine similarity in Go
- Distance threshold: if similarity > 0.85 (distance < 0.15), same cluster
- Greedy clustering: first event seeds cluster, add similar subsequent events
- Typical volumes (100–500 recent events) → <500ms execution time
- Performance note: if consolidating 10k+ historical events becomes slow, optimize with pgvector distance ordering later (not critical path)

### 3.4 Fact type

**Update `internal/memory/types.go`:**

```go
// Fact represents a durable, synthesized belief derived from events.
// Stored as Record with _memory.type=fact.
// Not directly exposed; retrieve via Recall (facts are just records).
// Fields in metadata:
//   _memory.type: "fact"
//   _memory.synthesized_from: [event_ids]
//   _memory.conflict_with: [fact_ids] (if conflicts detected)
//   _memory.created_at: timestamp
type Fact struct {
	ID              string         `json:"id"`
	Text            string         `json:"text"`
	SynthesizedFrom []string       `json:"synthesized_from"` // event IDs
	ConflictWith    []string       `json:"conflict_with"`    // if any
	CreatedAt       time.Time      `json:"created_at"`
	Metadata        map[string]any `json:"metadata"`
}

// Helper to extract Fact from Record
func FactFromRecord(r *store.Record) (*Fact, error)
```

---

## 4. Implementation Notes

**File changes:**
- `internal/reasoner/reasoner.go` — interface
- `internal/reasoner/openai.go` — OpenAI implementation
- `internal/reasoner/fake.go` — Fake implementation
- `internal/reasoner/reasoner_test.go` — tests (Fake reasoner)
- `internal/config/config.go` — add fields
- `internal/config/load.go` — load from env
- `internal/bootstrap/context.go` — wire Reasoner
- `internal/memory/types.go` — add Fact type + helper
- `internal/memory/memory.go` — add ConsolidateRecent method
- `internal/memory/memory_test.go` — unit tests

**Clustering algorithm (hand-rolled, simple):**

```go
// Greedy clustering: group similar events together
// Threshold: cosine similarity > 0.85 (distance < 0.15)
func clusterBySimilarity(events []*Event, threshold float64) [][]*Event {
	clusters := [][]*Event{}
	used := make(map[string]bool)
	
	for _, e := range events {
		if used[e.ID] {
			continue
		}
		cluster := []*Event{e}
		used[e.ID] = true
		
		for _, other := range events {
			if used[other.ID] {
				continue
			}
			if cosineSimilarity(e.Vector, other.Vector) > threshold {
				cluster = append(cluster, other)
				used[other.ID] = true
			}
		}
		clusters = append(clusters, cluster)
	}
	return clusters
}

// cosineSimilarity computes cosine similarity between two vectors.
// Result in range [0, 1]. Higher = more similar.
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	
	if normA == 0 || normB == 0 {
		return 0
	}
	
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
```

**Performance:** For typical volumes (100–500 recent events per week), clustering is O(n²) but <500ms. Only optimize to pgvector distance ordering if consolidating large historical batches (10k+) becomes a reported bottleneck.

**Conflict detection helper:**

```go
// Check if synthesized fact conflicts with any existing fact in namespace
func (m *Memory) checkConflict(ctx context.Context, namespace, newFact string) ([]string, error) {
	// Query all facts in namespace: Predicate{Field: "_memory.type", Op: "=", Value: "fact"}
	// For each existing fact:
	//   Extract entity, property from metadata
	//   Extract entity, property from newFact (simple regex or LLM parse)
	//   If same entity+property but different value → conflict
	// Return list of conflicting fact IDs
}
```

**Backward compatibility:**
- No changes to existing Memory methods
- New method only
- Reasoner is optional (if nil, ConsolidateRecent returns error)
- Existing code unaffected

---

## 5. Acceptance Criteria

### Reasoner abstraction
- [ ] `internal/reasoner/` package exists
- [ ] `Reasoner` interface defined: `Reason(ctx, texts []string) (string, error)`, `Model()`, `Driver()`
- [ ] `reasoner.OpenAI` implements Reasoner
- [ ] `reasoner.Fake` implements Reasoner
- [ ] OpenAI calls `chat.completions` with structured prompt
- [ ] Fake reasoner produces deterministic output

### Config & bootstrap
- [ ] `STASH_REASONER_DRIVER` and `STASH_REASONER_MODEL` loaded from env
- [ ] Validation: both must be set (or both empty for Fake)
- [ ] Bootstrap wires Reasoner into context
- [ ] Reasoner passed to Memory constructor

### Consolidation method
- [ ] `Memory.ConsolidateRecent(ctx, namespace, timeWindow, limit)` exists
- [ ] Returns ([]string, error) — fact IDs created
- [ ] Queries recent events within timeWindow
- [ ] Clusters events by semantic similarity (cosine distance threshold 0.15 or similar)
- [ ] Calls Reasoner.Reason() for each cluster
- [ ] Stores facts as Records with `_memory.type=fact`
- [ ] Fact metadata includes `_memory.synthesized_from=[event_ids]`
- [ ] Simple conflict check: log warning if fact conflicts, include `_memory.conflict_with=[ids]`
- [ ] Returns empty slice if < 2 events in timeWindow

### Fact type
- [ ] `Fact` type defined in `types.go`
- [ ] `FactFromRecord()` helper extracts Fact from Record

### Testing
- [ ] Unit tests: ConsolidateRecent with Fake reasoner
- [ ] Unit tests: clustering logic (events grouped correctly)
- [ ] Unit tests: conflict detection (detects same entity+property, different value)
- [ ] User tests: create 5 events, consolidate, verify facts stored and searchable
- [ ] User tests: verify fact metadata contains synthesized_from
- [ ] User tests: verify conflict metadata on overlapping facts
- [ ] `go vet` and `staticcheck` pass
- [ ] No new external dependencies (uses existing openai SDK)

---

## 6. Explicit Assumptions

- Reasoner model is required (STASH_REASONER_DRIVER and STASH_REASONER_MODEL must both be set or both unset)
- Clustering uses cosine similarity threshold of ~0.15 (0.85 similarity) — tunable later
- Limit of 10 facts per consolidation pass is reasonable (prevents one consolidation from dominating)
- timeWindow default should be 7 days (caller can pass different value)
- Conflict detection is best-effort heuristic (not guaranteed to catch all conflicts)
- Fact text is stored in Record.Text (same field as events)
- Facts are invisible to Recall() by default (filtered by `_memory.type=event`)

---

## 7. Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| LLM synthesis API failures | Reasoner errors propagate; caller retries or skips |
| Clustering produces huge clusters | Limit to top 10 clusters by size; others left unconsolidated |
| Conflict detection misses real conflicts | Document as Phase 2 heuristic; Phase 2+ adds dedicated detection |
| Memory exhaustion (many events + vectors in-memory) | timeWindow default is 7 days; clusters typically small (<500 events) |
| Prompt injection in event text | LLM sees arbitrary user input; mitigate with strict output parsing |
| Cosine similarity O(n²) on large batches | Hand-rolled is fine for typical volumes. If consolidating 10k+ events becomes slow, optimize with pgvector distance ordering (future task). |

---

## 8. Definition of Done

- Code compiles without warnings
- All unit tests pass (Fake reasoner)
- Consolidation produces facts with correct metadata
- Conflicts detected and logged
- Clustering logic verified with deterministic tests
- Backward compatible (no changes to existing Memory API)
- `go vet` and `staticcheck` pass
- Ready for review
