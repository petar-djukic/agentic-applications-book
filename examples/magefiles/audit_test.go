// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureManifest = `schema_version: 1
runtime:
  module: github.com/Nokia-Bell-Labs/declarative-agents
  release: v0.20260814.4
  go_module: github.com/Nokia-Bell-Labs/declarative-agents/agent-core
  go_version: v0.20260803.0
examples:
  - id: sagas
    chapter: C17
    kind: chapter-application
    status: planned
    srd: docs/srd/srd-sagas.yaml
  - id: executor
    chapter: C17
    kind: catalog-family
    status: implemented
    provenance:
      upstream: Nokia-Bell-Labs/declarative-agents
      path: applications/catalog/agents/executor
      release: v0.20260814.4
      simplified: dropped the REST surface
`

// writeFixture lays out an examples root and a book root that satisfy
// every audit check, so each test breaks exactly one invariant.
func writeFixture(t *testing.T) (examplesRoot, bookRoot string) {
	t.Helper()
	root := t.TempDir()
	examplesRoot = filepath.Join(root, "examples")
	bookRoot = root
	mustWrite(t, filepath.Join(examplesRoot, "MANIFEST.yaml"), fixtureManifest)
	mustWrite(t, filepath.Join(examplesRoot, "applications", "sagas", "agents", "machine.yaml"),
		"# Copyright (c) 2026 Nokia\n# SPDX-License-Identifier: BSD-3-Clause\nname: sagas\n")
	mustWrite(t, filepath.Join(examplesRoot, "docs", "srd", "srd-sagas.yaml"),
		"id: srd-sagas\nrealizes: [srd-17.1]\n")
	mustWrite(t, filepath.Join(examplesRoot, "catalog", "agents", "executor", "profile.yaml"),
		"# Copyright (c) 2026 Nokia\n# SPDX-License-Identifier: BSD-3-Clause\nname: executor\n")
	mustWrite(t, filepath.Join(bookRoot, "docs", "ARCHITECTURE.yaml"),
		"chapters:\n  - id: C17\n")
	mustWrite(t, filepath.Join(bookRoot, "docs", "srd", "srd-17-sagas.yaml"),
		"id: srd-17.1\n")
	return examplesRoot, bookRoot
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func auditFindings(t *testing.T, examplesRoot, bookRoot string) []string {
	t.Helper()
	findings, _ := auditExamples(examplesRoot, bookRoot)
	return findings
}

func requireFinding(t *testing.T, findings []string, fragment string) {
	t.Helper()
	for _, finding := range findings {
		if strings.Contains(finding, fragment) {
			return
		}
	}
	t.Fatalf("no finding contains %q; findings = %v", fragment, findings)
}

func TestAuditPassesOnCompleteFixture(t *testing.T) {
	examplesRoot, bookRoot := writeFixture(t)
	findings, pending := auditExamples(examplesRoot, bookRoot)
	if len(findings) != 0 {
		t.Fatalf("findings on a complete fixture: %v", findings)
	}
	if len(pending) != 0 {
		t.Fatalf("pending on a complete fixture: %v", pending)
	}
}

func TestAuditPassesOnEmptyManifest(t *testing.T) {
	root := t.TempDir()
	examplesRoot := filepath.Join(root, "examples")
	mustWrite(t, filepath.Join(examplesRoot, "MANIFEST.yaml"),
		"schema_version: 1\nruntime:\n  module: m\n  release: v1\nexamples: []\n")
	findings, _ := auditExamples(examplesRoot, root)
	if len(findings) != 0 {
		t.Fatalf("findings on the empty manifest: %v", findings)
	}
}

func TestAuditRejectsSecondRelease(t *testing.T) {
	examplesRoot, bookRoot := writeFixture(t)
	manifest := strings.Replace(fixtureManifest, "release: v0.20260814.4\n      simplified",
		"release: v0.20260813.1\n      simplified", 1)
	mustWrite(t, filepath.Join(examplesRoot, "MANIFEST.yaml"), manifest)
	requireFinding(t, auditFindings(t, examplesRoot, bookRoot), "differs from the pinned runtime release")
}

func TestAuditRejectsStrippedCatalogHeader(t *testing.T) {
	examplesRoot, bookRoot := writeFixture(t)
	mustWrite(t, filepath.Join(examplesRoot, "catalog", "agents", "executor", "profile.yaml"),
		"name: executor\n")
	requireFinding(t, auditFindings(t, examplesRoot, bookRoot), "BSD-3-Clause header missing")
}

func TestAuditRejectsUnresolvableChapterID(t *testing.T) {
	examplesRoot, bookRoot := writeFixture(t)
	mustWrite(t, filepath.Join(bookRoot, "docs", "ARCHITECTURE.yaml"), "chapters:\n  - id: C11\n")
	requireFinding(t, auditFindings(t, examplesRoot, bookRoot), "does not resolve into docs/ARCHITECTURE.yaml")
}

func TestAuditRejectsListingMarkerInCatalog(t *testing.T) {
	examplesRoot, bookRoot := writeFixture(t)
	mustWrite(t, filepath.Join(examplesRoot, "catalog", "agents", "executor", "machine.yaml"),
		"# Copyright (c) 2026 Nokia\n# SPDX-License-Identifier: BSD-3-Clause\n# listing:begin c17-1\nname: executor\n")
	requireFinding(t, auditFindings(t, examplesRoot, bookRoot), "listing marker inside catalog/")
}

func TestAuditRejectsIncompleteProvenance(t *testing.T) {
	examplesRoot, bookRoot := writeFixture(t)
	manifest := strings.Replace(fixtureManifest, "      simplified: dropped the REST surface\n", "", 1)
	mustWrite(t, filepath.Join(examplesRoot, "MANIFEST.yaml"), manifest)
	requireFinding(t, auditFindings(t, examplesRoot, bookRoot), "provenance.simplified is empty")
}

// The require line is the shape a real go.mod carries: the nested
// agent-core module, not the repository path. Against the pre-#35
// matcher this test fails, because that matcher looked for the
// repository path followed by a space and a real line has "/agent-core"
// there instead.
func TestAuditRejectsUnpinnedApplicationModule(t *testing.T) {
	examplesRoot, bookRoot := writeFixture(t)
	mustWrite(t, filepath.Join(examplesRoot, "applications", "sagas", "go.mod"),
		"module example.test/sagas\n\nrequire github.com/Nokia-Bell-Labs/declarative-agents/agent-core v0.20260101.0\n")
	requireFinding(t, auditFindings(t, examplesRoot, bookRoot), "does not pin v0.20260803.0")
}

func TestAuditAcceptsPinnedApplicationModule(t *testing.T) {
	examplesRoot, bookRoot := writeFixture(t)
	mustWrite(t, filepath.Join(examplesRoot, "applications", "sagas", "go.mod"),
		"module example.test/sagas\n\nrequire github.com/Nokia-Bell-Labs/declarative-agents/agent-core v0.20260803.0\n")
	if findings := auditFindings(t, examplesRoot, bookRoot); len(findings) != 0 {
		t.Fatalf("a correctly pinned module must not be a finding: %v", findings)
	}
}

// A path that merely starts with the runtime module string is a
// different module and must not be mistaken for it.
func TestAuditIgnoresLookalikeModulePath(t *testing.T) {
	examplesRoot, bookRoot := writeFixture(t)
	mustWrite(t, filepath.Join(examplesRoot, "applications", "sagas", "go.mod"),
		"module example.test/sagas\n\nrequire github.com/Nokia-Bell-Labs/declarative-agents/agent-core-extras v0.0.1\n")
	if findings := auditFindings(t, examplesRoot, bookRoot); len(findings) != 0 {
		t.Fatalf("a lookalike module path must not be audited as the runtime: %v", findings)
	}
}

func TestAuditRejectsDanglingRealizes(t *testing.T) {
	examplesRoot, bookRoot := writeFixture(t)
	mustWrite(t, filepath.Join(examplesRoot, "docs", "srd", "srd-sagas.yaml"),
		"id: srd-sagas\nrealizes: [srd-99.9]\n")
	requireFinding(t, auditFindings(t, examplesRoot, bookRoot), "realizes id srd-99.9 does not resolve")
}

func TestAuditReportsPendingBindings(t *testing.T) {
	examplesRoot, bookRoot := writeFixture(t)
	if err := os.Remove(filepath.Join(bookRoot, "docs", "ARCHITECTURE.yaml")); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(bookRoot, "docs", "srd")); err != nil {
		t.Fatal(err)
	}
	findings, pending := auditExamples(examplesRoot, bookRoot)
	if len(findings) != 0 {
		t.Fatalf("pending bindings must not be findings: %v", findings)
	}
	if len(pending) != 2 {
		t.Fatalf("pending = %v, want the chapter-id and srd bindings", pending)
	}
}

func TestAuditRejectsUnknownManifestField(t *testing.T) {
	examplesRoot, bookRoot := writeFixture(t)
	mustWrite(t, filepath.Join(examplesRoot, "MANIFEST.yaml"),
		"schema_version: 1\nunknown_field: 1\nruntime:\n  module: m\n  release: v1\nexamples: []\n")
	requireFinding(t, auditFindings(t, examplesRoot, bookRoot), "unknown_field")
}

// --- invariant enforcement (GH-32) ------------------------------------

const fixtureEnforcedSRD = `id: srd-sagas
realizes: [srd-17.1]
invariants:
  - id: I1
    enforce:
      kind: output-schema-excludes
      agent: agents/root
      operation: query_records_filtered
      fields: [documents]
