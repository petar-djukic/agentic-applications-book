// Copyright (c) 2026 Petar Djukic
// SPDX-License-Identifier: MIT

// Command swarmlive runs the blackboard loop against a real local
// Chroma and Ollama (srd-large-context-swarm X2). It is reached only
// through the opt-in integration:swarm target, never by audit, test, or
// demo.
//
// The ingest path is the answer to #29: each fixture document goes in
// through the shipped memory-write block -- a real embedding from the
// live Ollama, a real add to the live Chroma -- tagged source corpus,
// agent ingest, round 0. No new catalog block is needed; fixture
// loading is what the already-copied write path does.
//
// What a live run proves is scoped honestly (A-6, X3): the loop runs on
// a real model and real storage, the collection lifecycle holds, the
// workers write provenance-tagged findings, and a Final entry lands
// under the request's id. Answer quality is not asserted, and the
// handle-discipline invariant I1 is not provable here -- that assertion
// needs the canned demo's wire log.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	runtimePackage = "github.com/Nokia-Bell-Labs/declarative-agents/agent-core/cmd/agent"

	// The addresses the copied corpus-rest.yaml declares. A live run
	// needs no staging: real servers on the declared ports are exactly
	// what the shipped profiles bind.
	chromaRoot = "http://127.0.0.1:8000"
	ollamaRoot = "http://127.0.0.1:11434"
	chromaBase = chromaRoot + "/api/v2/tenants/default_tenant/databases/default_database"

	// The models the shipped declarations name.
	chatModel  = "qwen3:8b"
	embedModel = "qwen3-embedding:8b"
)

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
}

type corpus struct {
	Documents []struct {
		ID   string `yaml:"id"`
		Text string `yaml:"text"`
	} `yaml:"documents"`
}

func main() {
	code, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "swarmlive:", err)
		os.Exit(1)
	}
	os.Exit(code)
}

func run() (int, error) {
	if reason := gate(); reason != "" {
		// A missing dependency is a recorded skip, not a failure: the
		// target must not fail a checkout without local servers (X2).
		fmt.Println("swarmlive: SKIP -", reason)
		return 0, nil
	}

	expected, err := loadExpectations("testdata/expected.yaml")
	if err != nil {
		return 0, err
	}
	docs, err := loadCorpus("testdata/corpus/documents.yaml")
	if err != nil {
		return 0, err
	}

	binary, cleanup, err := buildRuntime()
	if err != nil {
		return 0, err
	}
	defer cleanup()

	collection := fmt.Sprintf("%s-live-%d", expected.Request.Collection, os.Getpid())
	// The collection is torn down whatever happens after this point: a
	// leaked collection is a defect, not residue (I3).
	defer func() { _ = deleteCollection(collection) }()

	fmt.Printf("swarmlive: ingesting %d fixture documents through memory-write\n", len(docs.Documents))
	for _, doc := range docs.Documents {
		if err := ingestDocument(binary, collection, doc.ID, doc.Text); err != nil {
			return 0, fmt.Errorf("ingest %s: %w", doc.ID, err)
		}
	}

	fmt.Println("swarmlive: running rlm-root against live servers")
	if err := runRoot(binary, expected, collection); err != nil {
		return 0, fmt.Errorf("root run: %w", err)
	}

	if err := assertLive(expected, docs, collection); err != nil {
		return 0, err
	}
	fmt.Println("swarmlive: PASS - live loop ran; lifecycle, provenance, and the Final entry checked (answer quality is not asserted)")
	return 0, nil
}

// gate reports why a live run cannot proceed, or "" when it can. The
// probes mirror what the shipped declarations will actually resolve.
func gate() string {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(chromaRoot + "/api/v2/heartbeat")
	if err != nil {
		return fmt.Sprintf("Chroma is unavailable at %s: %v; start it with `chroma run --path <dir>` and rerun", chromaRoot, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Chroma heartbeat at %s returned %d", chromaRoot, resp.StatusCode)
	}
	for _, model := range []string{chatModel, embedModel} {
		if reason := probeOllamaModel(client, model); reason != "" {
			return reason
		}
	}
	return ""
}

func probeOllamaModel(client *http.Client, model string) string {
	resp, err := client.Get(ollamaRoot + "/api/tags")
	if err != nil {
		return fmt.Sprintf("Ollama is unavailable at %s: %v", ollamaRoot, err)
	}
	defer resp.Body.Close()
	var payload struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Sprintf("decode Ollama tags: %v", err)
	}
	for _, m := range payload.Models {
		if strings.EqualFold(m.Name, model) || strings.EqualFold(m.Name, model+":latest") {
			return ""
		}
	}
	return fmt.Sprintf("the Ollama model %q is not pulled; `ollama pull %s` and rerun", model, model)
}

