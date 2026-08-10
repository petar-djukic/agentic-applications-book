# Agentic Applications

**Building applications from agents — declared, verified, and
reversible.** A book on agentic application architecture, LLM tool use,
and declarative agents: agents as configuration (states, transitions,
signals, budgets in YAML), tool contracts with typed side effects and
undo, and rollbacks for loops that do not converge.

**Work in progress, written in the open.** Chapters run first as articles
at [Mesh Intelligence](https://meshintelligence.substack.com?utm_source=github&utm_campaign=agentic-applications-book) and consolidate here as they stabilize. The
framework behind the approach —
[Declarative Agents](https://github.com/Nokia-Bell-Labs/declarative-agents) —
was designed by the author at Nokia Bell Labs and released as open source
(Go runtime, worked orchestrator/generator example, and a white paper of
eleven design patterns).

## Contents

| Chapter | State |
|---|---|
| [Introduction](01-introduction.md) | drafted |
| [You Can't Edit the Model](02-you-cant-edit-the-model.md) | stub — article queued |
| [Every Agent Needs an Undo](03-every-agent-needs-an-undo.md) | stub — article queued |
| [How to Give an Agent an API](04-give-an-agent-an-api.md) | stub — article queued |
| [Your CLI Is Already a Tool Contract](05-your-cli-is-already-a-tool-contract.md) | stub — article queued |
| [Tools Are Words, Not Sentences](06-tools-are-words-not-sentences.md) | stub — article queued |
| [How to Build an Agent with Rollbacks](07-rollbacks.md) | stub — article queued |

## Building the PDF

Requires [mage](https://magefile.org/), pandoc, xelatex, and plantuml.
Renders through the vendored
[Eisvogel](https://github.com/Wandmalfarbe/pandoc-latex-template)
template.

```bash
mage all      # figures + PDF into generated-files/
mage clean    # remove generated artifacts
```

## Author

Petar Djukic — Principal AI Architect, 20+ years of production systems,
69 US patents, PhD in Computer Engineering. The companion volume,
[Agentic Coding](https://github.com/petar-djukic/agentic-coding-book),
covers building software with coding agents.

## License

MIT for the build machinery. Book text © 2026 Petar Djukic; quotation
with attribution welcome.
