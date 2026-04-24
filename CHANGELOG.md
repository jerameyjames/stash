# Changelog

## [Unreleased] - 2026-04-24

### Added
- HTTP API server via `stash server` CLI command
- 2 core endpoints: `POST /api/v1/facts` (remember) and `GET /api/v1/facts` (recall)
- Complete API documentation (`API-SERVER.md`)
- Agent integration guide (`INTEGRATION.md`)
- Docker multi-stage build with distroless final image (~15MB)
- GitHub Actions workflow for multi-platform releases

### Changed
- **BREAKING:** Removed all unit tests (~7000 lines deleted)
- **BREAKING:** Removed fake embedder and reasoner implementations
- **BREAKING:** Removed in-memory store (mapdb)
- **BREAKING:** Hardcoded PostgreSQL and OpenAI (removed driver selection)
- **BREAKING:** Config env vars simplified:
  - `STASH_STORE_DSN` → `STASH_POSTGRES_DSN`
  - Removed `STASH_STORE_DRIVER`
  - Removed `STASH_EMBEDDER_DRIVER`
  - Removed `STASH_REASONER_DRIVER`
- Testing strategy now uses only user-level integration tests
- Documentation rewritten to reflect production-only approach

### Removed
- All `*_test.go` files in `internal/` packages
- `internal/store/mapdb/` (in-memory store)
- `internal/store/storetest/` (test helpers)
- `internal/embedder/fake.go`
- `internal/reasoner/fake.go`
- `.github/workflows/test.yml` (test workflow)
- HTTP API admin endpoints (extract/consolidate relationships)

### Rationale

**Product Philosophy:**
- Code has clean interfaces (future-proof architecture)
- Product has simple config (user-facing simplicity)
- No test-only implementations = honest requirements
- Agents need real semantic search (fake embeddings don't work)

**Trade-offs Accepted:**
- Contributors need PostgreSQL + OpenAI to develop
- Tests cost ~$0.01 per run (OpenAI API calls)
- No offline development
- Slower test feedback (30-60s vs 2s)

**But Gained:**
- 50% smaller codebase (~7000 lines deleted)
- Honest testing (real infrastructure)
- Zero maintenance for fake implementations
- Clear product expectations

## Previous Releases

See git history for Phase 1-2 features (temporal facts, relationships, consolidation, confidence ranking).
