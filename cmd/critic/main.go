// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

// Command critic reviews drafted chapters against the constitutions and
// each chapter's SRD. It runs two layers. Layer 1 is deterministic: the
// CHECK rules that a program can verify by inspection -- the retest box
// (voice.yaml V-R1), the banned words (V-B1), the excerpt length cap
// (V-S3), the create pointer and takeaway question (V-S6, V-S7), and
// the verdict taxonomy (argument.yaml A-3). Layer 2 sends the chapter,
// its SRD, and the constitutions to Claude for the READ rules a program
// cannot judge; it runs only when ANTHROPIC_API_KEY is set, so a
// keyless environment skips it with a notice instead of failing.
//
// The critic never edits a chapter. It prints findings anchored to a
// rule id and a location, and exits non-zero when any finding is
// blocking. Chapters with road-map status other than "drafted" are
// reported and skipped; with nothing drafted the critic exits zero.
//
// Usage: critic <root> [chapter-id]
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"gopkg.in/yaml.v3"
)

const (
	criticModel     = anthropic.ModelClaudeSonnet5
	excerptLineCap  = 40
	severityBlock   = "blocking"
	severityAdvise  = "advisory"
	retestBoxLabel  = "**The retest.**"
	takeawayWindow  = 15 // prose lines from the end searched for the takeaway question
	criticMaxTokens = 16000
)

// finding is one critic result, anchored to a rule and a location.
type finding struct {
	Rule     string // constitution rule id, e.g. V-B1
	File     string
	Line     int    // 0 when the finding is file-level
	Severity string // severityBlock or severityAdvise
	Text     string
}

func (f finding) String() string {
	loc := f.File
	if f.Line > 0 {
		loc = fmt.Sprintf("%s:%d", f.File, f.Line)
	}
	return fmt.Sprintf("%s: %s [%s]: %s", loc, f.Rule, f.Severity, f.Text)
}

// --- docs-tree loading -------------------------------------------------

type architecture struct {
	Structure struct {
		Parts []struct {
			Chapters []rosterChapter `yaml:"chapters"`
		} `yaml:"parts"`
	} `yaml:"structure"`
}

type rosterChapter struct {
	ID   string `yaml:"id"`
	File string `yaml:"file"`
	SRD  string `yaml:"srd"`
}

type roadMap struct {
	Chapters []struct {
		ID     string `yaml:"id"`
		Status string `yaml:"status"`
	} `yaml:"chapters"`
}

type srdDoc struct {
	Meta struct {
		PaperFields string `yaml:"paper_fields"`
	} `yaml:"meta"`
}

