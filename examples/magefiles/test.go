// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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
