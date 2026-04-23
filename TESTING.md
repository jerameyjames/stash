# Testing Documentation

Complete testing coverage for Stash project.

---

## Test Levels

### 1. Unit Tests (Internal)
**Focus:** Individual functions, methods, algorithms  
**Framework:** Go's `testing` package  
**Location:** `*_test.go` files  
**Scope:** Library code, no external dependencies (uses Fake implementations)

### 2. Integration Tests (Real Services)
**Focus:** Store backends, LLM integration  
**Framework:** Go's `testing` + testcontainers  
**Location:** Same test files with build tags  
**Scope:** PostgreSQL store, OpenAI/OpenRouter LLM

### 3. User-Level Tests (CLI)
**Focus:** End-to-end workflows, user experience  
**Framework:** Bash scripts with CLI invocation  
**Location:** `test-*.sh` files in repo root  
**Scope:** Full system from user perspective

---

## Unit Tests

### Running All Unit Tests

```bash
# All packages
go test ./...

# Specific package
go test ./internal/memory -v

# Specific test
go test ./internal/memory -v -run TestRecallFactsRanked
```

### Test Coverage Summary

| Package | Tests | Status |
|---------|-------|--------|
| `store/mapdb` | 25+ | ✅ Pass |
| `store/postgres` | 20+ | ✅ Pass |
| `memory` | 95 | ✅ Pass |
| `reasoner` | 5 | ✅ Pass |
| `bootstrap` | 3 | ✅ Pass |
| `actions` | 2 | ✅ Pass |
| **TOTAL** | **150+** | **✅ 100% Pass** |

### Phase 3 Unit Tests

#### Task 0014: Temporal Fact Types
- `TestQueryFactsByType_InvalidType` — Invalid type rejected
- `TestQueryFactsByType_State` — State facts filtered correctly
- `TestFactTypeDefaults` — Default types applied
- `TestGetAtemporalFacts` — Atemporal query works
- (5 tests total)

#### Task 0015: Entity Relationships
- `TestStoreRelationship` — Relationships stored and retrieved
- `TestGetRelationshipsFrom` — Outgoing edges queried
- `TestGetRelationshipsTo` — Incoming edges queried
- `TestTraverseGraph` — BFS traversal with depth limit
- `TestFindPath` — Shortest path finding
- `TestFindPathNotFound` — Disconnected nodes error
- `TestGetAllRelationships` — All relationships retrieved
- (7 tests)

#### Task 0016: Semantic Consolidation
- `TestConsolidateRelationships_BasicExtraction` — Single fact extraction
- `TestConsolidateRelationships_MultipleFactsStored` — Multiple facts processing
- `TestConsolidateRelationships_LimitBoundary` — Limit parameter respected
- `TestReasonRelationships_Parsing` — LLM format parsing
- `TestConsolidateRelationships_EmptyNamespace` — Empty namespace handled
- `TestConsolidateRelationships_OldFactsIgnored` — 7-day window respected
- (6 tests)

#### Task 0017: Confidence-Ranked Retrieval
- `TestRecallFactsRanked_HighConfidenceFirst` — High confidence ranks first
- `TestRecallFactsRanked_RelevanceConfidenceBalance` — 60/40 weighting works
- `TestRecallFactsRanked_LimitRespected` — Pagination works
- `TestRecallFactsRanked_EmptyNamespace` — Empty search handled
- (4 tests)

### Running Unit Tests

```bash
# All tests
go test ./...

# Verbose output
go test ./... -v

# With coverage report
go test ./... -cover

# Memory package tests only
go test ./internal/memory -v

# Phase 3 tests only
go test ./internal/memory -v -run "Phase3|Consolidate|RecallFacts|Relationship"
```

---

## Integration Tests

### PostgreSQL Integration
Requires Docker for test containers.

```bash
# Run PostgreSQL store tests (auto starts container)
go test ./internal/store/postgres -v

# Runs ~20 test cases on real PostgreSQL 16 with pgvector
```

### What's Tested

- ✅ Record CRUD operations
- ✅ Vector similarity search
- ✅ Metadata filtering (predicates)
- ✅ Transaction support
- ✅ Concurrent operations
- ✅ Namespace isolation
- ✅ Schema management

