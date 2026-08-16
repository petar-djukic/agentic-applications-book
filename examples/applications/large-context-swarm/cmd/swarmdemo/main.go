// Copyright (c) 2026 Petar Djukic
// SPDX-License-Identifier: MIT

// Command swarmdemo runs the blackboard loop against loopback stubs and
// asserts the six properties srd-large-context-swarm D3 names.
//
// The assertions are made against the stubs' request log rather than
// against the run's output. That distinction is the point: whether a
// worker sent a where_document filter, and whether corpus text ever
// reached a request the root made, are facts about what went out on the
// wire. Nothing the run returns could establish either.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const runtimePackage = "github.com/Nokia-Bell-Labs/declarative-agents/agent-core/cmd/agent"

// expectations mirrors the parts of testdata/expected.yaml this command
// reads. The file is the authority on what the demo checks, so changing
// the corpus and changing the assertions stays one edit.
type expectations struct {
	Request struct {
		Task       string `yaml:"task"`
		Collection string `yaml:"collection"`
		Anchor     string `yaml:"anchor"`
		NResults   int    `yaml:"n_results"`
		RoundBound int    `yaml:"round_bound"`
		ID         string `yaml:"id"`
		Agent      string `yaml:"agent"`
	} `yaml:"request"`
	MultiHop struct {
		AnswerContains []string `yaml:"answer_contains"`
	} `yaml:"multi_hop"`
}

type corpus struct {
	Documents []struct {
		ID   string `yaml:"id"`
		Text string `yaml:"text"`
	} `yaml:"documents"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "swarmdemo:", err)
		os.Exit(1)
	}
}

func run() error {
	expected, err := loadExpectations("testdata/expected.yaml")
	if err != nil {
		return err
	}
	docs, err := loadCorpus("testdata/corpus/documents.yaml")
	if err != nil {
		return err
	}

	chroma, err := newStub("testdata/responses/chroma.yaml")
	if err != nil {
		return err
	}
	ollama, err := newStub("testdata/responses/ollama.yaml")
	if err != nil {
		return err
	}
	stopChroma, err := serve(chromaAddr, chroma)
	if err != nil {
		return err
	}
	defer stopChroma()
	stopOllama, err := serve(ollamaAddr, ollama)
	if err != nil {
		return err
	}
	defer stopOllama()

	binary, cleanup, err := buildRuntime()
	if err != nil {
		return err
	}
	defer cleanup()

	requestPath, err := writeRequest(expected)
	if err != nil {
		return err
	}

	runErr := runRoot(binary, requestPath)

	records := append(chroma.requests(), ollama.requests()...)
	if missing := append(chroma.unmatched(), ollama.unmatched()...); len(missing) > 0 {
		return fmt.Errorf("no fixture covers: %s", strings.Join(dedupe(missing), ", "))
	}
	if runErr != nil {
		return fmt.Errorf("root run: %w", runErr)
	}
	return assertAll(records, expected, docs)
}

// buildRuntime builds the pinned agent binary into a temporary
// directory. The root dispatches workers through self_invoke, which
// launches a child process, so a real binary on disk is required --
// `go run` would leave nothing for the child boundary to exec.
func buildRuntime() (string, func(), error) {
	dir, err := os.MkdirTemp("", "swarmdemo-")
	if err != nil {
		return "", nil, err
	}
	binary := filepath.Join(dir, "agent")
	build := exec.Command("go", "build", "-o", binary, runtimePackage)
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		os.RemoveAll(dir)
		return "", nil, fmt.Errorf("build runtime: %w", err)
	}
	return binary, func() { os.RemoveAll(dir) }, nil
}

func writeRequest(expected expectations) (string, error) {
	payload := map[string]any{
		"task":        expected.Request.Task,
		"collection":  expected.Request.Collection,
		"anchor":      expected.Request.Anchor,
		"n_results":   expected.Request.NResults,
		"round_bound": expected.Request.RoundBound,
		"id":          expected.Request.ID,
		"agent":       expected.Request.Agent,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	path := filepath.Join(os.TempDir(), "swarmdemo-request.json")
	return path, os.WriteFile(path, encoded, 0o644)
}

func runRoot(binary, requestPath string) error {
	command := exec.Command(binary,
		"--profile", "agents/rlm-root/profile.yaml",
		"--request", requestPath,
		"--child-agent-binary", binary,
	)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	done := make(chan error, 1)
	if err := command.Start(); err != nil {
		return err
	}
	go func() { done <- command.Wait() }()
	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Minute):
		_ = command.Process.Kill()
		return fmt.Errorf("root run exceeded five minutes")
	}
}

func loadExpectations(path string) (expectations, error) {
	var out expectations
	content, err := os.ReadFile(path)
	if err != nil {
		return out, fmt.Errorf("read %s: %w", path, err)
	}
	return out, yaml.Unmarshal(content, &out)
}

func loadCorpus(path string) (corpus, error) {
	var out corpus
	content, err := os.ReadFile(path)
	if err != nil {
		return out, fmt.Errorf("read %s: %w", path, err)
	}
	return out, yaml.Unmarshal(content, &out)
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
