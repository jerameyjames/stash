# Stash — Product Brief

> Read this before any task. It tells you what we're building, why, and the decisions already made.
> This is the source of truth for product direction. Code-level rules live in AGENTS.md.

---

## The one-line pitch

**Stash gives stateless AI models persistent, verifiable memory.**

---

## The problem

Today's LLMs are amnesiac. Every session starts from zero. They can't remember what you told them, can't track decisions over time, can't accumulate knowledge. The standard workaround — stuffing everything into the context window — is expensive, fragile, and doesn't scale. RAG helps but is shallow: it treats knowledge as a bag of words with no temporal structure, no contradiction detection, no sense of what was true *then* vs. *now*.

Current LLMs have another deeper problem: they confuse skills with facts. A 70B model stores the capital of France the same way it stores how to reason about code — both baked into weights, neither auditable, neither updatable without retraining. That's wrong. **Skills belong in weights. Facts belong in storage.**

Stash fixes the storage half.

---

## What Stash is

A self-hosted, single-user memory layer for AI systems. It sits between the model and the world, giving any LLM:

- **Persistent episodic memory** — things that happened, when they happened, how important they were.
- **Working context** — what's actively being thought about right now (a Frame).
- **Semantic retrieval** — find relevant memories by meaning, not just keywords.
- **Grounding** — the model can only answer from what's in the store. No hallucinated facts.

Stash is not an agent framework. Not an LLM. Not a RAG pipeline. It's a **primitive** — the memory substrate that other things are built on top of.

---

## What Stash is NOT

These are explicit non-goals. Do not build toward them.

- ❌ Not multi-tenant or multi-user. Single memory space, single instance.
- ❌ Not an LLM wrapper. Stash is model-agnostic — any model can use it.
- ❌ Not a full agent system. Memory is a primitive, not an orchestrator.
- ❌ Not a hosted SaaS. Self-hosted, runs on the user's own machine or server.
- ❌ Not a vector database replacement. Postgres + pgvector is the storage layer.
- ❌ Not trying to solve reasoning, planning, or tool use — those are the model's job.

---

## The core architecture

Three layers, clean separation, one-way dependencies:

```
Model (external)
      ↑
  Kernel (future — orchestrates memory + model)
      ↑
  Memory (internal/brain — episodic + working frame)
      ↑
  Embedder (internal/embedder — text → vector)
      ↑
  Store (internal/store — records, vectors, metadata)
      ↑
  Postgres + pgvector
```

Each layer knows nothing about the layers above it. The store doesn't know what a "fact" is. The embedder doesn't know what "memory" means. Memory doesn't know what model it's serving. Clean primitives, composable by design.

**Unix philosophy applied to intelligence:**
- Store = filesystem (persistence primitive)
- Embedder = text transformer (text → vector)
- Memory = intelligence layer (uses store + embedder, adds memory semantics)
- Kernel = coordinator (future — orchestrates everything)
- Model = reasoner (external, stateless, replaceable)

---

## Key architectural decisions (already made, don't revisit)

**Storage:**
- Single Go binary + Postgres. No second database.
- All memory data in `Record.Metadata` as JSONB. No schema changes for new memory concepts.
- System metadata namespaced under `"_memory"` to prevent collision with caller data.
- Untyped vector column (`vector` without dimension) for backend-agnostic flexibility. No HNSW index until proven necessary at scale.
- Soft delete by default. Hard delete (`Purge`) is explicit.
- Caller-provided TEXT IDs. No auto-generated integer IDs — enables pre-computation, deduplication, cross-system references.

**Embedder:**
- Interface-first: `Embed(ctx, text) ([]float32, error)` + `Model() string` + `Dims() int`.
- Two implementations: `OpenAI` (uses OpenAI-compatible SDK, works with any compatible endpoint) and `Fake` (deterministic, for tests).
- Model string passed as-is to the API — no prefix stripping. OpenRouter needs `"openai/text-embedding-3-small"` with the prefix; OpenAI direct needs it without. Caller's responsibility to use the right format for their endpoint.
- No default model or dimensions. Both required at construction. Fail loudly if missing.
- Model string used as the vector key in store — `"openai/text-embedding-3-small"` → vector slot. Makes re-embedding auditable.

