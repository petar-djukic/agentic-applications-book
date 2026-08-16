<!-- Copyright (c) 2026 Petar Djukic -->
<!-- SPDX-License-Identifier: MIT -->

# Adopted: the instrumentation contracts from Build a Coding Agent

Seven section requirements documents, copied byte-identical from
[petar-djukic/agentic-coding-book](https://github.com/petar-djukic/agentic-coding-book)
at commit `9770b4f`, where they were `docs/srd/srd-6.*.yaml` — the
drafting contracts for that book's Part VI (Instrumentation). That book
retired the part at
[agentic-coding-book#127](https://github.com/petar-djukic/agentic-coding-book/issues/127)
on the decision that teaching agent instrumentation belongs in this
volume; it keeps citing its instrumented run data as evidence but no
longer teaches observability. This directory is the hand-off.

These are holdings, not chapters. Nothing in this book binds to them
yet, and they must be adapted — not moved — before anything does.

| File | Owns |
|---|---|
| `srd-6.1-the-black-box-problem.yaml` | Making the observability question urgent before any schema |
| `srd-6.2-what-to-log.yaml` | A derivable logging schema; the `monitor` role |
| `srd-6.3-reading-the-logs.yaml` | Log reading as a skill with a method |
| `srd-6.4-the-economics.yaml` | Cost at task granularity; the ~$1.32 per-task floor |
| `srd-6.5-building-intuition.yaml` | Why cost prediction stays a human skill |
| `srd-6.6-non-determinism.yaml` | Run variance; the retry ceiling (pass@1 0.7743 → pass@18 0.795) |
| `srd-6.7-stats-and-post-run-analysis.yaml` | The generation report; the `analyst` role |

## Caveats that travel with the material

Stated in [#13](https://github.com/petar-djukic/agentic-applications-book/issues/13)
and binding on any adaptation:

- Turn-level figures cover **659 of 2,706 task ids** — the instrumented
  subset. Every figure drawn from them must say so.
- The **"$139.24 for 44,628 lines"** economics figure does not reproduce
  against any dataset
  ([agentic-coding-book#92](https://github.com/petar-djukic/agentic-coding-book/issues/92)).
  It must not be reused without that flag.
- The contracts cite the pinned dataset snapshot recorded in the source
  repo's outline header; the datasets themselves live in `go-unix-utils`
  and `cobbler-scaffold`, not here.

## What still points at the other repo

The contracts were written against the coding book's apparatus and carry
its references verbatim: constitution rules from its
`docs/constitutions/voice.yaml` and `argument.yaml` (`V-*`, `A-*` ids),
its chapter and part identifiers (`C6.x`, `P6`), its question chain
(`Q8`), and `prior_material` paths into its corpora. None of these
resolve here. That is expected for a copy and wrong for an adaptation.

## Why this directory is not `docs/srd/`

Two reasons, one mechanical and one of order.

Mechanical: the examples audit
(`examples/magefiles/audit.go`, `auditSRDRealizes`) treats the
existence of `docs/srd/` as the signal to start enforcing the example
SRDs' `realizes:` bindings. The swarm example carries the provisional
id `srd-rlm.1`
([#27](https://github.com/petar-djukic/agentic-applications-book/issues/27)),
so creating `docs/srd/` before the real chapter SRDs exist turns the
audit red.

Of order: this book has no VISION, ARCHITECTURE, constitutions, or SRD
framework yet — that is epic
[#6](https://github.com/petar-djukic/agentic-applications-book/issues/6)
(#7–#10). Whether instrumentation becomes a part of its own here or is
woven through the rebuild chapters is a decision
[#7](https://github.com/petar-djukic/agentic-applications-book/issues/7)
owns, and adapting seven contracts against documents that do not exist
yet would only be redone.

## Obligations on adaptation

When #6 lands and a home is chosen, the adaptation must: re-anchor each
contract to this book's constitutions and chapter ids; replace the
question-chain and `prior_material` references with this book's own or
drop them; carry the coverage and reproducibility caveats above into
every derived figure; and delete this directory once the adapted
contracts exist in `docs/srd/`, so the copy cannot drift into a fork.