// ingestDocument runs the shipped memory-write block for one fixture
// document: live embed, live add, provenance tagged as corpus/ingest.
func ingestDocument(binary, collection, id, text string) error {
	entry := map[string]any{
		"content":    text,
		"id":         id,
		"collection": collection,
		"source":     "corpus",
		"agent":      "ingest",
		"round":      0,
	}
	requestPath, err := writeJSON(entry, "swarmlive-ingest-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(requestPath)

	appDir, err := os.Getwd()
	if err != nil {
		return err
	}
	profile := filepath.Join(appDir, "..", "..", "catalog", "agents", "knowledge-manager", "memory-write", "profile.yaml")
	return runAgent(binary, profile, requestPath, nil, 2*time.Minute)
}

func runRoot(binary string, expected expectations, collection string) error {
	payload := map[string]any{
		"task":        expected.Request.Task,
		"collection":  collection,
		"anchor":      expected.Request.Anchor,
		"n_results":   expected.Request.NResults,
		"round_bound": expected.Request.RoundBound,
		"id":          expected.Request.ID,
		"agent":       expected.Request.Agent,
	}
	requestPath, err := writeJSON(payload, "swarmlive-request-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(requestPath)

	appDir, err := os.Getwd()
	if err != nil {
		return err
	}
	env := []string{"RLM_AGENT_BIN=" + binary, "RLM_APP_DIR=" + appDir}
	profile := filepath.Join(appDir, "agents", "rlm-root", "profile.yaml")
	// A live model on local hardware is slow; the bound is generous
	// where the canned demo's is tight.
	return runAgent(binary, profile, requestPath, env, 15*time.Minute)
}

func runAgent(binary, profile, requestPath string, extraEnv []string, timeout time.Duration) error {
	workDir, err := os.MkdirTemp("", "swarmlive-work-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workDir)
	command := exec.Command(binary, "--profile", profile, "--request", requestPath, "--directory", workDir)
	command.Dir = workDir
	command.Env = append(os.Environ(), extraEnv...)
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
	case <-time.After(timeout):
		_ = command.Process.Kill()
		return fmt.Errorf("agent run exceeded %s", timeout)
	}
}

// --- live assertions ---------------------------------------------------

// assertLive checks what a live run can prove: every corpus document is
// in the collection, at least one worker finding carries the derived
// provenance, and the Final entry landed under the request's id.
func assertLive(expected expectations, docs corpus, collection string) error {
	id, err := collectionID(collection)
	if err != nil {
		return err
	}
	corpusIDs, err := getIDs(id, map[string]any{"source": "corpus"})
	if err != nil {
		return err
	}
	if len(corpusIDs) != len(docs.Documents) {
		return fmt.Errorf("live A1: %d corpus entries in the collection, want %d", len(corpusIDs), len(docs.Documents))
	}
	fmt.Printf("ok   live A1: all %d corpus entries present\n", len(corpusIDs))

	findings, err := getIDs(id, map[string]any{"agent": "rlm-worker"})
	if err != nil {
		return err
	}
	if len(findings) == 0 {
		return fmt.Errorf("live A2: no entry tagged agent rlm-worker; the workers wrote nothing")
	}
	fmt.Printf("ok   live A2: %d provenance-tagged worker finding(s)\n", len(findings))

	finalIDs, err := getIDs(id, map[string]any{"agent": expected.Request.Agent})
	if err != nil {
		return err
	}
	found := false
	for _, f := range finalIDs {
		if f == expected.Request.ID {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("live A3: no Final entry with id %s tagged agent %s", expected.Request.ID, expected.Request.Agent)
	}
	fmt.Println("ok   live A3: the Final entry landed under the request's id")
	return nil
}

func collectionID(name string) (string, error) {
	body, err := postJSON(chromaBase+"/collections", map[string]any{"name": name, "get_or_create": true})
	if err != nil {
		return "", fmt.Errorf("resolve collection %s: %w", name, err)
	}
	var resolved struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &resolved); err != nil || resolved.ID == "" {
		return "", fmt.Errorf("resolve collection %s: no id in %s", name, body)
	}
	return resolved.ID, nil
}

// getIDs retrieves the record ids matching a metadata filter through
// Chroma's filter-only get endpoint -- no embedding needed to read back
// what the run wrote.
func getIDs(collectionID string, where map[string]any) ([]string, error) {
	body, err := postJSON(chromaBase+"/collections/"+collectionID+"/get",
		map[string]any{"where": where, "include": []string{"metadatas"}})
	if err != nil {
		return nil, err
	}
	var parsed struct {
		IDs []string `json:"ids"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode get response: %w", err)
	}
	return parsed.IDs, nil
}

func deleteCollection(name string) error {
	_, err := postJSON(chromaBase+"/collections/"+name+"/delete", map[string]any{})
	return err
}

func postJSON(url string, payload any) ([]byte, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%s returned %d: %s", url, resp.StatusCode, buf.String())
	}
	return buf.Bytes(), nil
}

func writeJSON(payload any, pattern string) (string, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	if _, err := file.Write(encoded); err != nil {
		file.Close()
		return "", err
	}
	return file.Name(), file.Close()
}

func buildRuntime() (string, func(), error) {
	dir, err := os.MkdirTemp("", "swarmlive-")
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
