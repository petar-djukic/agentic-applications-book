// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"strings"
	"testing"
)

var testBanned = []string{"critical", "key", "at the heart of", "game changer"}

func rules(fs []finding) []string {
	var out []string
	for _, f := range fs {
		out = append(out, f.Rule)
	}
	return out
}

func TestBannedWordInProse(t *testing.T) {
	lines := []string{"This is a critical decision.", "Plain sentence."}
	fs := checkBannedWords("ch.md", lines, testBanned)
	if len(fs) != 1 || fs[0].Rule != "V-B1" || fs[0].Line != 1 {
		t.Fatalf("want one V-B1 at line 1, got %v", fs)
	}
}

func TestBannedPhraseMatches(t *testing.T) {
	lines := []string{"It sits at the heart of the design."}
	fs := checkBannedWords("ch.md", lines, testBanned)
	if len(fs) != 1 {
		t.Fatalf("want one finding for the banned phrase, got %v", fs)
	}
}

func TestBannedWordInsideFenceIgnored(t *testing.T) {
	lines := []string{"```yaml", "critical: true", "```", "Clean prose."}
	if fs := checkBannedWords("ch.md", lines, testBanned); len(fs) != 0 {
		t.Fatalf("fenced content must not be scanned, got %v", fs)
	}
}

func TestBannedWordQuotedIsTermOfArt(t *testing.T) {
	lines := []string{`The paper calls this a "critical path" throughout.`}
	if fs := checkBannedWords("ch.md", lines, testBanned); len(fs) != 0 {
		t.Fatalf("quoted terms of art keep their quotes and pass, got %v", fs)
	}
}

func TestBannedWordBoundary(t *testing.T) {
	lines := []string{"The keyboard and the monkey."} // "key" inside a word
	if fs := checkBannedWords("ch.md", lines, testBanned); len(fs) != 0 {
		t.Fatalf("substring inside a longer word must not match, got %v", fs)
	}
}

func TestExcerptOverCap(t *testing.T) {
	var b strings.Builder
	b.WriteString("prose\n```yaml\n")
	for i := 0; i < excerptLineCap+5; i++ {
		b.WriteString("- item\n")
	}
	b.WriteString("```\n")
	fs := checkExcerptLength("ch.md", strings.Split(b.String(), "\n"))
	if len(fs) != 1 || fs[0].Rule != "V-S3" {
		t.Fatalf("want one V-S3, got %v", fs)
	}
}

func TestExcerptUnderCapPasses(t *testing.T) {
	lines := []string{"```", "a", "b", "```"}
	if fs := checkExcerptLength("ch.md", lines); len(fs) != 0 {
		t.Fatalf("short excerpt must pass, got %v", fs)
	}
}

func TestRetestBoxPresent(t *testing.T) {
	lines := []string{"intro", "> **The retest.** The claim, restated. Two years later, this chapter runs it."}
	if fs := checkRetestBox("ch.md", lines); len(fs) != 0 {
		t.Fatalf("retest box present must pass, got %v", fs)
	}
}

func TestRetestBoxMissing(t *testing.T) {
	lines := []string{"intro", "> **A different box.** Not the retest."}
	fs := checkRetestBox("ch.md", lines)
	if len(fs) != 1 || fs[0].Rule != "V-R1" {
		t.Fatalf("want one V-R1, got %v", fs)
	}
}

func TestVerdictPresent(t *testing.T) {
	for _, text := range []string{
		"The claim holds with residue: the paper left out budgets.",
		"The mechanism aged out.",
		"The claim holds.",
	} {
		if fs := checkVerdict("ch.md", text); len(fs) != 0 {
			t.Fatalf("verdict %q must pass, got %v", text, fs)
		}
	}
}

func TestVerdictMissing(t *testing.T) {
	fs := checkVerdict("ch.md", "The chapter ends without judgment.")
	if len(fs) != 1 || fs[0].Rule != "A-3" {
		t.Fatalf("want one A-3, got %v", fs)
	}
}

func TestCreatePointerByFormula(t *testing.T) {
	lines := []string{"To make it do routing, you add one machine and one word."}
	if fs := checkCreatePointer("ch.md", lines); len(fs) != 0 {
		t.Fatalf("create-pointer formula must pass, got %v", fs)
	}
}

func TestCreatePointerByHeading(t *testing.T) {
	lines := []string{"## The create pointer", "Two sentences."}
	if fs := checkCreatePointer("ch.md", lines); len(fs) != 0 {
		t.Fatalf("create heading must pass, got %v", fs)
	}
}

func TestCreatePointerMissing(t *testing.T) {
	lines := []string{"## Summary", "No pointer here."}
	fs := checkCreatePointer("ch.md", lines)
	if len(fs) != 1 || fs[0].Rule != "V-S6" {
		t.Fatalf("want one V-S6, got %v", fs)
	}
}

func TestTakeawayPresent(t *testing.T) {
	lines := []string{"body", "What does the next paper leave for the machine to decide?"}
	if fs := checkTakeaway("ch.md", lines); len(fs) != 0 {
		t.Fatalf("closing question must pass, got %v", fs)
	}
}

func TestTakeawayMissing(t *testing.T) {
	var lines []string
	for i := 0; i < takeawayWindow+10; i++ {
		lines = append(lines, "A plain declarative sentence.")
	}
	fs := checkTakeaway("ch.md", lines)
	if len(fs) != 1 || fs[0].Rule != "V-S7" {
		t.Fatalf("want one V-S7, got %v", fs)
	}
}

func TestTakeawayEarlyQuestionDoesNotCount(t *testing.T) {
	lines := []string{"Is this a question?"} // then a long tail of prose
	for i := 0; i < takeawayWindow+10; i++ {
		lines = append(lines, "Tail prose without questions.")
	}
	fs := checkTakeaway("ch.md", lines)
	if len(fs) != 1 {
		t.Fatalf("a question far from the close must not satisfy V-S7, got %v", fs)
	}
}

func TestSubstrateChapterSkipsPaperForm(t *testing.T) {
	text := "# Title\n\nProse without a retest box, verdict, pointer, or question.\n"
	fs := layerOne("ch.md", text, testBanned, false)
	for _, f := range fs {
		switch f.Rule {
		case "V-R1", "A-3", "V-S6", "V-S7":
			t.Fatalf("substrate chapter must skip paper-form rule %s", f.Rule)
		}
	}
}

func TestPaperChapterRunsAllChecks(t *testing.T) {
	text := "# Title\n\nProse without any of the required form.\n"
	fs := layerOne("ch.md", text, testBanned, true)
	got := map[string]bool{}
	for _, r := range rules(fs) {
		got[r] = true
	}
	for _, want := range []string{"V-R1", "A-3", "V-S6", "V-S7"} {
		if !got[want] {
			t.Fatalf("paper chapter missing expected finding %s in %v", want, rules(fs))
		}
	}
}