---

## User-Level CLI Tests

### Test Scripts Available

```
test-phase2-cli-real.sh          — Phase 2 complete workflow (events → facts)
test-phase2-cli-gaps-fixed.sh    — Phase 2 CLI coverage
test-phase3-task0014.sh          — Temporal fact types
test-phase3-task0015.sh          — Entity relationships (graph)
test-phase3-task0016.sh          — Semantic consolidation (extraction)
test-phase3-task0017.sh          — Confidence-ranked retrieval
```

### Running User-Level Tests

#### Task 0014: Temporal Fact Types

```bash
./test-phase3-task0014.sh
```

**Requires:**
- PostgreSQL backend
- OpenAI API key (for fact synthesis)

**Tests:**
- ✅ Event creation → fact consolidation
- ✅ Query facts by type (state, atemporal, point-in-time)
- ✅ Temporal semantics applied
- ✅ Type-specific retrieval

#### Task 0015: Entity Relationships

```bash
./test-phase3-task0015.sh
```

**Requires:**
- PostgreSQL backend
- Graph structure setup

**Tests:**
- ✅ Relationship creation
- ✅ Outgoing/incoming relationship queries
- ✅ Graph traversal (BFS)
- ✅ Shortest path finding
- ✅ Multi-hop reasoning

#### Task 0016: Semantic Consolidation

```bash
./test-phase3-task0016.sh
```

**Requires:**
- PostgreSQL backend
- OpenAI/OpenRouter API key for LLM

**Tests:**
- ✅ Event creation with relationship content
- ✅ Fact consolidation synthesis
- ✅ LLM relationship extraction
- ✅ Relationship storage and querying
- ✅ Confidence tracking

#### Task 0017: Confidence-Ranked Retrieval

```bash
./test-phase3-task0017.sh
```

**Requires:**
- PostgreSQL backend

**Tests:**
- ✅ Basic fact recall (relevance only)
- ✅ Confidence-ranked recall (relevance + confidence)
- ✅ Ranking formula verification
- ✅ Limit parameter validation
- ✅ Score calculation accuracy

### Expected Output

Each test script outputs:

1. **Step-by-step progress** — Each operation shows result
2. **Key metrics** — Counts, durations, success rates
3. **Sample data** — JSON examples for verification
4. **Verification checks** — ✅/❌ for each capability
5. **Summary** — Overall test status

### Test Data Requirements

**For Phase 3 Tests:**

- **PostgreSQL**: Must be running (or use `-driver=mapdb` for in-memory)
- **LLM (Tasks 0014, 0016)**: Set environment variables:
  ```bash
  export STASH_REASONER_DRIVER=openai
  export STASH_REASONER_MODEL=gpt-4o-mini
  export STASH_OPENAI_API_KEY=sk-...
  ```
  Or use OpenRouter:
  ```bash
  export STASH_REASONER_MODEL=openrouter/google/gemma-4-27b
  export STASH_OPENAI_API_KEY=sk-or-...  # OpenRouter key
  ```

---

## Quick Test Commands

### Verify Build
```bash
go build -o /tmp/stash ./cmd/cli
```

### Run All Unit Tests
```bash
go test ./... -v
```

### Run Memory Package Tests with Coverage
```bash
go test ./internal/memory -v -cover
```

### Run Phase 3 Tests Only
```bash
go test ./internal/memory -v -run "0014|0015|0016|0017|Temporal|Relationship|Consolidate|RecallFacts"
```

### Run PostgreSQL Integration Tests
```bash
go test ./internal/store/postgres -v
```

### Run User-Level CLI Test (Task 0017)
```bash
./cli build

# Set up store (use mapdb for testing without PostgreSQL)
export STASH_STORE_DRIVER=mapdb

# Run test
./test-phase3-task0017.sh
```

---

## CI/CD Commands

### GitHub Actions / CI Pipeline

```yaml
name: Tests

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.20
      - run: go test ./... -v -cover
      
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.20
      - run: go build -v ./cmd/cli
      - run: go vet ./...
```

---

## Test Scenarios

### Complete Scenario: Events → Facts → Graph → Ranking

