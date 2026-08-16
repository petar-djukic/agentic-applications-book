// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// listingMarker is the extraction marker convention shared with the
// root magefiles extraction check: a region opens with
// "listing:begin <id>" and closes with "listing:end <id>" inside a
// comment. Audit's only use here is the boundary rule -- no marker may
// appear under catalog/.
var listingMarker = regexp.MustCompile(`listing:(begin|end)\b`)

var releaseShape = regexp.MustCompile(`^v\d`)

// Audit validates the manifest and the ARCHITECTURE constraints:
// pinned-runtime, listing-boundary, provenance completeness, BSD-3
// headers under catalog/, and the two pending bindings into the book's
// docs tree.
func Audit() error {
	findings, pending := auditExamples(".", "..")
	for _, line := range pending {
		fmt.Println("audit: pending:", line)
	}
	if len(findings) > 0 {
		for _, finding := range findings {
			fmt.Fprintln(os.Stderr, "audit:", finding)
		}
		return fmt.Errorf("audit: %d finding(s)", len(findings))
	}
	fmt.Println("audit: all checks passed")
	return nil
}

// auditExamples runs every check and returns findings plus the pending
// bindings. Pending is a distinct state: a check whose reference
// document has not landed yet is reported, not passed, not failed, and
// flips to enforced the moment the document exists.
func auditExamples(examplesRoot, bookRoot string) (findings []string, pending []string) {
	manifest, err := loadManifest(examplesRoot)
	if err != nil {
		return []string{err.Error()}, nil
	}
	findings = append(findings, auditManifestShape(manifest)...)
	findings = append(findings, auditEntries(examplesRoot, manifest)...)
	findings = append(findings, auditCatalogHeaders(examplesRoot)...)
	findings = append(findings, auditListingBoundary(examplesRoot)...)
	findings = append(findings, auditApplicationModules(examplesRoot, manifest)...)

	chapterFindings, chapterPending := auditChapterIDs(bookRoot, manifest)
	findings = append(findings, chapterFindings...)
	pending = append(pending, chapterPending...)

	srdFindings, srdPending := auditSRDRealizes(examplesRoot, bookRoot, manifest)
	findings = append(findings, srdFindings...)
	pending = append(pending, srdPending...)
	return findings, pending
}

func auditManifestShape(manifest Manifest) []string {
	var findings []string
	if manifest.SchemaVersion != 1 {
		findings = append(findings, fmt.Sprintf("manifest schema_version = %d, want 1", manifest.SchemaVersion))
	}
	if manifest.Runtime.Module == "" {
		findings = append(findings, "manifest runtime.module is empty")
	}
	if !releaseShape.MatchString(manifest.Runtime.Release) {
		findings = append(findings, fmt.Sprintf("manifest runtime.release %q is not a release tag", manifest.Runtime.Release))
	}
	return findings
}

func auditEntries(examplesRoot string, manifest Manifest) []string {
	var findings []string
	seen := map[string]bool{}
	for _, entry := range manifest.Examples {
		if entry.ID == "" {
			findings = append(findings, "manifest entry with empty id")
			continue
		}
		if seen[entry.ID] {
			findings = append(findings, fmt.Sprintf("%s: duplicate manifest id", entry.ID))
		}
		seen[entry.ID] = true
		if entry.Kind != kindChapterApplication && entry.Kind != kindCatalogFamily {
			findings = append(findings, fmt.Sprintf("%s: kind %q unknown", entry.ID, entry.Kind))
		}
		if !validStatuses[entry.Status] {
			findings = append(findings, fmt.Sprintf("%s: status %q unknown", entry.ID, entry.Status))
		}
		if entry.Chapter == "" {
			findings = append(findings, fmt.Sprintf("%s: chapter binding is empty", entry.ID))
		}
		if info, err := os.Stat(entryDir(examplesRoot, entry)); err != nil || !info.IsDir() {
			findings = append(findings, fmt.Sprintf("%s: directory %s missing", entry.ID, entryDir(examplesRoot, entry)))
		}
		findings = append(findings, auditProvenance(entry, manifest.Runtime)...)
	}
	return findings
}

func auditProvenance(entry Entry, runtime Runtime) []string {
	if entry.Kind != kindCatalogFamily {
		return nil
	}
	if entry.Provenance == nil {
		return []string{fmt.Sprintf("%s: catalog-family entry has no provenance block", entry.ID)}
	}
	var findings []string
	for field, value := range map[string]string{
		"upstream": entry.Provenance.Upstream, "path": entry.Provenance.Path,
		"release": entry.Provenance.Release, "simplified": entry.Provenance.Simplified,
	} {
		if value == "" {
			findings = append(findings, fmt.Sprintf("%s: provenance.%s is empty", entry.ID, field))
		}
	}
	if entry.Provenance.Release != "" && entry.Provenance.Release != runtime.Release {
		findings = append(findings, fmt.Sprintf(
			"%s: provenance.release %s differs from the pinned runtime release %s",
			entry.ID, entry.Provenance.Release, runtime.Release))
	}
	return findings
}

