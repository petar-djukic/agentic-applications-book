// Copyright (c) 2026 Petar Djukic
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// The loopback stubs. corpus-rest.yaml names 127.0.0.1:11434 and
// 127.0.0.1:8000 and allows no other host, and invoke_llm's
// provider_url defaults to the same Ollama address, so serving both
// there is what keeps the demo offline without editing a copied
// declaration.
const (
	ollamaAddr = "127.0.0.1:11434"
	chromaAddr = "127.0.0.1:8000"
)

// fixture mirrors the response-script files under testdata/responses/.
// The shape is agent-core's mock REST binding: routes keyed by method
// and literal path, each an ordered script whose last entry repeats
// once exhausted.
type fixture struct {
	Routes []route `yaml:"routes"`
}

type route struct {
	Method    string     `yaml:"method"`
	Path      string     `yaml:"path"`
	Responses []response `yaml:"responses"`
}

type response struct {
	Status int `yaml:"status"`
	Body   any `yaml:"body"`
}

// record is one request the stub received. The demo's assertions are
// made against these rather than against what the run returned: whether
// a worker sent a where_document filter, and whether any root request
// carried document text, are properties of what went out.
type record struct {
	Method string
	Path   string
	Body   []byte
}

type stub struct {
	mu      sync.Mutex
	scripts map[string][]response // method+path -> remaining script
	cursor  map[string]int
	log     []record
	unknown []string
}

func newStub(paths ...string) (*stub, error) {
	s := &stub{scripts: map[string][]response{}, cursor: map[string]int{}}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read fixture %s: %w", path, err)
		}
		var f fixture
		if err := yaml.Unmarshal(content, &f); err != nil {
			return nil, fmt.Errorf("parse fixture %s: %w", path, err)
		}
		for _, r := range f.Routes {
			if len(r.Responses) == 0 {
				return nil, fmt.Errorf("%s: route %s %s declares no responses", path, r.Method, r.Path)
			}
			s.scripts[r.Method+" "+r.Path] = r.Responses
		}
	}
	return s, nil
}

func (s *stub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	key := r.Method + " " + r.URL.Path

	s.mu.Lock()
	s.log = append(s.log, record{Method: r.Method, Path: r.URL.Path, Body: body})
	script, ok := s.scripts[key]
	if !ok {
		s.unknown = append(s.unknown, key)
		s.mu.Unlock()
		// A route the fixtures do not cover is a defect in the
		// fixtures, not a transient failure. Say which one, and fail
		// the request so the run stops rather than proceeding on a
		// response nobody wrote.
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprintf(w, `{"error":"no fixture for %s"}`, key)
		return
	}
	index := s.cursor[key]
	if index >= len(script) {
		index = len(script) - 1 // last entry repeats
	}
	s.cursor[key] = index + 1
	chosen := script[index]
	s.mu.Unlock()

	payload, err := json.Marshal(chosen.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(chosen.Status)
	_, _ = w.Write(payload)
}

// requests returns a copy of the recorded log.
func (s *stub) requests() []record {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]record, len(s.log))
	copy(out, s.log)
	return out
}

func (s *stub) unmatched() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.unknown))
	copy(out, s.unknown)
	return out
}

// serve starts one listener and returns a shutdown func. Binding
// explicitly rather than through ListenAndServe so a port already in
// use fails here, with the address named, instead of racing the run.
func serve(addr string, handler http.Handler) (func(), error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("bind %s: %w (is a real Chroma or Ollama already running?)", addr, err)
	}
	server := &http.Server{Handler: handler}
	go func() { _ = server.Serve(listener) }()
	return func() { _ = server.Close() }, nil
}

// bodyContains reports whether any recorded request body holds needle.
func bodyContains(records []record, needle string) bool {
	for _, r := range records {
		if bytes.Contains(r.Body, []byte(needle)) {
			return true
		}
	}
	return false
}
