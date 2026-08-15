<!-- Copyright (c) 2026 Petar Djukic -->
<!-- SPDX-License-Identifier: MIT -->

# Examples

This directory holds the book's runnable artifacts: one declarative
application per chapter rebuild, plus reference profiles copied from
declarative-agents. Every claim a chapter makes about a mechanism has a
machine here that the reader can clone and run.

## Authority directions

Three rules decide which document governs which.

1. Code governs prose. Each chapter's listings are extracted regions of
   example source, never retyped. The audit compares fence against
   marked region byte-for-byte and reports drift as a finding.
2. Upstream governs catalog copies. `catalog/` holds canonical agent
   families copied from declarative-agents, pinned by release in
   `MANIFEST.yaml`, simplified but never forked.
3. Book SRDs govern example SRDs. `docs/srd/` states what each
   application must contain and prove, and cites the chapter SRDs it
   realizes rather than restating them.

## Pinned runtime

Every example runs on declarative-agents at the single release recorded
in `MANIFEST.yaml`. The audit fails a module that pins anything else. A
chapter that needs a newer runtime forces one bump for the whole book,
made in one commit, on purpose.

## Layout

| Path | Role |
|---|---|
| `MANIFEST.yaml` | Index of examples: chapter binding, kind, status, listings, provenance; the pinned runtime release |
| `docs/` | Vision, architecture, road-map, and per-application SRDs for this directory |
| `magefiles/` | The `audit`, `test`, and `demo` targets that enforce the contract |
| `applications/<artifact>/` | One directory per chapter rebuild: `agents/`, `testdata/`, `demo.yaml` |
| `catalog/agents/<family>/` | Canonical profiles copied from declarative-agents |

Applications are named by artifact (`stateflow`, `crag`, `sagas`);
`MANIFEST.yaml` binds each to a stable chapter id, so renumbering
chapters never touches a directory name.

## Demos run canned

Each application ships `testdata/` fixtures with deterministic canned
model responses. `mage demo` needs no live model and no credentials.

## Third-party material

Files under `catalog/` are copied from
[declarative-agents](https://github.com/Nokia-Bell-Labs/declarative-agents)
(BSD-3-Clause, Nokia Bell Labs) and retain their copyright headers. Each
copied family carries a `provenance:` block in `MANIFEST.yaml` recording
the upstream path, the release copied from, and what was simplified.
Listings extract from `applications/`, never from `catalog/`, so
BSD-3-covered source is not reproduced in the built book; the audit
enforces the boundary.
