// Copyright (c) 2026 Petar Djukic
// SPDX-License-Identifier: MIT

package main

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// The assertions are what the demo's verdict rests on, so they are
// tested against synthetic request logs rather than trusted. Each case
// builds the log a correct run would produce, then breaks exactly one
// property.

const addPath = "/api/v2/tenants/default_tenant/databases/default_database/collections/rlm-demo-collection/add"
const queryPath = "/api/v2/tenants/default_tenant/databases/default_database/collections/rlm-demo-collection/query"
const collectionsPath = "/api/v2/tenants/default_tenant/databases/default_database/collections"
const deletePath = "/api/v2/tenants/default_tenant/databases/default_database/collections/rlm-demo-collection/delete"

func goodQuery() record {
	return record{Method: "POST", Path: queryPath,
		Body: []byte(`{"query_embeddings":[[0.1]],"where":{"source":"corpus"},"where_document":{"$contains":"RLM-DEMO-1"},"n_results":3}`)}
}

func goodAdd(document string) record {
	return record{Method: "POST", Path: addPath,
		Body: []byte(`{"documents":"` + document + `","metadatas":[{"source":"derived","agent":"rlm-worker","round":1}]}`)}
}

// goodLog is the log a correct run produces: three worker queries plus
// the root's collect, four writes, create-or-gets, and a teardown.
func goodLog(finalAnswer string) []record {
	log := []record{{Method: "POST", Path: collectionsPath, Body: []byte(`{"name":"rlm-demo-collection"}`)}}
	for i := 0; i < 4; i++ {
		log = append(log, goodQuery())
	}
	log = append(log,
		workerChat("reading the observations cluster"),
		workerChat("reading the deployments cluster"),
		workerChat("reading the calibration ledger"),
		goodAdd("Station Kelvin recorded a salinity anomaly. [RLM-DEMO-1]"),
		goodAdd("CTD-114 was the only conductivity sensor. [RLM-DEMO-1]"),
		goodAdd("Renata Oyelaran calibrated it. [RLM-DEMO-1]"),
		goodAdd(finalAnswer),
		record{Method: "POST", Path: deletePath, Body: []byte(`{}`)},
	)
	return log
}

