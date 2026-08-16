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

| Arriving in | Content |
|---|---|
| [#26](https://github.com/petar-djukic/agentic-applications-book/issues/26) | The blackboard blocks, as a catalog family this module references |
| [#24](https://github.com/petar-djukic/agentic-applications-book/issues/24) | `agents/rlm-root/` and `agents/rlm-worker/` |
| [#25](https://github.com/petar-djukic/agentic-applications-book/issues/25) | `testdata/` fixture corpus and `demo.yaml` |

Deployment surface — Helm, kind, lifecycle-manager actors, Job workers —
stays upstream in declarative-agents, per non-goal N3 in
[`docs/VISION.yaml`](../../docs/VISION.yaml).
