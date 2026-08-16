// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/magefile/mage/sh"
)

const (
	bookSlug   = "agentic-applications"
	outputDir  = "generated-files"
	figuresDir = "figures"
	csl        = "templates/ieee.csl"
	template   = "templates/eisvogel.latex"
)

// All generates figures then builds the PDF (default target).
func All() error {
	if err := Figures(); err != nil {
		return err
	}
	return PDF()
}

// Figures renders all PlantUML diagrams to PNG.
func Figures() error {
	pumls, err := filepath.Glob(filepath.Join(figuresDir, "*.puml"))
	if err != nil {
		return err
	}
	if len(pumls) == 0 {
		fmt.Println("no .puml files found")
		return nil
	}
	sort.Strings(pumls)
	for _, puml := range pumls {
		png := strings.TrimSuffix(puml, ".puml") + ".png"
		pumlInfo, err := os.Stat(puml)
		if err != nil {
			return err
		}
		if pngInfo, err := os.Stat(png); err == nil && pngInfo.ModTime().After(pumlInfo.ModTime()) {
			continue
		}
		fmt.Printf("plantuml %s\n", puml)
		if err := sh.Run("plantuml", "-tpng", "-o", ".", puml); err != nil {
			return fmt.Errorf("plantuml %s: %w", puml, err)
		}
	}
	return nil
}

// PDF compiles the numbered markdown chapters into a PDF via the
// Eisvogel pandoc template.
func PDF() error {
	if err := Figures(); err != nil {
		return err
	}

	mds, err := discoverMarkdownChapters()
	if err != nil {
		return err
	}
	if len(mds) == 0 {
		return fmt.Errorf("no [0-9][0-9]-*.md chapter files found")
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", outputDir, err)
	}

	date := time.Now().Format("2006-01-02")
	out := filepath.Join(outputDir, fmt.Sprintf("%s-v%s.pdf", bookSlug, date))

	args := []string{
		"--citeproc",
		"--csl=" + csl,
		"--bibliography=references.yaml",
		"--template=" + template,
		"--from", "markdown",
		"--pdf-engine=xelatex",
		"--listings",
	}
	args = append(args, mds...)
	args = append(args, "-o", out)

	fmt.Printf("generating %s from %d chapters\n", out, len(mds))
	if err := sh.Run("pandoc", args...); err != nil {
		return fmt.Errorf("pandoc: %w", err)
	}
	return nil
}

// Outline generates the book outline from docs/ARCHITECTURE.yaml,
// docs/srd/, and docs/road-map.yaml, and renders it to PDF through the
// same Eisvogel pipeline as the book. The markdown intermediate is kept
// beside the PDF so a review can diff it.
func Outline() error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", outputDir, err)
	}
	date := time.Now().Format("2006-01-02")
	md := filepath.Join(outputDir, fmt.Sprintf("%s-outline-%s.md", bookSlug, date))
	pdf := filepath.Join(outputDir, fmt.Sprintf("%s-outline-%s.pdf", bookSlug, date))

	document, err := sh.Output("go", "-C", "cmd/genoutline", "run", ".", "../..")
	if err != nil {
		return fmt.Errorf("genoutline: %w", err)
	}
	if err := os.WriteFile(md, []byte(document+"\n"), 0o644); err != nil {
		return err
	}

	fmt.Printf("generating %s\n", pdf)
	if err := sh.Run("pandoc",
		"--template="+template,
		"--from", "markdown",
		"--pdf-engine=xelatex",
		md, "-o", pdf,
	); err != nil {
		return fmt.Errorf("pandoc: %w", err)
	}
	return nil
}

// Critic reviews drafted chapters against the constitutions and their
// SRDs: deterministic form checks always, plus an LLM critique when
// ANTHROPIC_API_KEY is set (a keyless run skips it with a notice). Pass
// a chapter id to review one chapter, or "all" to review every drafted
// chapter. Exits non-zero on a blocking finding; never edits.
func Critic(chapter string) error {
	args := []string{"-C", "cmd/critic", "run", ".", "../.."}
	if chapter != "all" {
		args = append(args, chapter)
	}
	return sh.RunV("go", args...)
}

// Clean removes generated PNGs and PDFs.
func Clean() error {
	pngs, _ := filepath.Glob(filepath.Join(figuresDir, "*.png"))
	for _, f := range pngs {
		fmt.Printf("rm %s\n", f)
		os.Remove(f)
	}
	entries, err := os.ReadDir(outputDir)
	if err == nil {
		for _, e := range entries {
			path := filepath.Join(outputDir, e.Name())
			fmt.Printf("rm %s\n", path)
			os.Remove(path)
		}
	}
	return nil
}

func discoverMarkdownChapters() ([]string, error) {
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}
	var mds []string
	for _, e := range entries {
		name := e.Name()
		if len(name) >= 3 && name[0] >= '0' && name[0] <= '9' && name[1] >= '0' && name[1] <= '9' && name[2] == '-' && strings.HasSuffix(name, ".md") {
			mds = append(mds, name)
		}
	}
	sort.Strings(mds)
	return mds, nil
}
