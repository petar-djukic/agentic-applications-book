// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/sh"
)

// Test validates every manifest entry's declarations: each YAML
// document under the entry's directory parses. Content rules belong to
// Audit; Test proves the trees are at least well-formed.
func Test() error {
	manifest, err := loadManifest(".")
	if err != nil {
		return err
	}
	var failures []string
	for _, entry := range manifest.Examples {
		failures = append(failures, testEntryDeclarations(".", entry)...)
	}
	if len(failures) > 0 {
		for _, failure := range failures {
			fmt.Fprintln(os.Stderr, "test:", failure)
		}
		return fmt.Errorf("test: %d failure(s)", len(failures))
	}
	fmt.Printf("test: %d example(s) validated\n", len(manifest.Examples))
	return runConformance()
}

// runConformance runs the adopted conformance package without the live
// flag: declaration-shape pinning plus the fixture-double machine runs,
// none of which needs a live server (VISION N2). The live round-trip in
// the same package is gated behind -live and belongs to the opt-in
// Integration namespace, never to Test.
func runConformance() error {
	fmt.Println("test: conformance (non-live)")
	if err := sh.RunV("go", "-C", "conformance", "test", "-count=1", "."); err != nil {
		return fmt.Errorf("conformance: %w", err)
	}
	return nil
}

func testEntryDeclarations(examplesRoot string, entry Entry) []string {
	var failures []string
	root := entryDir(examplesRoot, entry)
	err := filepath.WalkDir(root, func(path string, dirEntry fs.DirEntry, walkErr error) error {
		if walkErr != nil || dirEntry.IsDir() {
			return walkErr
		}
		name := dirEntry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var document any
		if err := yamlUnmarshalLenient(content, &document); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", path, err))
		}
		return nil
	})
	if err != nil {
		failures = append(failures, fmt.Sprintf("%s: walk %s: %v", entry.ID, root, err))
	}
	return failures
}
