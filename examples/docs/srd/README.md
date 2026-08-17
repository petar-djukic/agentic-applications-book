<!-- Copyright (c) 2026 Petar Djukic -->
<!-- SPDX-License-Identifier: MIT -->

# Example SRDs

One software requirements document per chapter application, named
`srd-<artifact>.yaml` after the manifest id. An example SRD is a
software spec, not a second writing spec: it states what the module
must contain, export, and prove — the declarations it ships, the fixture
obligations of its `testdata/`, and what its demo must exercise. None of
that belongs to any chapter SRD.

A catalog family may carry one too, optionally. Its subject is
different: the contract of the copy — what came over from upstream,
what was deliberately dropped, and what the copied blocks guarantee —
because there is nowhere else to state it. `srd-knowledge-manager.yaml`
is the example. The audit checks any declared `srd:` exists and parses,
whatever the entry's kind; only chapter applications are required to
declare one.

The binding runs one way. Each chapter-application SRD carries a
`realizes:` list of chapter SRD ids that resolve into the book's
`docs/srd/`, and the audit checks that every id resolves. The SRD
cites; it never restates. A chapter renumber therefore breaks nothing
here, and a dangling id surfaces as a finding rather than silent rot.
A catalog-family SRD carries no `realizes:` list — a copy realizes no
chapter — though any ids one did carry would be held to the same
resolution check.