// voiceDoc pulls the banned-word list out of voice.yaml so the critic
// never carries its own copy of V-B1.
type voiceDoc struct {
	Banned []struct {
		ID    string   `yaml:"id"`
		Words []string `yaml:"words"`
	} `yaml:"banned"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: critic <root> [chapter-id]")
		os.Exit(2)
	}
	root := os.Args[1]
	chapterID := ""
	if len(os.Args) > 2 {
		chapterID = os.Args[2]
	}
	code, err := run(root, chapterID)
	if err != nil {
		fmt.Fprintln(os.Stderr, "critic:", err)
		os.Exit(2)
	}
	os.Exit(code)
}

func run(root, chapterID string) (int, error) {
	roster, err := loadRoster(root)
	if err != nil {
		return 0, err
	}
	statuses, err := loadStatuses(root)
	if err != nil {
		return 0, err
	}
	bannedWords, err := loadBannedWords(root)
	if err != nil {
		return 0, err
	}

	var selected []rosterChapter
	if chapterID != "" {
		ch, ok := roster[chapterID]
		if !ok {
			return 0, fmt.Errorf("chapter id %q is not in the docs/ARCHITECTURE.yaml roster", chapterID)
		}
		if statuses[ch.ID] != "drafted" {
			fmt.Printf("chapter %s is not drafted (status: %s); nothing to review\n", ch.ID, statusOrNone(statuses[ch.ID]))
			return 0, nil
		}
		selected = append(selected, ch)
	} else {
		for _, ch := range orderedRoster(root) {
			if statuses[ch.ID] == "drafted" {
				selected = append(selected, ch)
			}
		}
		if len(selected) == 0 {
			fmt.Println("no drafted chapters; nothing to review")
			return 0, nil
		}
	}

	blocking := 0
	for _, ch := range selected {
		findings, err := reviewChapter(root, ch, bannedWords)
		if err != nil {
			return 0, err
		}
		fmt.Printf("== %s (%s): %d finding(s)\n", ch.ID, ch.File, len(findings))
		for _, f := range findings {
			fmt.Println("  " + f.String())
			if f.Severity == severityBlock {
				blocking++
			}
		}
	}
	if blocking > 0 {
		fmt.Printf("%d blocking finding(s)\n", blocking)
		return 1, nil
	}
	return 0, nil
}

func statusOrNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}

// orderedRoster returns roster chapters in book order. Errors were
// already surfaced by loadRoster, so this pass ignores them.
func orderedRoster(root string) []rosterChapter {
	var arch architecture
	content, _ := os.ReadFile(filepath.Join(root, "docs", "ARCHITECTURE.yaml"))
	_ = yaml.Unmarshal(content, &arch)
	var out []rosterChapter
	for _, p := range arch.Structure.Parts {
		out = append(out, p.Chapters...)
	}
	return out
}

func loadRoster(root string) (map[string]rosterChapter, error) {
	var arch architecture
	content, err := os.ReadFile(filepath.Join(root, "docs", "ARCHITECTURE.yaml"))
	if err != nil {
		return nil, fmt.Errorf("read docs/ARCHITECTURE.yaml: %w", err)
	}
	if err := yaml.Unmarshal(content, &arch); err != nil {
		return nil, fmt.Errorf("parse docs/ARCHITECTURE.yaml: %w", err)
	}
	roster := make(map[string]rosterChapter)
	for _, p := range arch.Structure.Parts {
		for _, ch := range p.Chapters {
			roster[ch.ID] = ch
		}
	}
	if len(roster) == 0 {
		return nil, fmt.Errorf("docs/ARCHITECTURE.yaml declares no chapters")
	}
	return roster, nil
}

func loadStatuses(root string) (map[string]string, error) {
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

func loadBannedWords(root string) ([]string, error) {
	var doc voiceDoc
	content, err := os.ReadFile(filepath.Join(root, "docs", "constitutions", "voice.yaml"))
	if err != nil {
		return nil, fmt.Errorf("read docs/constitutions/voice.yaml: %w", err)
	}
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return nil, fmt.Errorf("parse docs/constitutions/voice.yaml: %w", err)
	}
	for _, b := range doc.Banned {
		if b.ID == "V-B1" {
			return b.Words, nil
		}
	}
	return nil, fmt.Errorf("docs/constitutions/voice.yaml carries no V-B1 word list")
}

// --- per-chapter review ------------------------------------------------

func reviewChapter(root string, ch rosterChapter, bannedWords []string) ([]finding, error) {
	chapterPath := filepath.Join(root, ch.File)
	content, err := os.ReadFile(chapterPath)
	if err != nil {
		return nil, fmt.Errorf("chapter %s: read %s: %w", ch.ID, ch.File, err)
	}
	text := string(content)

	paperForm := true
	if ch.SRD != "" {
		var doc srdDoc
		srdContent, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(ch.SRD)))
		if err != nil {
			return nil, fmt.Errorf("chapter %s: read %s: %w", ch.ID, ch.SRD, err)
		}
		if err := yaml.Unmarshal(srdContent, &doc); err != nil {
			return nil, fmt.Errorf("chapter %s: parse %s: %w", ch.ID, ch.SRD, err)
		}
		paperForm = doc.Meta.PaperFields != "not-applicable"
	}

	findings := layerOne(ch.File, text, bannedWords, paperForm)

	llmFindings, err := layerTwo(root, ch, text)
	if err != nil {
		return nil, err
	}
	findings = append(findings, llmFindings...)
	return findings, nil
}

// --- layer 1: deterministic form checks --------------------------------

// layerOne runs every CHECK rule a program can verify. paperForm gates
// the Implement-a-Paper form rules: substrate chapters (SRD
// paper_fields: not-applicable) carry no retest box, verdict, create
// pointer, or takeaway question.
func layerOne(file, text string, bannedWords []string, paperForm bool) []finding {
	var findings []finding
	lines := strings.Split(text, "\n")
	findings = append(findings, checkBannedWords(file, lines, bannedWords)...)
	findings = append(findings, checkExcerptLength(file, lines)...)
	if paperForm {
		findings = append(findings, checkRetestBox(file, lines)...)
		findings = append(findings, checkVerdict(file, text)...)
		findings = append(findings, checkCreatePointer(file, lines)...)
		findings = append(findings, checkTakeaway(file, lines)...)
	}
	return findings
}

// proseLines yields the chapter's prose: every line outside fenced code
// blocks. Fences toggle on lines starting with three backticks.
func proseLines(lines []string) []int {
	var idx []int
	inFence := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			continue
		}
		if !inFence {
			idx = append(idx, i)
		}
	}
	return idx
}

// checkBannedWords flags V-B1 words in prose. A match inside double
// quotes is a term of art quoted from a source and stays (the rule's
// own note), so quoted spans are skipped.
func checkBannedWords(file string, lines []string, words []string) []finding {
	var findings []finding
	patterns := make([]*regexp.Regexp, len(words))
	for i, w := range words {
		patterns[i] = regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(w) + `\b`)
	}
	for _, i := range proseLines(lines) {
		line := lines[i]
		for j, p := range patterns {
			for _, loc := range p.FindAllStringIndex(line, -1) {
				if insideQuotes(line, loc[0]) {
					continue
				}
				findings = append(findings, finding{
					Rule: "V-B1", File: file, Line: i + 1, Severity: severityBlock,
					Text: fmt.Sprintf("banned word %q in prose", words[j]),
				})
			}
		}
	}
	return findings
}

// insideQuotes reports whether position pos on line falls inside a
// double-quoted span (straight or curly quotes).
func insideQuotes(line string, pos int) bool {
	depth := 0
	for i, r := range line {
		if i >= pos {
			break
		}
		switch r {
		case '"':
			depth ^= 1
		case '“': // left curly quote
			depth = 1
		case '”': // right curly quote
			depth = 0
		}
	}
	return depth == 1
}

// checkExcerptLength flags V-S3: fenced blocks longer than the ~40-line
// cap. The finding anchors to the opening fence.
func checkExcerptLength(file string, lines []string) []finding {
	var findings []finding
	inFence := false
	fenceStart := 0
	count := 0
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			if !inFence {
				inFence = true
				fenceStart = i + 1
				count = 0
			} else {
				inFence = false
				if count > excerptLineCap {
					findings = append(findings, finding{
						Rule: "V-S3", File: file, Line: fenceStart, Severity: severityBlock,
						Text: fmt.Sprintf("fenced excerpt runs %d lines; the cap is ~%d -- trim states or split", count, excerptLineCap),
					})
				}
			}
			continue
		}
		if inFence {
			count++
		}
	}
	return findings
}

var retestBoxRe = regexp.MustCompile(`^>\s*\*\*The retest\.\*\*`)

// checkRetestBox flags V-R1 when no blockquote opens with the bold
// retest label.
func checkRetestBox(file string, lines []string) []finding {
	for _, line := range lines {
		if retestBoxRe.MatchString(line) {
			return nil
		}
	}
	return []finding{{
		Rule: "V-R1", File: file, Severity: severityBlock,
		Text: "no retest box: expected a blockquote opening with " + retestBoxLabel,
	}}
}

var verdictRe = regexp.MustCompile(`(?i)\bholds with residue\b|\baged out\b|\bholds\b`)

// checkVerdict flags A-3 when none of the taxonomy's verdicts (holds,
// holds with residue, aged out) is stated.
func checkVerdict(file, text string) []finding {
	if verdictRe.MatchString(text) {
		return nil
	}
	return []finding{{
		Rule: "A-3", File: file, Severity: severityBlock,
		Text: "no verdict stated: the retest must end in exactly one of holds, holds with residue, or aged out",
	}}
}

var createHeadingRe = regexp.MustCompile(`(?i)^#{2,}\s.*create`)
var createProseRe = regexp.MustCompile(`(?i)\bto make it\b`)

// checkCreatePointer flags V-S6 when neither a create-section heading
// nor the "to make it do <extension>" formula appears.
func checkCreatePointer(file string, lines []string) []finding {
	for _, i := range proseLines(lines) {
		if createHeadingRe.MatchString(lines[i]) || createProseRe.MatchString(lines[i]) {
			return nil
		}
	}
	return []finding{{
		Rule: "V-S6", File: file, Severity: severityBlock,
		Text: `no create pointer: expected "to make it do <extension>, you add <pieces>" with named files and constructs`,
	}}
}

// checkTakeaway flags V-S7 when the closing prose carries no question.
func checkTakeaway(file string, lines []string) []finding {
	prose := proseLines(lines)
	start := len(prose) - takeawayWindow
	if start < 0 {
		start = 0
	}
	for _, i := range prose[start:] {
		if strings.Contains(lines[i], "?") {
			return nil
		}
	}
	return []finding{{
		Rule: "V-S7", File: file, Severity: severityBlock,
		Text: "no takeaway question: the chapter must close on the question the reader now asks of the next paper",
	}}
}

// --- layer 2: LLM critique ---------------------------------------------

// llmFindings is the schema the model fills. Rule ids must come from
// the constitutions; severity is the model's call, and only blocking
// findings gate the exit code.
type llmResponse struct {
	Findings []struct {
		RuleID   string `json:"rule_id"`
		Location string `json:"location"`
		Severity string `json:"severity"`
		Finding  string `json:"finding"`
	} `json:"findings"`
}

var llmSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"findings": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"rule_id":  map[string]any{"type": "string"},
					"location": map[string]any{"type": "string"},
					"severity": map[string]any{"type": "string", "enum": []string{severityBlock, severityAdvise}},
					"finding":  map[string]any{"type": "string"},
				},
				"required":             []string{"rule_id", "location", "severity", "finding"},
				"additionalProperties": false,
			},
		},
	},
	"required":             []string{"findings"},
	"additionalProperties": false,
}

