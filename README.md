# Stash

**Your AI has amnesia. We fixed it.**

Every LLM starts every conversation from zero. Stash gives your agent persistent memory — it remembers, recalls, consolidates, and learns across sessions. No more explaining yourself from scratch.

Open source. Self-hosted. Works with any MCP-compatible agent.

## Without vs With

| | Without Stash | With Stash |
|---|---|---|
| New session | "Who are you again?" | Picks up where you left off |
| Your preferences | Re-explain every time | Already knows them |
| Past mistakes | Repeats the same errors | Remembers what didn't work |
| Long projects | Loses track of goals | Tracks goals across weeks |
| Token cost | Grows every session | Only recalls what matters |
| Switching models | Start from zero again | Memory is model-agnostic |

## How It Works

Stash is a cognitive layer that sits between your AI agent and the world. Episodes become facts. Facts become relationships. Relationships become patterns. Patterns become wisdom.

```
your agent          ← Claude, GPT, local model, anything
  episodes          ← Raw observations, append-only
  facts             ← Synthesized beliefs with confidence
  relationships     ← Entity knowledge graph
  patterns          ← Higher-order abstractions
  goals · failures  ← Intent and learning
  hypotheses        ← Uncertainty with verification plans
postgres + pgvector ← Battle-tested infrastructure
```

Consolidation runs an 8-stage pipeline that turns raw observations into structured knowledge. Each stage only processes new data since the last run — no wasted work.

```
Episodes → Facts → Relationships → Causal Links
  → Goal Progress → Failure Patterns → Hypothesis Evidence → Decay
```

## Setup

**1. Start Postgres with pgvector**

```bash
docker compose up -d
```

**2. Configure your `.env`**

```bash
cp .env.example .env
```

You need: a `Postgres DSN`, an `OpenAI-compatible API key + base URL`, an `embedding model`, and a `reasoner model`. Works with OpenAI, Ollama, OpenRouter — anything compatible.

> The embedding dimension (`STASH_VECTOR_DIM`) must match your model. Use `1536` for `text-embedding-3-small`.

**3. Build and run**

```bash
go build -o stash ./cmd/cli

# Connect your agent via MCP (SSE)
./stash mcp serve --host 0.0.0.0 --port 8080 --with-consolidation

# Or stdio (Claude Desktop, etc.)
./stash mcp execute --with-consolidation
```

The `--with-consolidation` flag runs background consolidation alongside the MCP server — one process does both.

**4. Try it**

```bash
./stash remember "I prefer dark mode and vim keybindings" -n /users/alice
./stash recall "UI preferences" -n /users/alice
./stash consolidate run -n /users/alice
```

## Namespaces

Everything lives in a namespace — hierarchical paths like `/users/alice` or `/projects/stash`. Reads are recursive by default: querying `/users/alice` returns results from alice *and* all sub-namespaces. `/` means everything.

The `init` tool creates a `/self` scaffold for agent self-knowledge:

- `/self/capabilities` — What the agent can do well
- `/self/limits` — What the agent struggles with
- `/self/preferences` — How the agent works best

These are ordinary namespaces — the agent uses Stash on itself.

## Compatibility

Works with any MCP-compatible agent:

`Claude Desktop` · `Cursor` · `Windsurf` · `Cline` · `Continue` · `OpenAI Agents` · `Ollama` · `OpenRouter` · anything MCP

## License

Apache 2.0
