<!-- Copyright (c) 2026 Petar Djukic -->
<!-- SPDX-License-Identifier: MIT -->

# Section Requirements Documents

One SRD per chapter, written before the chapter drafts. The SRD is the
drafting contract: a chapter issue executes its SRD without re-deriving
the plan (process.yaml P-1). Simple by design — these are writing specs,
not systems ceremony.

Files are named per the roster in `../ARCHITECTURE.yaml` (each chapter's
`srd:` path); ids are `srd-<chapter-id>` in lowercase (`srd-c3`,
`srd-c11`, `srd-c-rlm`). The id is what other documents bind to — the
examples' `realizes:` lists resolve against these files, enforced by the
examples audit.

## Base fields

| Field | Content |
|---|---|
| `meta` | id, chapter id, title, file |
| `section_goal` | one phrase: what the chapter accomplishes for the reader |
| `goals` | `{id, goal}` list; ids are `G<n>.<m>` under the VISION goal `G<n>` they serve |
| `objective` | one sentence: how the goals are achieved |
| `prior_material` | `{path, offers}` list: what to quarry, and what each source offers |
| `citations` | `{id, role, note}` list; `role` is anchor, survey, evidence, or counterpoint; ids resolve in `../../references.yaml`, or the entry carries `gap:` naming what an update-references pass must find |
| `constitutions` | rule ids this chapter leans on hardest (V-*/A-*/P-*/N-*); cite, never restate |
| `content` | ordered list of what the chapter says, mapped onto its section structure |
| `links` | `requires` / `supports`: chapter ids in reading-order dependency |
| `gaps` | holes worth filling before or during drafting |
| `acceptance` | drafting-readiness checks, stated so a critic can verify them |

## Paper-chapter fields

Paper chapters (the map, Part II, Part III, the closer) add six fields
that encode the chapter form (voice.yaml V-S1..S7):

| Field | Content |
|---|---|
| `paper` | every paper the chapter rebuilds: title, authors-year, arXiv or venue id; the implemented paper is always the chapter's reference [1] |
| `claim_under_test` | the paper's claim restated falsifiably — a run could contradict it; this is what the retest box prints (V-R1) |
| `machine_paths` | the files that carry the mechanism: paths in declarative-agents, in a named sibling repository, or in `examples/`; the mapping table's source (V-S3) |
| `modification` | the exercise (V-S5): change X, run Y, expect Z, because W |
| `verdict_type` | the *predicted* verdict from the taxonomy (A-3, definitions.yaml); the draft may land elsewhere, and the SRD is updated when it does |
| `create_pointer` | the extension the chapter closes on (V-S6): what to build, which pieces to add |

For substrate chapters (Part I) these six fields do not apply; the SRD
carries `paper_fields: not-applicable` in `meta` instead of empty
fields, so the absence is a statement rather than an omission.

## Placeholders

A Create-tier chapter whose mechanism has not landed gets a placeholder
SRD rather than no SRD: real id, `status: blocked`, `blocked_on:`
carrying the road-map's named mechanism, plus whatever is already known
(paper, claim under test) so unblocking is a fill-in, not a restart.
The road-map (P-4) is the authority on what blocks; the placeholder
quotes it.

Canonical terminology is set in `../definitions.yaml`; register and
framing rules live in the constitutions. SRDs cite both and restate
neither.