// auditCatalogHeaders asserts every file under catalog/ keeps its
// upstream BSD-3-Clause header. A missing catalog/ directory means no
// copies exist, which is a valid state, not a finding.
func auditCatalogHeaders(examplesRoot string) []string {
	catalogRoot := filepath.Join(examplesRoot, "catalog")
	var findings []string
	err := filepath.WalkDir(catalogRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		head := content
		if len(head) > 512 {
			head = head[:512]
		}
		if !bytes.Contains(head, []byte("Copyright")) || !bytes.Contains(head, []byte("BSD-3-Clause")) {
			findings = append(findings, fmt.Sprintf("%s: BSD-3-Clause header missing", path))
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		findings = append(findings, fmt.Sprintf("walk catalog: %v", err))
	}
	return findings
}

// auditListingBoundary enforces that no listing marker appears under
// catalog/: listings extract from applications/ only, so BSD-3-covered
// upstream source never reaches the built book.
func auditListingBoundary(examplesRoot string) []string {
	catalogRoot := filepath.Join(examplesRoot, "catalog")
	var findings []string
	err := filepath.WalkDir(catalogRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if listingMarker.Match(content) {
			findings = append(findings, fmt.Sprintf("%s: listing marker inside catalog/", path))
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		findings = append(findings, fmt.Sprintf("walk catalog: %v", err))
	}
	return findings
}

// auditApplicationModules enforces the single-release closure on any Go
// module an application carries: a go.mod requiring the runtime must pin
// the manifest's go_version.
//
// The match is against runtime.go_module and every module nested under
// it, because the runtime is itself a nested module: the repository root
// publishes no go.mod, and a require line names
// .../declarative-agents/agent-core rather than the repository path. An
// earlier version compared against runtime.module followed by a space,
// which no real require line can contain, so the check never fired (#35).
func auditApplicationModules(examplesRoot string, manifest Manifest) []string {
	var findings []string
	pattern := filepath.Join(examplesRoot, "applications", "*", "go.mod")
	matches, _ := filepath.Glob(pattern)
	for _, path := range matches {
		content, err := os.ReadFile(path)
		if err != nil {
			findings = append(findings, fmt.Sprintf("read %s: %v", path, err))
			continue
		}
		for _, line := range strings.Split(string(content), "\n") {
			trimmed := strings.TrimSpace(line)
			if !requiresRuntime(trimmed, manifest.Runtime.GoModule) {
				continue
			}
			if !strings.HasSuffix(trimmed, " "+manifest.Runtime.GoVersion) {
				findings = append(findings, fmt.Sprintf(
					"%s: runtime requirement %q does not pin %s",
					path, trimmed, manifest.Runtime.GoVersion))
			}
		}
	}
	return findings
}

// requiresRuntime reports whether a go.mod line requires the runtime
// module or a module nested under it. A bare prefix test would also
// match a lookalike path, so the character after the prefix has to be a
// path separator or the space before the version.
func requiresRuntime(line, goModule string) bool {
	if goModule == "" {
		return false
	}
	index := strings.Index(line, goModule)
	if index < 0 {
		return false
	}
	rest := line[index+len(goModule):]
	return strings.HasPrefix(rest, " ") || strings.HasPrefix(rest, "/")
}

// auditChapterIDs resolves each entry's chapter id into the book's
// docs/ARCHITECTURE.yaml. Until that document lands (#7) the binding is
// pending.
func auditChapterIDs(bookRoot string, manifest Manifest) (findings, pending []string) {
	path := filepath.Join(bookRoot, "docs", "ARCHITECTURE.yaml")
	content, err := os.ReadFile(path)
	if err != nil {
		if len(manifest.Examples) > 0 {
			pending = append(pending, "chapter-id resolution (docs/ARCHITECTURE.yaml absent, #7)")
		}
		return nil, pending
	}
	for _, entry := range manifest.Examples {
		if entry.Chapter == "" {
			continue
		}
		if !strings.Contains(string(content), entry.Chapter) {
			findings = append(findings, fmt.Sprintf(
				"%s: chapter id %s does not resolve into docs/ARCHITECTURE.yaml", entry.ID, entry.Chapter))
		}
	}
	return findings, nil
}

// auditSRDRealizes checks each example SRD exists and that its
// realizes: ids resolve into the book's docs/srd/. Until that tree
// lands (#10) resolution is pending.
func auditSRDRealizes(examplesRoot, bookRoot string, manifest Manifest) (findings, pending []string) {
	bookSRD := filepath.Join(bookRoot, "docs", "srd")
	bookSRDExists := false
	if info, err := os.Stat(bookSRD); err == nil && info.IsDir() {
		bookSRDExists = true
	}
	pendingNeeded := false
	for _, entry := range manifest.Examples {
		if entry.Kind != kindChapterApplication {
			continue
		}
		if entry.SRD == "" {
			findings = append(findings, fmt.Sprintf("%s: chapter-application entry has no srd", entry.ID))
			continue
		}
		srdPath := filepath.Join(examplesRoot, filepath.FromSlash(entry.SRD))
		realizes, err := loadRealizes(srdPath)
		if err != nil {
			findings = append(findings, fmt.Sprintf("%s: %v", entry.ID, err))
			continue
		}
		if !bookSRDExists {
			pendingNeeded = true
			continue
		}
		for _, id := range realizes {
			if !srdIDResolves(bookSRD, id) {
				findings = append(findings, fmt.Sprintf(
					"%s: realizes id %s does not resolve into the book docs/srd/", entry.ID, id))
			}
		}
	}
	if pendingNeeded {
		pending = append(pending, "srd realizes resolution (docs/srd absent, #10)")
	}
	return findings, pending
}

func loadRealizes(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read srd: %w", err)
	}
	var document struct {
		Realizes []string `yaml:"realizes"`
	}
	if err := yamlUnmarshalLenient(content, &document); err != nil {
		return nil, fmt.Errorf("parse srd %s: %w", path, err)
	}
	return document.Realizes, nil
}

func srdIDResolves(bookSRD, id string) bool {
	entries, err := os.ReadDir(bookSRD)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := os.ReadFile(filepath.Join(bookSRD, entry.Name()))
		if err != nil {
			continue
		}
		if strings.Contains(string(content), id) {
			return true
		}
	}
	return false
}
