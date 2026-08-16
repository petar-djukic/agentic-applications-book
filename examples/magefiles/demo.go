// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// demoConfig is the shape of an application's demo.yaml: named steps,
// each an argv run in a directory relative to the application root.
// Steps run in order and inherit no extra environment, keeping demos
// canned -- deterministic fixtures, no credentials.
type demoConfig struct {
	Steps []demoStep `yaml:"steps"`
}

type demoStep struct {
	Name string   `yaml:"name"`
	Dir  string   `yaml:"dir,omitempty"`
	Argv []string `yaml:"argv"`
}

// Demo runs the canned demo of every chapter-application that has one,
// in manifest order. Only a planned entry is skipped: planned means no
// directory content exists yet, so there is no demo.yaml to read and
// failing on its absence would make the manifest's own planned status
// unusable. A partial entry does have a demo -- one that exercises less
// than the SRD asks for -- and skipping it would hide the evidence it
// does produce.
func Demo() error {
	ran, skipped, err := demoExamples(".")
	if err != nil {
		return err
	}
	fmt.Printf("demo: %d example(s) ran, %d skipped\n", ran, skipped)
	return nil
}

func demoExamples(examplesRoot string) (ran, skipped int, err error) {
	manifest, err := loadManifest(examplesRoot)
	if err != nil {
		return 0, 0, err
	}
	for _, entry := range manifest.Examples {
		if entry.Kind != kindChapterApplication {
			continue
		}
		if entry.Status == statusPlanned {
			fmt.Printf("demo: skip %s (status %s)\n", entry.ID, entry.Status)
			skipped++
			continue
		}
		if err := runEntryDemo(examplesRoot, entry); err != nil {
			return ran, skipped, fmt.Errorf("demo %s: %w", entry.ID, err)
		}
		ran++
	}
	return ran, skipped, nil
}

func runEntryDemo(examplesRoot string, entry Entry) error {
	root := entryDir(examplesRoot, entry)
	content, err := os.ReadFile(filepath.Join(root, "demo.yaml"))
	if err != nil {
		return fmt.Errorf("read demo.yaml: %w", err)
	}
	var config demoConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return fmt.Errorf("parse demo.yaml: %w", err)
	}
	if len(config.Steps) == 0 {
		return fmt.Errorf("demo.yaml declares no steps")
	}
	for _, step := range config.Steps {
		if len(step.Argv) == 0 {
			return fmt.Errorf("step %q has an empty argv", step.Name)
		}
		command := exec.Command(step.Argv[0], step.Argv[1:]...)
		command.Dir = filepath.Join(root, filepath.FromSlash(step.Dir))
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		fmt.Printf("demo %s: %s\n", entry.ID, step.Name)
		if err := command.Run(); err != nil {
			return fmt.Errorf("step %q: %w", step.Name, err)
		}
	}
	return nil
}