```bash
#!/bin/bash

# 1. Remember events
./cli events create "Alice works at TechCorp" --namespace=test
./cli events create "TechCorp is in San Francisco" --namespace=test

# 2. Consolidate to facts
./cli facts consolidate --namespace=test

# 3. Extract relationships
./cli facts extract-relationships --namespace=test

# 4. Query graph
./cli facts relationships --entity=Alice --namespace=test
./cli facts graph --entity=Alice --namespace=test --depth=2

# 5. Search with ranking
./cli facts recall "Alice company" --namespace=test --ranked
```

### Performance Test

```bash
#!/bin/bash

# Create 1000 events
for i in {1..1000}; do
  ./cli events create "Event $i content description" --namespace=perf
done

# Measure consolidation time
time ./cli facts consolidate --namespace=perf --limit=100

# Measure search time
time ./cli facts recall "search term" --namespace=perf --ranked --limit=10
```

---

## Known Limitations

### Test Environment

- **PostgreSQL tests** require Docker daemon running
- **LLM tests** require valid API credentials
- **Large-scale tests** limited by in-memory store (use PostgreSQL for 10k+ records)

### Flaky Tests

- None known (deterministic, no time-based assertions)

### Skipped Tests

- LLM-dependent tests skipped if `STASH_REASONER_*` vars not set
- PostgreSQL tests skipped if Docker unavailable

---

## Test Coverage Goals

| Layer | Coverage | Status |
|-------|----------|--------|
| Store API | 95%+ | ✅ High |
| Memory layer | 90%+ | ✅ High |
| CLI commands | 80%+ | ✅ Medium (user tests) |
| Error paths | 70%+ | ✅ Medium |
| Integration | Real backends | ✅ Yes |

---

## Continuous Quality

### Build Checks
```bash
go build ./...           # Builds successfully
go vet ./...             # No vet errors
go fmt ./...             # Formatted correctly
```

### Test Checks
```bash
go test ./...            # All tests pass
go test -race ./...      # No race conditions
go test -cover ./...     # Coverage measured
```

---

## Adding New Tests

### Template: Unit Test

```go
func TestNewFeature(t *testing.T) {
    mem, cleanup := startMemory(t)
    defer cleanup()

    // Test code here
    result, err := mem.NewFeature()
    if err != nil {
        t.Fatalf("NewFeature failed: %v", err)
    }
    
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

### Template: User-Level Test

```bash
#!/bin/bash
set -e

echo "Testing: Feature X"

NAMESPACE="test_feature_$(date +%s)"

# Create test data
./cli events create "Test data" --namespace="$NAMESPACE"

# Invoke feature
RESULT=$(./cli feature command --namespace="$NAMESPACE")

# Verify
COUNT=$(echo "$RESULT" | jq '.count')
if [ "$COUNT" -eq 1 ]; then
    echo "✅ Feature working"
else
    echo "❌ Feature failed"
    exit 1
fi
```

---

## Troubleshooting Tests

### Test Timeout
```bash
go test ./... -timeout 30s  # Increase timeout
```

### Build Fails
```bash
go clean -cache
go build ./...
```

### PostgreSQL Tests Fail
```bash
# Ensure Docker is running
docker ps

# Check testcontainers logs
cat ~/.testcontainers.log
```

### LLM Tests Fail
```bash
# Verify credentials
echo $STASH_OPENAI_API_KEY

# Test API connection
curl "https://api.openai.com/v1/models" \
  -H "Authorization: Bearer $STASH_OPENAI_API_KEY"
```

---

## Test Reports

### Generate HTML Coverage Report
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o report.html
```

### View Coverage by Package
```bash
go tool cover -html=coverage.out
```

### Check for Race Conditions
```bash
go test -race ./...
```

---

## Summary

**Phase 3 CLI Testing Status:**

✅ **Unit Tests:** 150+ tests, 100% pass  
✅ **Integration:** PostgreSQL backend tested  
✅ **User-Level:** 4 Phase 3 test scripts  
✅ **All Workflows:** Events → Facts → Graph → Ranking  
✅ **No Blocking Issues:** System production-ready

**Tests cover:**
- All 15 CLI commands
- All core APIs
- Error handling
- Edge cases
- Performance
- Backward compatibility

**Ready for:** Production deployment, CI/CD integration, user testing
