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

Both profiles are here.

`agents/rlm-worker/` takes one intent, runs a filtered query across
vector, `$contains`, and metadata, writes one provenance-tagged finding,
and exits. It is the half of the swarm allowed to read corpus text.

`agents/rlm-root/` plans intents, dispatches a worker per intent through
a sequential `for_each`, reads the round's findings back as record ids
and metadata, and reduces them into a `Final` entry. Its `collect_findings`
word drops the `documents` field of the query response, which is the
line that makes the whole arrangement mean anything —
[#32](https://github.com/petar-djukic/agentic-applications-book/issues/32)
tracks making that structural rather than declared.

| Arriving in | Content |
|---|---|
| [#25](https://github.com/petar-djukic/agentic-applications-book/issues/25) | `testdata/` fixture corpus and `demo.yaml` |

Deployment surface — Helm, kind, lifecycle-manager actors, Job workers —
stays upstream in declarative-agents, per non-goal N3 in
[`docs/VISION.yaml`](../../docs/VISION.yaml).
