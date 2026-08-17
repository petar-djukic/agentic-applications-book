// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
)

// Integration holds the opt-in live targets. Nothing here runs from
// Audit, Test, or Demo (VISION N2): each target needs local servers and
// reports SKIP with a recorded reason when they are absent.
type Integration mg.Namespace

// BlackboardMemory writes a provenance-tagged entry through the shipped
// memory-write block against a live local Chroma and Ollama, then reads
// it back by a metadata filter and by an exact substring of its
// content. The conformance case skips with a recorded reason when
// either server is absent, so this target reports SKIP rather than
// failing a checkout without the local dependencies. Adapted from the
// declarative-agents fork's integration_blackboard.go at c87a5322
// (GH-40).
func (Integration) BlackboardMemory() error {
	cmd := exec.Command("go", "-C", "conformance", "test",
		"-run", "^TestBlackboardMemoryLiveRoundtrip$", "-count=1", "-v", ".", "-live")
	var transcript bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &transcript)
	cmd.Stderr = io.MultiWriter(os.Stderr, &transcript)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run shipped blackboard write and filtered read: %w", err)
	}
	// A skipped case is a reported dependency gap, not evidence. Saying
	// PASS for both outcomes is how a target starts claiming proof it
	// never produced.
	if strings.Contains(transcript.String(), "--- SKIP") {
		fmt.Println("integration:blackboardMemory SKIP - the live roundtrip did not run; see the recorded reason above")
		return nil
	}
	fmt.Println("integration:blackboardMemory PASS - shipped memory-write stored a tagged entry retrieved by metadata and by exact substring")
	return nil
}

// Swarm runs the blackboard loop against a real local Chroma and
// Ollama (srd-large-context-swarm X2): fixture corpus ingested through
// the shipped memory-write block with live embeddings, rlm-root run on
// the live chat model, lifecycle and provenance asserted, collection
// torn down. Skips with a recorded reason when a server or model is
// absent; never reached by audit, test, or demo.
func (Integration) Swarm() error {
	cmd := exec.Command("go", "run", "./cmd/swarmlive")
	cmd.Dir = filepath.Join("applications", "large-context-swarm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run the live swarm loop: %w", err)
	}
	return nil
}