func testFixtures(t *testing.T) (expectations, corpus) {
	t.Helper()
	expected, err := loadExpectations("../../testdata/expected.yaml")
	if err != nil {
		t.Fatal(err)
	}
	docs, err := loadCorpus("../../testdata/corpus/documents.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return expected, docs
}

func fullAnswer(expected expectations) string {
	return strings.Join(expected.MultiHop.AnswerContains, " ")
}

func TestAllAssertionsPassOnAGoodRun(t *testing.T) {
	expected, docs := testFixtures(t)
	log := goodLog(fullAnswer(expected))
	if err := assertAll(log, expected, docs); err != nil {
		t.Fatalf("a correct run must pass every assertion: %v", err)
	}
}

func TestA1FailsWhenTheCollectionLeaks(t *testing.T) {
	expected, _ := testFixtures(t)
	log := goodLog(fullAnswer(expected))
	var kept []record
	for _, r := range log {
		if r.Path != deletePath {
			kept = append(kept, r)
		}
	}
	if r := assertCollectionLifecycle(kept, expected); r.passed {
		t.Fatal("a run that never deletes its collection must fail A1")
	}
}

func TestA2FailsBelowTwoWorkers(t *testing.T) {
	// Two queries means one worker plus the root's collect.
	log := []record{goodQuery(), goodQuery()}
	if r := assertFanOut(log); r.passed {
		t.Fatal("a single worker must fail A2; C2 requires at least two intents")
	}
}

func TestA3FailsOnAnUnfilteredQuery(t *testing.T) {
	log := []record{goodQuery(), goodQuery(), goodQuery(),
		{Method: "POST", Path: queryPath, Body: []byte(`{"query_embeddings":[[0.1]],"n_results":3}`)}}
	if r := assertQueryFilters(log); r.passed {
		t.Fatal("a query without where and where_document must fail A3")
	}
}

func TestA4FailsOnAnUntaggedWrite(t *testing.T) {
	log := []record{goodAdd("tagged"),
		{Method: "POST", Path: addPath, Body: []byte(`{"documents":"untagged"}`)}}
	if r := assertProvenanceTags(log); r.passed {
		t.Fatal("an untagged write must fail A4")
	}
}

// The assertion the rebuild rests on. A run that leaks one corpus
// sentence into any request has stopped being the mechanism the chapter
// describes, even though every other assertion still passes.
func workerChat(body string) record {
	return record{Method: "POST", Path: "/api/chat",
		Body: []byte(`{"messages":[{"role":"system","content":"You are ` + workerChatMarker + `"},{"content":"` + body + `"}]}`)}
}

func TestA5FailsWhenCorpusTextLeaks(t *testing.T) {
	expected, docs := testFixtures(t)
	leaked := distinctiveSentence(docs.Documents[0].Text)
	if leaked == "" {
		t.Fatal("the first corpus document has no sentence long enough to probe with")
	}
	log := append(goodLog(fullAnswer(expected)),
		workerChat("reading passages"), workerChat("reading more"),
		record{Method: "POST", Path: "/api/chat", Body: []byte(`{"messages":[{"content":"` + leaked + `"}]}`)})

	if r := assertNoCorpusTextAtRoot(log, docs); r.passed {
		t.Fatal("corpus text in a root-side request must fail A5")
	}
	// And the point of having it: everything else still passes, so A5
	// is the only thing standing between this run and a false verdict.
	if r := assertFinalAnswer(log, expected); !r.passed {
		t.Fatal("A6 should still pass on the leaking run, which is why A5 matters")
	}
}

// The worker is allowed to hold corpus text; its marked model calls are
// exempt from the scan. The same passage in a worker chat and a clean
// root side must pass.
func TestA5ExemptsWorkerReads(t *testing.T) {
	expected, docs := testFixtures(t)
	passage := distinctiveSentence(docs.Documents[0].Text)
	log := append(goodLog(fullAnswer(expected)),
		workerChat(passage), workerChat(passage))
	if r := assertNoCorpusTextAtRoot(log, docs); !r.passed {
		t.Fatalf("a worker reading passages must not fail A5: %s", r.detail)
	}
}

// The exemption is guarded: with no worker-marked calls the partition
// is stale and the scan proves nothing, so A5 fails rather than
// passing vacuously.
func TestA5FailsWithoutWorkerMarkedCalls(t *testing.T) {
	expected, docs := testFixtures(t)
	var log []record
	for _, r := range goodLog(fullAnswer(expected)) {
		if !strings.Contains(string(r.Body), workerChatMarker) {
			log = append(log, r)
		}
	}
	if r := assertNoCorpusTextAtRoot(log, docs); r.passed {
		t.Fatal("a log with no worker-marked model calls must fail A5")
	}
}

func TestA6FailsOnAnIncompleteAnswer(t *testing.T) {
	expected, _ := testFixtures(t)
	log := goodLog("Renata Oyelaran did something")
	if r := assertFinalAnswer(log, expected); r.passed {
		t.Fatal("a Final entry missing answer fragments must fail A6")
	}
}

// The fixtures the demo replays must cover every route the machines
// call, or the run dies on a response nobody wrote.
func TestFixturesParseAndDeclareResponses(t *testing.T) {
	for _, path := range []string{"../../testdata/responses/chroma.yaml", "../../testdata/responses/ollama.yaml"} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var f fixture
		if err := yaml.Unmarshal(content, &f); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if len(f.Routes) == 0 {
			t.Fatalf("%s declares no routes", path)
		}
		for _, r := range f.Routes {
			if len(r.Responses) == 0 {
				t.Fatalf("%s: route %s %s declares no responses", path, r.Method, r.Path)
			}
		}
	}
}
