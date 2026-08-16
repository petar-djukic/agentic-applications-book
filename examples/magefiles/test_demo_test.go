// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDeclarationsValidationCatchesMalformedYAML(t *testing.T) {
	examplesRoot, _ := writeFixture(t)
	mustWrite(t, filepath.Join(examplesRoot, "applications", "sagas", "agents", "broken.yaml"),
		"name: [unclosed\n")
	manifest, err := loadManifest(examplesRoot)
	if err != nil {
		t.Fatal(err)
	}
	failures := testEntryDeclarations(examplesRoot, manifest.Examples[0])
	if len(failures) != 1 {
		t.Fatalf("failures = %v, want the broken document alone", failures)
	}
}

func TestDeclarationsValidationPassesOnFixture(t *testing.T) {
	examplesRoot, _ := writeFixture(t)
	manifest, err := loadManifest(examplesRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range manifest.Examples {
		if failures := testEntryDeclarations(examplesRoot, entry); len(failures) != 0 {
			t.Fatalf("%s: %v", entry.ID, failures)
		}
	}
}

func TestDemoRunsDeclaredSteps(t *testing.T) {
	examplesRoot, _ := writeFixture(t)
	mustWrite(t, filepath.Join(examplesRoot, "applications", "sagas", "demo.yaml"),
		"steps:\n  - name: canned\n    argv: [\"true\"]\n")
	manifest, err := loadManifest(examplesRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err := runEntryDemo(examplesRoot, manifest.Examples[0]); err != nil {
		t.Fatal(err)
	}
}

func TestDemoFailsOnFailingStep(t *testing.T) {
	examplesRoot, _ := writeFixture(t)
	mustWrite(t, filepath.Join(examplesRoot, "applications", "sagas", "demo.yaml"),
		"steps:\n  - name: canned\n    argv: [\"false\"]\n")
	manifest, err := loadManifest(examplesRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err := runEntryDemo(examplesRoot, manifest.Examples[0]); err == nil {
		t.Fatal("a failing step must fail the demo")
	}
}

// A partial example has a demo that exercises less than its SRD asks
// for. It still runs: skipping it would hide the evidence it does
// produce.
func TestDemoRunsPartialEntries(t *testing.T) {
	examplesRoot, _ := writeFixture(t)
	mustWrite(t, filepath.Join(examplesRoot, "MANIFEST.yaml"),
		strings.Replace(fixtureManifest, "status: planned", "status: partial", 1))
	mustWrite(t, filepath.Join(examplesRoot, "applications", "sagas", "demo.yaml"),
		"steps:\n  - name: canned\n    argv: [\"true\"]\n")
	ran, skipped, err := demoExamples(examplesRoot)
	if err != nil {
		t.Fatal(err)
	}
	if ran != 1 || skipped != 0 {
		t.Fatalf("ran = %d, skipped = %d; want 1 and 0", ran, skipped)
	}
}

// The fixture's sagas entry is planned and carries no demo.yaml, which
// is the state every example passes through before it runs. Demo must
// report it rather than fail on the absent file.
func TestDemoSkipsPlannedEntries(t *testing.T) {
	examplesRoot, _ := writeFixture(t)
	ran, skipped, err := demoExamples(examplesRoot)
	if err != nil {
		t.Fatalf("a planned entry must not fail the demo: %v", err)
	}
	if ran != 0 || skipped != 1 {
		t.Fatalf("ran = %d, skipped = %d; want 0 and 1", ran, skipped)
	}
}

func TestDemoRunsImplementedEntries(t *testing.T) {
	examplesRoot, _ := writeFixture(t)
	mustWrite(t, filepath.Join(examplesRoot, "MANIFEST.yaml"),
		strings.Replace(fixtureManifest, "status: planned", "status: implemented", 1))
	mustWrite(t, filepath.Join(examplesRoot, "applications", "sagas", "demo.yaml"),
		"steps:\n  - name: canned\n    argv: [\"true\"]\n")
	ran, skipped, err := demoExamples(examplesRoot)
	if err != nil {
		t.Fatal(err)
	}
	if ran != 1 || skipped != 0 {
		t.Fatalf("ran = %d, skipped = %d; want 1 and 0", ran, skipped)
	}
}

func TestDemoFailsOnMissingSteps(t *testing.T) {
	examplesRoot, _ := writeFixture(t)
	mustWrite(t, filepath.Join(examplesRoot, "applications", "sagas", "demo.yaml"), "steps: []\n")
	manifest, err := loadManifest(examplesRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err := runEntryDemo(examplesRoot, manifest.Examples[0]); err == nil {
		t.Fatal("an empty demo must fail")
	}
}
