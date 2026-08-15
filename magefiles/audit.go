// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/magefile/mage/sh"
)

// Audit runs the examples module checks, then the listing-extraction
// check. Failures from either surface as findings.
func Audit() error {
	for _, target := range []string{"audit", "test"} {
		if err := sh.RunV("mage", "-d", "examples", target); err != nil {
			return fmt.Errorf("examples %s: %w", target, err)
		}
	}
	findings, err := extractionFindings(".")
	if err != nil {
		return err
	}
	if len(findings) > 0 {
		for _, finding := range findings {
			fmt.Fprintln(os.Stderr, "audit:", finding)
		}
		return fmt.Errorf("audit: %d extraction finding(s)", len(findings))
	}
	fmt.Println("audit: listing extraction clean")
	return nil
}

// chapterListingMarker precedes a fenced listing in a chapter:
//
//	<!-- listing: c17-1 source=sagas/agents/machine.yaml -->
//
// The source path is relative to examples/applications/. The marked
// region in the source file sits between comment lines containing
// "listing:begin <id>" and "listing:end <id>" -- the same convention
// the examples audit uses to keep markers out of catalog/. The fence
// body must match the region byte-for-byte. A fence without a marker
// is untracked prose and never checked.
var chapterListingMarker = regexp.MustCompile(`<!--\s*listing:\s*(\S+)\s+source=(\S+)\s*-->`)

// extractionFindings scans every numbered chapter for listing markers
// and compares each against its marked region under
// examples/applications/.
func extractionFindings(bookRoot string) ([]string, error) {
	chapters, err := filepath.Glob(filepath.Join(bookRoot, "[0-9][0-9]-*.md"))
	if err != nil {
		return nil, err
	}
	sort.Strings(chapters)
	var findings []string
	for _, chapter := range chapters {
		content, err := os.ReadFile(chapter)
		if err != nil {
			return nil, err
		}
		findings = append(findings, chapterFindings(bookRoot, chapter, string(content))...)
	}
	return findings, nil
}

func chapterFindings(bookRoot, chapter, content string) []string {
	var findings []string
	lines := strings.Split(content, "\n")
	for index := 0; index < len(lines); index++ {
		match := chapterListingMarker.FindStringSubmatch(lines[index])
		if match == nil {
			continue
		}
		id, source := match[1], match[2]
		fence, next, ok := fenceAfter(lines, index+1)
		if !ok {
			findings = append(findings, fmt.Sprintf("%s: listing %s has no fence after its marker", chapter, id))
			continue
		}
		index = next
		region, err := markedRegion(bookRoot, source, id)
		if err != nil {
			findings = append(findings, fmt.Sprintf("%s: listing %s: %v", chapter, id, err))
			continue
		}
		if fence != region {
			findings = append(findings, fmt.Sprintf(
				"%s: listing %s drifted from examples/applications/%s", chapter, id, source))
		}
	}
	return findings
}

// fenceAfter returns the body of the code fence that starts at or
// directly after start (blank lines allowed), and the index of its
// closing delimiter.
func fenceAfter(lines []string, start int) (body string, closing int, ok bool) {
	index := start
	for index < len(lines) && strings.TrimSpace(lines[index]) == "" {
		index++
	}
	if index >= len(lines) || !strings.HasPrefix(strings.TrimSpace(lines[index]), "```") {
		return "", 0, false
	}
	var collected []string
	for cursor := index + 1; cursor < len(lines); cursor++ {
		if strings.TrimSpace(lines[cursor]) == "```" {
			return strings.Join(collected, "\n"), cursor, true
		}
		collected = append(collected, lines[cursor])
	}
	return "", 0, false
}

// markedRegion reads the source under examples/applications/ and
// returns the lines strictly between the begin and end markers.
func markedRegion(bookRoot, source, id string) (string, error) {
	cleaned := filepath.Clean(filepath.FromSlash(source))
	if strings.HasPrefix(cleaned, "..") || filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("source %q escapes examples/applications/", source)
	}
	path := filepath.Join(bookRoot, "examples", "applications", cleaned)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("source examples/applications/%s unreadable: %w", source, err)
	}
	lines := strings.Split(string(content), "\n")
	begin, end := -1, -1
	for index, line := range lines {
		if strings.Contains(line, "listing:begin "+id) {
			begin = index
		}
		if strings.Contains(line, "listing:end "+id) && begin >= 0 {
			end = index
			break
		}
	}
	if begin < 0 || end < 0 {
		return "", fmt.Errorf("marked region %s missing in examples/applications/%s", id, source)
	}
	return strings.Join(lines[begin+1:end], "\n"), nil
}