`

const fixtureCleanDeclarations = `tools:
  - name: collect_findings
    output:
      schema:
        properties:
          ids: {type: array}
          metadatas: {type: array}
        additionalProperties: false
    config:
      operation: query_records_filtered
  - name: search_blackboard
    output:
      schema:
        properties:
          documents: {type: array}
        additionalProperties: false
    config:
      operation: some_other_operation
`

// writeEnforcedFixture layers an enforced invariant onto the standard
// fixture: the SRD gains an enforce block and the named agent gains a
// compliant declarations file.
func writeEnforcedFixture(t *testing.T) (examplesRoot, bookRoot string) {
	t.Helper()
	examplesRoot, bookRoot = writeFixture(t)
	mustWrite(t, filepath.Join(examplesRoot, "docs", "srd", "srd-sagas.yaml"), fixtureEnforcedSRD)
	mustWrite(t, filepath.Join(examplesRoot, "applications", "sagas", "agents", "root", "declarations.yaml"),
		fixtureCleanDeclarations)
	return examplesRoot, bookRoot
}

func TestAuditPassesCompliantEnforcedInvariant(t *testing.T) {
	examplesRoot, bookRoot := writeEnforcedFixture(t)
	if findings := auditFindings(t, examplesRoot, bookRoot); len(findings) != 0 {
		t.Fatalf("findings on a compliant enforced invariant: %v", findings)
	}
}

func TestAuditRejectsForbiddenFieldInOutputSchema(t *testing.T) {
	examplesRoot, bookRoot := writeEnforcedFixture(t)
	violating := strings.Replace(fixtureCleanDeclarations,
		"          ids: {type: array}",
		"          ids: {type: array}\n          documents: {type: array}", 1)
	mustWrite(t, filepath.Join(examplesRoot, "applications", "sagas", "agents", "root", "declarations.yaml"), violating)
	requireFinding(t, auditFindings(t, examplesRoot, bookRoot),
		`names "documents" in its output schema`)
}

func TestAuditRejectsOpenOutputSchema(t *testing.T) {
	examplesRoot, bookRoot := writeEnforcedFixture(t)
	for _, open := range []string{"additionalProperties: true", "type: object"} {
		violating := strings.Replace(fixtureCleanDeclarations, "additionalProperties: false", open, 1)
		mustWrite(t, filepath.Join(examplesRoot, "applications", "sagas", "agents", "root", "declarations.yaml"), violating)
		requireFinding(t, auditFindings(t, examplesRoot, bookRoot), "open output schema")
	}
}

func TestAuditIgnoresOtherOperations(t *testing.T) {
	// search_blackboard names documents but binds a different operation;
	// the compliant fixture already passing proves the scoping, and this
	// test pins it against a rename of the unrelated word.
	examplesRoot, bookRoot := writeEnforcedFixture(t)
	renamed := strings.Replace(fixtureCleanDeclarations, "some_other_operation", "another_operation", 1)
	mustWrite(t, filepath.Join(examplesRoot, "applications", "sagas", "agents", "root", "declarations.yaml"), renamed)
	if findings := auditFindings(t, examplesRoot, bookRoot); len(findings) != 0 {
		t.Fatalf("findings for a word on a different operation: %v", findings)
	}
}

func TestAuditRejectsUnknownEnforceKind(t *testing.T) {
	examplesRoot, bookRoot := writeEnforcedFixture(t)
	srd := strings.Replace(fixtureEnforcedSRD, "output-schema-excludes", "no-such-kind", 1)
	mustWrite(t, filepath.Join(examplesRoot, "docs", "srd", "srd-sagas.yaml"), srd)
	requireFinding(t, auditFindings(t, examplesRoot, bookRoot), `enforce kind "no-such-kind" unknown`)
}

func TestAuditRejectsIncompleteEnforceBlock(t *testing.T) {
	examplesRoot, bookRoot := writeEnforcedFixture(t)
	srd := strings.Replace(fixtureEnforcedSRD, "      fields: [documents]\n", "", 1)
	mustWrite(t, filepath.Join(examplesRoot, "docs", "srd", "srd-sagas.yaml"), srd)
	requireFinding(t, auditFindings(t, examplesRoot, bookRoot), "needs agent, operation, and fields")
}

func TestAuditRejectsMissingDeclarationsForEnforcedAgent(t *testing.T) {
	examplesRoot, bookRoot := writeEnforcedFixture(t)
	srd := strings.Replace(fixtureEnforcedSRD, "agents/root", "agents/absent", 1)
	mustWrite(t, filepath.Join(examplesRoot, "docs", "srd", "srd-sagas.yaml"), srd)
	requireFinding(t, auditFindings(t, examplesRoot, bookRoot), "invariant I1: read")
}
