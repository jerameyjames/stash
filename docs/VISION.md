# Stash — Long-Term Vision

> This document describes where Stash is going, not where it is today.
> Read PRODUCT.md for current scope. Read this to understand why the current scope is shaped the way it is.

---

## The root problem with current AI

Today's LLMs are the most capable pattern-matchers ever built. They can write code, draft legal briefs, explain quantum mechanics. But they have a fundamental flaw that no amount of scaling fixes: **they don't remember, they don't learn, and they don't know what they don't know.**

Every session starts from zero. Every fact lives baked into weights — unauditable, unupdatable, unsourceable. Every answer is delivered with the same confidence whether the model is recalling something it knows well or confabulating something it doesn't. The model cannot tell the difference. Neither can you.

The field's response has been: make the model bigger. More parameters, more data, more compute. And it worked — for a while. But we are hitting the asymptote. The next doubling of compute produces a fraction of the last doubling's improvement. Scaling is no longer the answer.

The real answer is **architecture**. Specifically: **stop asking one giant neural network to do everything.**

---

## The insight Stash is built on

Intelligence has separable components. Human brains figured this out through evolution — different regions handle different jobs. Memory is not the same as reasoning. Reasoning is not the same as perception. Language is not the same as knowledge.

Current LLMs collapse all of these into one thing: weights. Skills, facts, memory, reasoning, language — all tangled in the same matrix multiplication. That's why they're expensive, unreliable, and unable to update without full retraining.

The separation that matters most — and that Stash addresses directly — is:

> **Weights hold skills. Storage holds facts.**

A model should know *how to reason*, *how to write*, *how to code*. It should not need to "remember" that your name is Mohamed or that you prefer dark mode or that you decided last Tuesday to use Go instead of Rust. Those are facts. Facts belong in storage — auditable, updatable, deletable, sourceable.

This is not a new idea. Every well-designed software system separates compute from state. Databases exist precisely because we don't want application logic to "remember" everything. We're just applying this 60-year-old principle to AI.

---

## What Stash becomes

Stash starts as a memory layer. It becomes something larger: **the cognitive substrate that makes stateless models stateful.**

The full system, built in phases:

### Phase 1 (now): The memory primitive
A model can store observations, recall relevant memories, and maintain a working context. This is the foundation. Everything else builds on it.

- `Remember` — store an event with its embedding.
- `Recall` — retrieve what's relevant to this moment.
- `Frame` — hold the current working context.

**The bet:** even this small primitive dramatically improves any LLM-based system that currently relies on context stuffing.

### Phase 2: Cognitive processes
Memory without maintenance is a junk drawer. Phase 2 adds the processes that make memory intelligent:

- **Consolidation** — episodes get processed into facts. Raw observations distilled into durable knowledge.
- **Contradiction detection** — when a new fact conflicts with an old one, surface it. Don't silently overwrite.
- **Decay** — facts get less weight over time unless reinforced. Stale data shouldn't dominate retrieval.
- **Reinforcement** — a pattern observed once is noise. Observed many times, it becomes a fact.
- **Reflection** — periodic passes that ask: what do we know? what's inconsistent? what's missing?

The goal: memory that behaves like memory. Not a database that accumulates sludge, but a system that refines its understanding over time.

### Phase 3: Semantic memory
Phase 1 is episodic (what happened). Phase 3 adds semantic (what is true):

- **Facts as first-class objects** — `"Mohamed prefers Go"` is not an episode, it's a belief. Typed, sourced, confidence-scored.
- **Entity relationships** — `Mohamed → works at → Cartona → is in → Egypt`. A graph layer over facts enables multi-hop reasoning the model can't do alone.
- **Temporal types** — atemporal facts ("Mohamed was born in Egypt"), state facts ("Mohamed is working on Stash"), point-in-time facts ("Mohamed deployed v0.1 on April 18, 2026"). Different retrieval strategies for each.

### Phase 4: The kernel
The kernel is the coordinator — the thin layer that sits between memory and model and decides:

- When to retrieve (before every model call? only when needed?).
- What to retrieve (the right facts for this query, not all facts).
- When to write back (what from this conversation is worth remembering?).
- How to ground the model (inject retrieved facts into the prompt, enforce honesty).

The kernel is not an agent framework. It's a protocol. The model is still the reasoner. The kernel just makes sure the model is reasoning over the right facts.

### Phase 5: The full cognitive system
The long-term vision: a system where:

1. **The model handles language and reasoning.** It doesn't need to "know" things. It reasons over what memory gives it.
2. **Memory handles persistence and truth.** It knows what's true, when it became true, and how confident we are.
3. **The kernel handles the loop.** It orchestrates retrieval, grounding, consolidation, and reflection.
4. **Tools handle exact operations.** Math, code execution, database queries — anything with a correct answer goes to a tool, not a neural net.

This is the "mixed system" that the field is slowly converging on. Not one giant model that does everything, but a system of specialized components each doing one thing well.

---

## The deeper ambition

Current AI systems are stateless by design. Every major LLM product — ChatGPT, Claude, Gemini — treats memory as an afterthought. Some bolt on a vector database. Some stuff conversation history into context. None of them treat memory as a first-class cognitive primitive.

Stash's ambition is to be the **infrastructure layer for AI memory** — the thing that makes any model, any agent, any AI-powered application genuinely stateful.

Not "memory" as in "here's a list of what you said last week." Memory as in:
- The system knows what it knows, and knows what it doesn't.
- The system knows when something became true and whether it's still true.
- The system can surface contradictions in its own beliefs.
- The system gets smarter over time without retraining.
- The data is yours — auditable, exportable, deletable.

**The one-sentence vision: Stash is what makes AI systems learn from experience.**

---

## What we are NOT trying to build

Being explicit about this matters as much as the vision itself.

- ❌ **Not AGI.** We're not solving reasoning, grounding, causality, or embodiment. We're solving one specific piece: persistent, structured, auditable memory.
- ❌ **Not a better LLM.** We don't train models. We make existing models more useful by giving them something they lack.
- ❌ **Not a hosted platform (yet).** Single-user, self-hosted first. Prove the primitive works before building a business around it.
- ❌ **Not a framework.** Stash is a library. It doesn't orchestrate agents, manage prompts, or define workflows. It does one thing: remember.
- ❌ **Not trying to replace the model.** The model reasons. Memory provides ground truth. These are complementary, not competing.

---

## The design principles that follow from this vision

Every technical decision in Stash should be traceable back to one of these:

**1. Memory is infrastructure, not a feature.**
It should be as invisible and reliable as a database. You don't think about your database when it's working. You shouldn't think about Stash either.

**2. Facts and skills must be separable.**
Anything that looks like "the model needs to remember X" is a signal that X belongs in Stash, not in the model's weights.

**3. Truth is earned, not assumed.**
Every fact in Stash has a source, a timestamp, and a confidence. "The model said so" is a source. A weak one. Human confirmation is stronger. Repeated observation is stronger still. Contradictions are surfaced, not hidden.

**4. The user owns their memory.**
Full export. Full deletion. No vendor lock-in. No training on user data. The memory is as personal as a journal and should be treated with the same respect.

**5. Simplicity compounds.**
The temptation in every AI project is to add intelligence to the infrastructure. Resist it. A simple, reliable memory store that any model can use is worth more than a "smart" memory system that's hard to understand and harder to debug. Phase 1 ships a dumb store. Phase 5 is still mostly a dumb store — the intelligence lives in the processes that run over it, not in the storage itself.

---

## The market bet

The AI field is converging on agent-based systems. Every major lab is building agents. Every major developer tool is becoming agentic. Agents need memory. Not context stuffing — real, persistent, structured memory that survives across sessions, tasks, and model versions.

Today, every team building agents rolls their own memory layer — usually a vector database with some ad-hoc retrieval logic bolted on. It works until it doesn't, and then it fails in subtle ways: stale facts, contradictory beliefs, lost context.

Stash is the layer they should be using instead. Open source, self-hosted, composable with any model or agent framework, built on boring reliable infrastructure (Go + Postgres), with the kind of principled design that scales from a personal assistant to a production agent system.

The window for this is now. The agent ecosystem is being built today. The teams that get memory right will outperform the teams that don't. Stash can be the memory layer those teams reach for.

---

## What "done" looks like

Stash is done when:

1. Any developer can add real persistent memory to any LLM application in under an hour.
2. A model using Stash cannot hallucinate a fact that isn't in the store.
3. Memory improves over time — consolidation, contradiction detection, and decay work well enough that the longer you use it, the better it gets.
4. The data is fully portable — export everything, delete everything, run anywhere.
5. The codebase is small enough that one developer can understand all of it in a day.

Point 5 is not a concession to simplicity — it's a feature. A system you can fully understand is a system you can fully trust. That's what we're building.