**Memory:**
- Concrete type, not an interface. One implementation. Extend it, don't replace it.
- Three MVP methods: `Remember`, `Recall`, `Frame`.
- `Frame` not `Context` — avoids collision with Go's `context.Context`.
- Working frame expiry handled lazily on read (no background goroutines).
- Single global frame for MVP (`"_memory.working_frame"` fixed ID).

**Testing:**
- Plumbing tests using `Fake` embedder + real Postgres via testcontainers-go.
- Semantic correctness (does Recall return the *right* events?) tested via CLI scripts with a real API key. Not in Go tests.
- No mocking the store — it already works, use it.

**One-way dependency strictly enforced:**
```
internal/store     ← knows nothing above it
      ↑
internal/embedder  ← knows nothing about memory
      ↑
internal/brain    ← uses store + embedder
```

---

## The product philosophy

**1. Separation of concerns applied to intelligence.**
The field's mistake was putting everything — skills, facts, memory, reasoning — into one giant model. Stash enforces the separation: weights hold skills, storage holds facts. This is the insight the whole project is built on.

**2. Boring is a feature.**
Each layer does one thing well. The store is the most boring component in the system by design. Boring means predictable, debuggable, replaceable. The exciting work is in Memory and above. The store just disappears into the background.

**3. Storage-agnostic from the start.**
Memory never writes SQL. It only calls `store.Store` interface methods. Postgres is the implementation, not the contract. This isn't premature abstraction — it's honest naming of what each layer actually is.

**4. Small and shippable over complete and speculative.**
MVP = 3 memory methods. Consolidation, forgetting, graph traversal, contradiction detection — all real features, all Phase 2+. Ship the primitive first. Let real usage tell us what matters next.

**5. Self-hosted, user-owned.**
The data is the user's. Exportable. Deletable. No cloud dependency. The binary + Postgres runs anywhere. This is a design principle, not just a deployment choice.

---

## Current build state

| Component | Status | Task file |
|---|---|---|
| `internal/store` | In progress | `docs/tasks/0001-store.md` |
| `internal/embedder` | Pending | `docs/tasks/0002-memory.md` |
| `internal/brain` | Pending | `docs/tasks/0002-memory.md` |
| Kernel | Not started | Phase 2 |
| Consolidation/decay | Not started | Phase 2 |
| Graph layer | Not started | Phase 4 |

---

## What success looks like (MVP)

A developer can:

1. Point Stash at a Postgres instance and an OpenAI-compatible embedding endpoint.
2. Call `Remember("user prefers dark mode")` and have it stored with a vector.
3. Call `Recall("what are the user's UI preferences?")` and get that memory back.
4. Call `Frame("let's work on the settings page")` and get an active working context.
5. Pass retrieved memories to any LLM as grounded context — and know the model can't hallucinate facts that aren't in the store.

That's it. That's the MVP. Everything else is Phase 2.

---

## What to ask when in doubt

- **"Should this logic live in the store or in memory?"**
  → If it requires understanding what a "fact" or "event" is: Memory. If it's pure persistence: Store.

- **"Should I add an interface here?"**
  → Only if there are two real implementations today or a concrete second one planned. Otherwise: concrete type.

- **"Should I add this feature?"**
  → Check the non-goals list. If it's not there, check the phase plan. If it's Phase 2+, put it in `TODO.md` and move on.

- **"Is this the right abstraction?"**
  → Can you explain it in one sentence without using the word "smart" or "intelligent"? If not, simplify.

---

## The name

**Stash** — a simple, honest name. You stash things so you can find them later. That's exactly what this does.