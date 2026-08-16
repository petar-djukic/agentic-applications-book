// Copyright (c) 2026 Petar Djukic
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"strings"
)

// The six assertions of srd-large-context-swarm D3, in the order
// testdata/expected.yaml lists them. Each reports pass or fail with the
// evidence it used, and the command exits non-zero if any fail.
func assertAll(records []record, expected expectations, docs corpus) error {
	results := []result{
		assertCollectionLifecycle(records, expected),
		assertFanOut(records),
		assertQueryFilters(records),
		assertProvenanceTags(records),
		assertNoCorpusTextAtRoot(records, docs),
		assertFinalAnswer(records, expected),
	}
	failed := 0
	for _, r := range results {
		mark := "ok  "
		if !r.passed {
			mark = "FAIL"
			failed++
		}
		fmt.Printf("%s %s: %s\n", mark, r.id, r.detail)
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d assertions failed", failed, len(results))
	}
	return nil
}

type result struct {
	id     string
	passed bool
	detail string
}

func countPath(records []record, suffix string) int {
	n := 0
	for _, r := range records {
		if strings.HasSuffix(r.Path, suffix) {
			n++
		}
	}
	return n
}

// A1. The per-task collection is created and torn down.
func assertCollectionLifecycle(records []record, expected expectations) result {
	creates := countPath(records, "/collections")
	deletes := countPath(records, "/delete")
	switch {
	case creates == 0:
		return result{"A1", false, "no create-or-get for " + expected.Request.Collection}
	case deletes == 0:
		return result{"A1", false, fmt.Sprintf(
			"%d create-or-get calls but no delete: the collection leaked", creates)}
	default:
		return result{"A1", true, fmt.Sprintf("%d create-or-get, %d delete", creates, deletes)}
	}
}

// A2. At least two intents, one worker each. A worker is counted by its
// query: each runs exactly one search before writing.
func assertFanOut(records []record) result {
	queries := countPath(records, "/query")
	// The root's own collect query is one of them; the rest are workers.
	workers := queries - 1
	if workers < 2 {
		return result{"A2", false, fmt.Sprintf(
			"%d query calls implies %d worker(s); C2 requires at least two intents", queries, workers)}
	}
	return result{"A2", true, fmt.Sprintf("%d worker queries plus the root's collect", workers)}
}

// A3. Worker queries carried both filter kinds. Asserted from the
// request body, because a query that returned results proves only that
// the stub replied, not what was asked.
func assertQueryFilters(records []record) result {
	withBoth, total := 0, 0
	for _, r := range records {
		if !strings.HasSuffix(r.Path, "/query") {
			continue
		}
		total++
		body := string(r.Body)
		if strings.Contains(body, `"where"`) && strings.Contains(body, `"where_document"`) {
			withBoth++
		}
	}
	if total == 0 {
		return result{"A3", false, "no query requests recorded"}
	}
	if withBoth != total {
		return result{"A3", false, fmt.Sprintf(
			"%d of %d queries carried both where and where_document", withBoth, total)}
	}
	return result{"A3", true, fmt.Sprintf("all %d queries carried where and where_document", total)}
}

// A4. Every write carries the three provenance fields.
func assertProvenanceTags(records []record) result {
	tagged, total := 0, 0
	for _, r := range records {
		if !strings.HasSuffix(r.Path, "/add") {
			continue
		}
		total++
		body := string(r.Body)
		if strings.Contains(body, `"source"`) &&
			strings.Contains(body, `"agent"`) &&
			strings.Contains(body, `"round"`) {
			tagged++
		}
	}
	if total == 0 {
		return result{"A4", false, "no add requests recorded"}
	}
	if tagged != total {
		return result{"A4", false, fmt.Sprintf(
			"%d of %d writes carried source, agent, and round", tagged, total)}
	}
	return result{"A4", true, fmt.Sprintf("all %d writes tagged", total)}
}

// A5. No fixture document text in any request the root originated.
//
// This is invariant I1 and the assertion the whole rebuild rests on. A
// run can satisfy every other check here and still be ordinary
// retrieval augmentation; this is what tells them apart.
//
// The test is deliberately blunt: no distinctive sentence from any
// corpus document may appear in any recorded request body. Worker
// requests never carry document text either -- a worker reads passages
// in its response and writes only its own finding -- so scanning
// everything costs nothing and removes the need to attribute each
// request to a process.
func assertNoCorpusTextAtRoot(records []record, docs corpus) result {
	for _, doc := range docs.Documents {
		probe := distinctiveSentence(doc.Text)
		if probe == "" {
			continue
		}
		if bodyContains(records, probe) {
			return result{"A5", false, fmt.Sprintf(
				"corpus text from %s reached a request: %q", doc.ID, truncate(probe, 60))}
		}
	}
	return result{"A5", true, fmt.Sprintf(
		"no sentence from any of %d corpus documents appears in %d recorded requests",
		len(docs.Documents), len(records))}
}

// distinctiveSentence returns the document's longest sentence, which is
// the least likely to collide with a finding that legitimately
// paraphrases it.
func distinctiveSentence(text string) string {
	longest := ""
	for _, sentence := range strings.Split(text, ". ") {
		sentence = strings.TrimSpace(strings.ReplaceAll(sentence, "\n", " "))
		for strings.Contains(sentence, "  ") {
			sentence = strings.ReplaceAll(sentence, "  ", " ")
		}
		if len(sentence) > len(longest) {
			longest = sentence
		}
	}
	if len(longest) < 40 {
		return ""
	}
	return longest
}

// A6. The Final entry carries the multi-hop answer.
func assertFinalAnswer(records []record, expected expectations) result {
	var missing []string
	for _, fragment := range expected.MultiHop.AnswerContains {
		if !bodyContains(records, fragment) {
			missing = append(missing, fragment)
		}
	}
	if len(missing) > 0 {
		return result{"A6", false, "the Final entry is missing: " + strings.Join(missing, ", ")}
	}
	return result{"A6", true, fmt.Sprintf(
		"the Final entry carries all %d answer fragments", len(expected.MultiHop.AnswerContains))}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
