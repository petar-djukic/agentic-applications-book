<!-- Copyright (c) 2026 Petar Djukic -->
<!-- SPDX-License-Identifier: MIT -->

# large-context-swarm

A declarative rebuild of Recursive Language Models (arXiv 2512.24601).
The corpus never enters a context window: it lives in a per-task Chroma
collection the agents treat as a blackboard. A root decomposes the task
into intents, ephemeral workers each search the collection and write one
provenance-tagged finding back, and the root reduces those findings into
a `Final` entry. Only handles and constant-size metadata cross the
root's boundary.

**Status: planned.** Nothing here runs yet. The contract this module
must satisfy is
[`docs/srd/srd-large-context-swarm.yaml`](../../docs/srd/srd-large-context-swarm.yaml);
the work is tracked under
[epic #20](https://github.com/petar-djukic/agentic-applications-book/issues/20).

The blackboard blocks this module composes are already here, as the
`knowledge-manager` catalog family
([`../../catalog/agents/knowledge-manager/`](../../catalog/agents/knowledge-manager/),
contract in
[`docs/srd/srd-knowledge-manager.yaml`](../../docs/srd/srd-knowledge-manager.yaml)).
This module references them; it copies nothing.

`agents/rlm-worker/` is here: one intent in, a filtered query across
vector, `$contains`, and metadata, one provenance-tagged finding out,
then exit. It is the half of the swarm allowed to read corpus text.

| Arriving in | Content |
|---|---|
| [#31](https://github.com/petar-djukic/agentic-applications-book/issues/31) | `agents/rlm-root/` — planning, fan-out, collection by metadata alone, reduction to `Final` |
| [#25](https://github.com/petar-djukic/agentic-applications-book/issues/25) | `testdata/` fixture corpus and `demo.yaml` |

Deployment surface — Helm, kind, lifecycle-manager actors, Job workers —
stays upstream in declarative-agents, per non-goal N3 in
[`docs/VISION.yaml`](../../docs/VISION.yaml).
