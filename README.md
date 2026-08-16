# Agentic Applications

**Agent papers as declarative machines — use, modify, create.** A book on
agentic application architecture: each chapter takes one published agent
paper, rebuilds its mechanism as a declarative state machine you can
clone and run, and reports what the rebuild found the paper had left
unsaid.

Rebuilding a paper on a different substrate is a **conceptual
replication**, not a reproduction. The verdicts are honest ones —
*holds*, *holds with residue*, *aged out* — and a claim that survives
with named residue is the most useful result.

**Work in progress, written in the open.** Chapters run first as
articles at [Mesh Intelligence](https://meshintelligence.substack.com?utm_source=github&utm_campaign=agentic-applications-book) and consolidate here as they
stabilize. The runtime is [declarative-agents](https://github.com/Nokia-Bell-Labs/declarative-agents), designed by the
author at Nokia Bell Labs and released as open source (Go runtime,
worked orchestrator/generator example, white paper of eleven design
patterns).

## Contents

**[Introduction](01-introduction.md)** — the method, and what a rebuild
proves.

### [Part I — The Substrate](02-part-i-the-substrate.md)

What a declarative machine is, and what a tool contract declares.

| Chapter | State |
|---|---|
| [You Can't Edit the Model](03-you-cant-edit-the-model.md) | drafted |
| [Every Agent Needs an Undo](04-every-agent-needs-an-undo.md) | stub |
| [How to Give an Agent an API](05-give-an-agent-an-api.md) | stub |
| [Your CLI Is Already a Tool Contract](06-your-cli-is-already-a-tool-contract.md) | stub |
| [Tools Are Words, Not Sentences](07-tools-are-words-not-sentences.md) | stub |
| [How to Build an Agent with Rollbacks](08-rollbacks.md) | stub |

### [The Map](09-the-map.md)

| Chapter | Paper |
|---|---|
| [CoALA With a File Path in Every Box](09-the-map.md) | Cognitive Architectures for Language Agents (2309.02427) |

### [Part II — Papers, Used and Modified](10-part-ii-used-and-modified.md)

| Chapter | Paper |
|---|---|
| [The Paper That Says What My Whole Repo Says](11-stateflow.md) | StateFlow (2403.11322) |
| [ReAct Is Fifteen Lines. The Other Forty-Five Are the Job.](12-react.md) | ReAct (2210.03629) |
| [Kambhampati Was Right, and My Compiler Agrees](13-llm-modulo.md) | LLM-Modulo (2402.01817) |
| [Retrieval That Admits It Found Nothing](14-crag.md) | CRAG (2401.15884), Adaptive-RAG (2403.14403) |
| [The Cheap Model Answers Most Questions](15-frugalgpt.md) | FrugalGPT (2305.05176), RouteLLM (2406.18665) |
| [A Judge You Can Cross-Examine](16-llm-as-judge.md) | LLM-as-a-Judge (2306.05685), CRITIC (2305.11738) |
| [The Best Agent-Safety Paper Is From 1987](17-sagas.md) | Sagas (SIGMOD 1987), SagaLLM (2503.11951) |

### [Part III — Papers, Created](18-part-iii-created.md)

Each of these lands in the runtime first, then gets documented. Listed
as commitments, not completed material.

| Chapter | Paper |
|---|---|
| [Sample Five Times, Keep the Agreement](19-self-consistency.md) | Self-Consistency (2203.11171) |
| [The Retry That Remembers Why It Failed](20-reflexion.md) | Reflexion (2303.11366) |
| [Three Models Walk into a Disagreement](21-debate.md) | Multiagent Debate (2305.14325) |
| [The Agent That Files Its Own Playbook](22-workflow-memory.md) | Agent Workflow Memory (2409.07429), Voyager (2305.16291) |
| [Ask Every Model, Then Ask One More](23-mixture-of-agents.md) | Mixture-of-Agents (2406.04692) |
| The Context Window Is an Environment (chapter not yet drafted; [example runs](examples/applications/large-context-swarm/)) | Recursive Language Models (2512.24601) |

### The Closer

| Chapter | Paper |
|---|---|
| [The Paper My State Machine Refused](24-tree-of-thoughts.md) | Tree of Thoughts (2305.10601) |

## Building the PDF

Requires [mage](https://magefile.org/), pandoc, xelatex, and plantuml.
Renders through the vendored
[Eisvogel](https://github.com/Wandmalfarbe/pandoc-latex-template)
template.

```bash
mage all      # figures + PDF into generated-files/
mage clean    # remove generated artifacts
```

The book's runnable artifacts live under [examples/](examples/): one
declarative application per chapter rebuild, plus catalog profiles
copied from declarative-agents at a pinned release. Listings in drafted
chapters are extracted regions of that source, and `mage audit` checks
the manifest, the provenance pins, and every listing byte-for-byte
against the code it extracts from. The
[examples README](examples/README.md) states the full contract.

## Author

Petar Djukic — Principal AI Architect, 20+ years of production systems,
69 US patents, PhD in Computer Engineering. The companion volume,
[Agentic Coding](https://github.com/petar-djukic/agentic-coding-book),
covers building software with coding agents.

## License

MIT for the build machinery. Book text © 2026 Petar Djukic; quotation
with attribution welcome.
