<!-- Copyright (c) 2026 Petar Djukic -->
<!-- SPDX-License-Identifier: MIT -->

# Example SRDs

One software requirements document per application, named
`srd-<artifact>.yaml` after the manifest id. An example SRD is a
software spec, not a second writing spec: it states what the module
must contain, export, and prove — the declarations it ships, the fixture
obligations of its `testdata/`, and what its demo must exercise. None of
that belongs to any chapter SRD.

The binding runs one way. Each example SRD carries a `realizes:` list of
chapter SRD ids that resolve into the book's `docs/srd/`, and the audit
checks that every id resolves. The SRD cites; it never restates. A
chapter renumber therefore breaks nothing here, and a dangling id
surfaces as a finding rather than silent rot.

The book's `docs/srd/` does not exist yet
([#10](https://github.com/petar-djukic/agentic-applications-book/issues/10)).
Until it lands, the audit reports `realizes:` resolution as pending; the
lists are written against the chapter ids the book README already
carries.
