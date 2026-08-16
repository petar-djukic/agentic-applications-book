// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFixture lays out a two-part docs tree: one paper chapter, one
// substrate chapter, one blocked chapter.
func writeFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write := func(path, content string) {
		t.Helper()
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("docs/ARCHITECTURE.yaml", `
structure:
  parts:
    - id: P1
      title: "Part I"
      chapters:
        - {id: C3, title: "Substrate", srd: docs/srd/srd-c3.yaml}
    - id: P2
      title: "Part II"
      chapters:
        - {id: C11, title: "Paper", srd: docs/srd/srd-c11.yaml}
        - {id: C21, title: "Blocked", srd: docs/srd/srd-c21.yaml}
`)
	write("docs/road-map.yaml", `
chapters:
  - {id: C3, status: drafted}
  - {id: C11, status: stub}
  - {id: C21, status: stub}
`)
	write("docs/srd/srd-c3.yaml", `
meta: {id: srd-c3, paper_fields: not-applicable}
section_goal: what the substrate chapter accomplishes
`)
	write("docs/srd/srd-c11.yaml", `
meta: {id: srd-c11}
paper:
  - {title: "A Paper", authors_year: "Someone (2024)", arxiv: "1234.5678"}
claim_under_test: a falsifiable claim
verdict_type: holds
`)
	write("docs/srd/srd-c21.yaml", `
meta: {id: srd-c21}
status: blocked
blocked_on: {mechanism: the missing mechanism}
paper:
  - {title: "Blocked Paper", authors_year: "Other (2023)", arxiv: "8765.4321"}
`)
	return root
}

func TestRenderOrdersPartsAndChapters(t *testing.T) {
	out, err := render(writeFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	order := []string{"## Part I", "### C3", "## Part II", "### C11", "### C21"}
	last := -1
	for _, marker := range order {
		i := strings.Index(out, marker)
		if i < 0 {
			t.Fatalf("missing %q in output", marker)
		}
		if i < last {
			t.Fatalf("%q appears out of order", marker)
		}
		last = i
	}
}

func TestRenderChapterKinds(t *testing.T) {
	out, err := render(writeFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"what the substrate chapter accomplishes", // substrate: section goal, no paper block
		"**Paper:** Someone (2024). *A Paper*. 1234.5678",
		"**Claim under test.** a falsifiable claim",
		"**Predicted verdict:** holds",
		"**Blocked on:** the missing mechanism",
		"*Status: drafted*", // road-map status when the SRD has none
		"*Status: blocked*", // SRD status wins when present
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in output:\n%s", want, out)
		}
	}
	if strings.Contains(strings.Split(out, "## Part II")[0], "**Paper:**") {
		t.Fatal("substrate chapter must not render a paper block")
	}
}

func TestRenderFailsWithoutSRDTree(t *testing.T) {
	root := writeFixture(t)
	if err := os.RemoveAll(filepath.Join(root, "docs", "srd")); err != nil {
		t.Fatal(err)
	}
	if _, err := render(root); err == nil {
		t.Fatal("a missing docs/srd must be a hard error")
	}
}

func TestRenderFailsOnMissingRosterSRD(t *testing.T) {
	root := writeFixture(t)
	if err := os.Remove(filepath.Join(root, "docs", "srd", "srd-c11.yaml")); err != nil {
		t.Fatal(err)
	}
	_, err := render(root)
	if err == nil || !strings.Contains(err.Error(), "C11") {
		t.Fatalf("a roster SRD that cannot load must fail naming the chapter; got %v", err)
	}
}
