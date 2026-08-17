// Copyright (c) 2026 Nokia
// SPDX-License-Identifier: BSD-3-Clause
//
// Adopted from declarative-agents applications/catalog/conformance/
// serve.go at fork commit c87a5322 (GH-40): the shipped-profile staging
// subset only (CopyShippedProfile and its closure copier). The fork
// file's async serve harness stays behind -- the blackboard tests do
// not use it. The staging code is unchanged.

package conformance

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// CopyShippedProfile stages a shipped profile under its catalog-relative path
// so a family test can run the wrapper an operator actually ships rather than
// a synthesized reconstruction. relProfile is relative to the catalog root
// (e.g. "agents/runtime-state-reader/profile.yaml").
//
// The requested profile directory is copied recursively. YAML references that
// leave that directory are then copied transitively under the same
// catalog-relative paths; a referenced sibling profile brings its whole family
// directory. This preserves package-root references such as
// agents/critic/profile.yaml without copying unrelated generated catalog
// artifacts. /opt/agent-core references need no staging because --core-root
// remaps them onto the checkout (spec.SetAgentCoreInstallRoot).
//
// patches applies simultaneous exact string replacements only within the
// requested family for the few values the harness must control (chiefly
// hard-coded listen addresses). Transitive dependencies remain byte-identical.
func CopyShippedProfile(t *testing.T, relProfile string, patches map[string]string) string {
	t.Helper()
	root := ProfilesRoot()
	srcProfile := filepath.Clean(ProfilePath(relProfile))
	if err := requirePathWithin(root, srcProfile); err != nil {
		t.Fatalf("copy shipped profile %s: %v", relProfile, err)
	}
	replacer, err := newProfileReplacer(patches)
	if err != nil {
		t.Fatalf("copy shipped profile %s: %v", relProfile, err)
	}
	stage := &shippedProfileStage{
		sourceRoot: root,
		targetRoot: t.TempDir(),
		patchRoot:  filepath.Dir(srcProfile),
		replacer:   replacer,
		copied:     make(map[string]bool),
	}
	if err := stage.copyClosure(filepath.Dir(srcProfile)); err != nil {
		t.Fatalf("copy shipped profile %s: %v", relProfile, err)
	}
	relative, _ := filepath.Rel(root, srcProfile)
	return filepath.Join(stage.targetRoot, relative)
}

type shippedProfileStage struct {
	sourceRoot string
	targetRoot string
	patchRoot  string
	replacer   *strings.Replacer
	copied     map[string]bool
	pending    []string
}

var shippedProfileTemplatePattern = regexp.MustCompile(`\$\{[^}\r\n]+\}`)

func (s *shippedProfileStage) copyClosure(initial string) error {
	s.pending = append(s.pending, initial)
	for len(s.pending) > 0 {
		source := s.pending[0]
		s.pending = s.pending[1:]
		if s.copied[source] {
			continue
		}
		info, err := os.Lstat(source)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if err := s.copyDirectory(source); err != nil {
				return err
			}
			continue
		}
		if info.Mode().IsRegular() {
			if err := s.copyFile(source); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *shippedProfileStage) copyDirectory(source string) error {
	s.copied[source] = true
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		return s.copyFile(path)
	})
}

func (s *shippedProfileStage) copyFile(source string) error {
	source = filepath.Clean(source)
	if s.copied[source] {
		return nil
	}
	if err := requirePathWithin(s.sourceRoot, source); err != nil {
		return err
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(s.sourceRoot, source)
	if err != nil {
		return err
	}
	target := filepath.Join(s.targetRoot, relative)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	content := string(data)
	if requirePathWithin(s.patchRoot, source) == nil {
		content = s.replacer.Replace(content)
	}
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		return err
	}
	s.copied[source] = true
	if filepath.Ext(source) == ".yaml" || filepath.Ext(source) == ".yml" {
		return s.enqueueYAMLReferences(source, data)
	}
	return nil
}

func (s *shippedProfileStage) enqueueYAMLReferences(source string, data []byte) error {
	var document yaml.Node
	inspectable := shippedProfileTemplatePattern.ReplaceAll(data, []byte("template_value"))
	if err := yaml.Unmarshal(inspectable, &document); err != nil {
		return fmt.Errorf("parse YAML dependencies in %s: %w", source, err)
	}
	var visit func(*yaml.Node)
	visit = func(node *yaml.Node) {
		if node.Kind == yaml.ScalarNode {
			if dependency := s.resolveYAMLDependency(source, node.Value); dependency != "" {
				if isProfileFilename(filepath.Base(dependency)) {
					dependency = filepath.Dir(dependency)
				}
				s.pending = append(s.pending, dependency)
			}
		}
		for _, child := range node.Content {
			visit(child)
		}
	}
	visit(&document)
	return nil
}

func (s *shippedProfileStage) resolveYAMLDependency(source, value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, "*?[") ||
		strings.HasPrefix(value, "/opt/agent-core/") {
		return ""
	}
	pathLike := strings.HasSuffix(value, ".yaml") || strings.HasSuffix(value, ".yml") ||
		strings.HasPrefix(value, "agents/") || strings.HasPrefix(value, "testdata/") ||
		strings.HasPrefix(value, "./") || strings.HasPrefix(value, "../")
	if !pathLike || filepath.IsAbs(value) {
		return ""
	}
	var candidate string
	if strings.HasPrefix(value, "agents/") || strings.HasPrefix(value, "testdata/") {
		candidate = filepath.Join(s.sourceRoot, filepath.FromSlash(value))
	} else {
		candidate = filepath.Join(filepath.Dir(source), filepath.FromSlash(value))
	}
	candidate = filepath.Clean(candidate)
	if requirePathWithin(s.sourceRoot, candidate) != nil {
		return ""
	}
	if info, err := os.Stat(candidate); err == nil && (info.IsDir() || info.Mode().IsRegular()) {
		return candidate
	}
	return ""
}

func requirePathWithin(root, candidate string) error {
	relative, err := filepath.Rel(filepath.Clean(root), filepath.Clean(candidate))
	if err != nil {
		return fmt.Errorf("compare catalog path: %w", err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) ||
		filepath.IsAbs(relative) {
		return fmt.Errorf("path escapes catalog root: %s", candidate)
	}
	return nil
}

func isProfileFilename(name string) bool {
	return name == "profile.yaml" ||
		strings.HasPrefix(name, "profile-") && strings.HasSuffix(name, ".yaml") ||
		strings.HasSuffix(name, "-profile.yaml")
}

func newProfileReplacer(patches map[string]string) (*strings.Replacer, error) {
	keys := make([]string, 0, len(patches))
	for old := range patches {
		if old == "" {
			return nil, fmt.Errorf("profile patch contains an empty match")
		}
		keys = append(keys, old)
	}
	sort.Strings(keys)
	for i, key := range keys {
		for _, other := range keys[i+1:] {
			if strings.Contains(other, key) {
				return nil, fmt.Errorf("profile patches %q and %q overlap", key, other)
			}
		}
	}
	pairs := make([]string, 0, len(keys)*2)
	for _, old := range keys {
		pairs = append(pairs, old, patches[old])
	}
	return strings.NewReplacer(pairs...), nil
}