const llmSystem = `You are the critic for the book "Agentic Applications", which rebuilds
agent papers as declarative-agents machines and retests their claims.
You are given the book's constitutions (rule files), one chapter's SRD
(its drafting contract), and the chapter draft with line numbers.

Review the draft against the READ-kind rules -- the judgment calls a
program cannot check: message discipline (V-M1, V-M2), register (V-G1
to V-G4), the McKinsey test (V-B2), replication honesty (A-1, A-6),
the forced-decisions argument (A-7), and whether the draft delivers
its SRD's section_goal, goals, and content beats.

Every finding must cite one rule id from the constitutions (or the SRD
path when the draft diverges from its contract), a location (a line
number or section heading in the draft), a severity, and one or two
sentences of specifics. Severity "blocking" is for violations of the
constitutions or the SRD contract; "advisory" is for weaker prose that
still complies. Do not restate the deterministic CHECK findings
(retest box, banned words, excerpt length, create pointer, takeaway,
verdict presence) -- a separate layer checks those. Report nothing when
the draft complies; an empty findings list is a valid answer.`

// layerTwo sends the chapter, its SRD, and the constitutions to Claude
// and returns the model's findings. Without ANTHROPIC_API_KEY it prints
// a notice and returns nothing: keyless environments keep layer 1.
func layerTwo(root string, ch rosterChapter, chapterText string) ([]finding, error) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		fmt.Println("ANTHROPIC_API_KEY not set; skipping LLM critique (layer 2)")
		return nil, nil
	}

	prompt, err := buildPrompt(root, ch, chapterText)
	if err != nil {
		return nil, err
	}

	client := anthropic.NewClient()
	resp, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     criticModel,
		MaxTokens: criticMaxTokens,
		System: []anthropic.TextBlockParam{{
			Text: llmSystem,
		}},
		OutputConfig: anthropic.OutputConfigParam{
			Format: anthropic.JSONOutputFormatParam{Schema: llmSchema},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("chapter %s: LLM critique: %w", ch.ID, err)
	}

	var out string
	for _, block := range resp.Content {
		if b, ok := block.AsAny().(anthropic.TextBlock); ok {
			out = b.Text
			break
		}
	}
	if out == "" {
		return nil, fmt.Errorf("chapter %s: LLM critique returned no text (stop_reason %s)", ch.ID, resp.StopReason)
	}

	var parsed llmResponse
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		return nil, fmt.Errorf("chapter %s: parse LLM findings: %w", ch.ID, err)
	}

	var findings []finding
	for _, f := range parsed.Findings {
		sev := f.Severity
		if sev != severityBlock {
			sev = severityAdvise
		}
		findings = append(findings, finding{
			Rule: f.RuleID, File: ch.File, Severity: sev,
			Text: fmt.Sprintf("(LLM, at %s) %s", f.Location, f.Finding),
		})
	}
	return findings, nil
}

// buildPrompt assembles constitutions + SRD + numbered chapter text.
func buildPrompt(root string, ch rosterChapter, chapterText string) (string, error) {
	var b strings.Builder
	for _, name := range []string{"voice.yaml", "argument.yaml", "process.yaml", "venue.yaml"} {
		content, err := os.ReadFile(filepath.Join(root, "docs", "constitutions", name))
		if err != nil {
			return "", fmt.Errorf("read constitution %s: %w", name, err)
		}
		fmt.Fprintf(&b, "=== constitution: docs/constitutions/%s ===\n%s\n", name, content)
	}
	if ch.SRD != "" {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(ch.SRD)))
		if err != nil {
			return "", fmt.Errorf("read %s: %w", ch.SRD, err)
		}
		fmt.Fprintf(&b, "=== SRD: %s ===\n%s\n", ch.SRD, content)
	}
	fmt.Fprintf(&b, "=== chapter draft: %s (line-numbered) ===\n", ch.File)
	for i, line := range strings.Split(chapterText, "\n") {
		fmt.Fprintf(&b, "%4d\t%s\n", i+1, line)
	}
	return b.String(), nil
}
