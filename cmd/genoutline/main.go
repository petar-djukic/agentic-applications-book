// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

// Command genoutline renders the book's outline as markdown on stdout:
// part by part, each chapter with its paper, claim under test, predicted
// verdict, and status. The roster in docs/ARCHITECTURE.yaml owns the
// order and the part structure; each chapter's SRD under docs/srd/
// supplies the content; docs/road-map.yaml supplies the drafting
// status. A missing tree or an unreadable roster SRD is a hard error --
// the outline is a review artifact, and a silently partial one would
// review the wrong book.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type architecture struct {
	Structure struct {
		Parts []part `yaml:"parts"`
	} `yaml:"structure"`
}

type part struct {
	ID       string    `yaml:"id"`
	Title    string    `yaml:"title"`
	Chapters []chapter `yaml:"chapters"`
}

type chapter struct {
	ID    string `yaml:"id"`
	Title string `yaml:"title"`
	SRD   string `yaml:"srd"`
}

type srd struct {
	Meta struct {
		ID          string `yaml:"id"`
		Title       string `yaml:"title"`
		PaperFields string `yaml:"paper_fields"`
	} `yaml:"meta"`
	Status      string `yaml:"status"`
	SectionGoal string `yaml:"section_goal"`
	Paper       []struct {
		Title       string `yaml:"title"`
		AuthorsYear string `yaml:"authors_year"`
		Arxiv       string `yaml:"arxiv"`
	} `yaml:"paper"`
	ClaimUnderTest string `yaml:"claim_under_test"`
	VerdictType    string `yaml:"verdict_type"`
	BlockedOn      struct {
		Mechanism string `yaml:"mechanism"`
	} `yaml:"blocked_on"`
}

type roadMap struct {
	Chapters []struct {
		ID     string `yaml:"id"`
		Status string `yaml:"status"`
	} `yaml:"chapters"`
}

func main() {
	if err := run(os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "genoutline:", err)
		os.Exit(1)
	}
}

func run(out *os.File) error {
	// The docs tree root: the first argument when given (mage runs the
	// generator from its own module directory), the working directory
	// otherwise.
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	document, err := render(root)
	if err != nil {
		return err
	}
	_, err = out.WriteString(document)
	return err
}

// render builds the outline markdown from the docs tree under root.
func render(root string) (string, error) {
	arch, err := loadArchitecture(root)
	if err != nil {
		return "", err
	}
	drafting, err := loadRoadMap(root)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "srd")); err != nil {
		return "", fmt.Errorf("docs/srd is missing: %w", err)
	}

	var b strings.Builder
	writeHeader(&b)
	for _, p := range arch.Structure.Parts {
		fmt.Fprintf(&b, "## %s\n\n", p.Title)
		for _, c := range p.Chapters {
			doc, err := loadSRD(root, c.SRD)
			if err != nil {
				return "", fmt.Errorf("chapter %s: %w", c.ID, err)
			}
			writeChapter(&b, c, doc, drafting[c.ID])
		}
	}
	return b.String(), nil
}

func writeHeader(b *strings.Builder) {
	b.WriteString("---\n")
	b.WriteString("title: \"Agentic Applications: Outline\"\n")
	fmt.Fprintf(b, "date: %s\n", time.Now().Format("2006-01-02"))
	b.WriteString("---\n\n")
	b.WriteString("Generated from docs/ARCHITECTURE.yaml, docs/srd/, and docs/road-map.yaml by `mage outline`. Do not edit.\n\n")
}

func writeChapter(b *strings.Builder, c chapter, doc srd, drafting string) {
	fmt.Fprintf(b, "### %s — %s\n\n", c.ID, c.Title)
	status := doc.Status
	if status == "" {
		status = drafting
	}
	fmt.Fprintf(b, "*Status: %s*\n\n", status)

	if doc.Meta.PaperFields == "not-applicable" {
		fmt.Fprintf(b, "%s\n\n", strings.TrimSpace(doc.SectionGoal))
		return
	}
	for _, p := range doc.Paper {
		fmt.Fprintf(b, "**Paper:** %s. *%s*. %s\n\n", p.AuthorsYear, p.Title, p.Arxiv)
	}
	if claim := strings.TrimSpace(doc.ClaimUnderTest); claim != "" {
		fmt.Fprintf(b, "**Claim under test.** %s\n\n", claim)
	}
	if doc.VerdictType != "" {
		fmt.Fprintf(b, "**Predicted verdict:** %s\n\n", doc.VerdictType)
	}
	if doc.Status == "blocked" && doc.BlockedOn.Mechanism != "" {
		fmt.Fprintf(b, "**Blocked on:** %s\n\n", doc.BlockedOn.Mechanism)
	}
}

func loadArchitecture(root string) (architecture, error) {
	var arch architecture
	content, err := os.ReadFile(filepath.Join(root, "docs", "ARCHITECTURE.yaml"))
	if err != nil {
		return arch, fmt.Errorf("read docs/ARCHITECTURE.yaml: %w", err)
	}
	if err := yaml.Unmarshal(content, &arch); err != nil {
		return arch, fmt.Errorf("parse docs/ARCHITECTURE.yaml: %w", err)
	}
	if len(arch.Structure.Parts) == 0 {
		return arch, fmt.Errorf("docs/ARCHITECTURE.yaml declares no parts")
	}
	return arch, nil
}

func loadRoadMap(root string) (map[string]string, error) {
	var rm roadMap
	content, err := os.ReadFile(filepath.Join(root, "docs", "road-map.yaml"))
	if err != nil {
		return nil, fmt.Errorf("read docs/road-map.yaml: %w", err)
	}
	if err := yaml.Unmarshal(content, &rm); err != nil {
		return nil, fmt.Errorf("parse docs/road-map.yaml: %w", err)
	}
	statuses := make(map[string]string, len(rm.Chapters))
	for _, c := range rm.Chapters {
		statuses[c.ID] = c.Status
	}
	return statuses, nil
}

func loadSRD(root, path string) (srd, error) {
	var doc srd
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		return doc, fmt.Errorf("read %s: %w", path, err)
	}
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return doc, fmt.Errorf("parse %s: %w", path, err)
	}
	return doc, nil
}
