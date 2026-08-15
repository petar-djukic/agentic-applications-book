// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"path/filepath"
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
