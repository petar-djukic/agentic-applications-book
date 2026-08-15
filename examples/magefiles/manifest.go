// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Manifest mirrors MANIFEST.yaml. Decoding is strict: a field the shape
// does not document is a parse error, so the manifest and this code
// cannot drift apart silently.
type Manifest struct {
	SchemaVersion int     `yaml:"schema_version"`
	Runtime       Runtime `yaml:"runtime"`
	Examples      []Entry `yaml:"examples"`
}

type Runtime struct {
	Module  string `yaml:"module"`
	Release string `yaml:"release"`
}

type Entry struct {
	ID         string      `yaml:"id"`
	Chapter    string      `yaml:"chapter"`
	Kind       string      `yaml:"kind"`
	Status     string      `yaml:"status"`
	SRD        string      `yaml:"srd,omitempty"`
	Listings   []string    `yaml:"listings,omitempty"`
	Provenance *Provenance `yaml:"provenance,omitempty"`
}

type Provenance struct {
	Upstream   string `yaml:"upstream"`
	Path       string `yaml:"path"`
	Release    string `yaml:"release"`
	Simplified string `yaml:"simplified"`
}

const (
	kindChapterApplication = "chapter-application"
	kindCatalogFamily      = "catalog-family"
)

var validStatuses = map[string]bool{"planned": true, "partial": true, "implemented": true}

// loadManifest reads MANIFEST.yaml under examplesRoot.
func loadManifest(examplesRoot string) (Manifest, error) {
	var manifest Manifest
	path := filepath.Join(examplesRoot, "MANIFEST.yaml")
	content, err := os.ReadFile(path)
	if err != nil {
		return manifest, fmt.Errorf("read manifest: %w", err)
	}
	decoder := yaml.NewDecoder(newBytesReader(content))
	decoder.KnownFields(true)
	if err := decoder.Decode(&manifest); err != nil {
		return manifest, fmt.Errorf("parse %s: %w", path, err)
	}
	return manifest, nil
}

// entryDir maps a manifest entry to the directory it describes.
func entryDir(examplesRoot string, entry Entry) string {
	if entry.Kind == kindCatalogFamily {
		return filepath.Join(examplesRoot, "catalog", "agents", entry.ID)
	}
	return filepath.Join(examplesRoot, "applications", entry.ID)
}
