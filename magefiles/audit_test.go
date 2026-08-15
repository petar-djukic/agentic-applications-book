// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureSource = `# Copyright (c) 2026 Nokia
# listing:begin c17-1
name: sagas
budget:
  max_iterations: 8
# listing:end c17-1
`

const fixtureChapter = "# Sagas\n\n" +
	"<!-- listing: c17-1 source=sagas/agents/machine.yaml -->\n" +
	"```yaml\n" +
	"name: sagas\n" +
	"budget:\n" +
	"  max_iterations: 8\n" +
	"```\n"

func writeExtractionFixture(t *testing.T) string {
	t.Helper()
	bookRoot := t.TempDir()
	source := filepath.Join(bookRoot, "examples", "applications", "sagas", "agents", "machine.yaml")
	if err := os.MkdirAll(filepath.Dir(source), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte(fixtureSource), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bookRoot, "17-sagas.md"), []byte(fixtureChapter), 0o644); err != nil {
		t.Fatal(err)
	}
	return bookRoot
}

func TestExtractionPassesOnMatchingListing(t *testing.T) {
	bookRoot := writeExtractionFixture(t)
	findings, err := extractionFindings(bookRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings on a matching listing: %v", findings)
	}
}

func TestExtractionPassesOnZeroMarkers(t *testing.T) {
	bookRoot := t.TempDir()
	chapter := "# Chapter\n\n```yaml\nuntracked: fence\n```\n"
	if err := os.WriteFile(filepath.Join(bookRoot, "03-chapter.md"), []byte(chapter), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, err := extractionFindings(bookRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings with zero markers: %v", findings)
	}
}

func TestExtractionCatchesByteDrift(t *testing.T) {
	bookRoot := writeExtractionFixture(t)
	drifted := strings.Replace(fixtureChapter, "max_iterations: 8", "max_iterations: 9", 1)
	if err := os.WriteFile(filepath.Join(bookRoot, "17-sagas.md"), []byte(drifted), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, err := extractionFindings(bookRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || !strings.Contains(findings[0], "drifted") {
		t.Fatalf("findings = %v, want one drift finding", findings)
	}
}

func TestExtractionCatchesDanglingMarker(t *testing.T) {
	bookRoot := writeExtractionFixture(t)
	source := filepath.Join(bookRoot, "examples", "applications", "sagas", "agents", "machine.yaml")
	if err := os.WriteFile(source, []byte("name: sagas\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, err := extractionFindings(bookRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || !strings.Contains(findings[0], "marked region c17-1 missing") {
		t.Fatalf("findings = %v, want one dangling-marker finding", findings)
	}
}

func TestExtractionCatchesMarkerWithoutFence(t *testing.T) {
	bookRoot := writeExtractionFixture(t)
	chapter := "# Sagas\n\n<!-- listing: c17-1 source=sagas/agents/machine.yaml -->\n\nprose instead\n"
	if err := os.WriteFile(filepath.Join(bookRoot, "17-sagas.md"), []byte(chapter), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, err := extractionFindings(bookRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || !strings.Contains(findings[0], "no fence after its marker") {
		t.Fatalf("findings = %v, want one missing-fence finding", findings)
	}
}

func TestExtractionRejectsEscapingSource(t *testing.T) {
	bookRoot := writeExtractionFixture(t)
	chapter := "<!-- listing: c17-1 source=../../secret.yaml -->\n```yaml\nx: 1\n```\n"
	if err := os.WriteFile(filepath.Join(bookRoot, "17-sagas.md"), []byte(chapter), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, err := extractionFindings(bookRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || !strings.Contains(findings[0], "escapes examples/applications/") {
		t.Fatalf("findings = %v, want one escape finding", findings)
	}
}
