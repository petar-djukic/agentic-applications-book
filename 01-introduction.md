# Introduction

An agent paper is a mechanism plus a benchmark. Strip the benchmark and
the mechanism usually fits in a page of YAML you can read, run, and
change — and the parts the paper never specified turn out to be exactly
the parts a running machine forces you to write down.

That is the whole method of this book. Each chapter takes one published
paper, rebuilds its mechanism as a declarative machine in
[declarative-agents](https://github.com/Nokia-Bell-Labs/declarative-agents), and reports what survived the move. The
pedagogy is Use–Modify–Create: run the shipped machine, change one
declared thing and watch the behavior move, then take the closing
pointer and build the next thing yourself.

## What a rebuild proves, and what it does not

Rebuilding a paper on a different substrate is a **conceptual
replication**, not a reproduction. No chapter claims to have rerun a
paper's benchmarks unless it did. What the rebuild tests is whether the
claim still makes sense once every decision the paper left implicit has
to be declared: the signal alphabet, the budgets, the terminal states,
and who owns the exit.

Verdicts accumulate across the book, and three kinds are honest:

| Verdict | Meaning |
|---|---|
| Holds | The mechanism works as described on this substrate |
| Holds with residue | It works, and the rebuild names what the paper left out |
| Aged out | The claim no longer earns its complexity |

A claim surviving with named residue is the most useful outcome, not a
disappointing one.

## How the book is arranged

**Part I — The substrate.** What a declarative machine is: states and
transitions in a file, tools with typed inputs and declared side effects,
undo as part of the tool contract. Six short chapters. Readers who know
the runtime can skip to the map.

**The map.** CoALA's taxonomy of memory, actions, and decision
procedures, with a file path in every box — the vocabulary the rest of
the book uses.

**Part II — Papers, used and modified.** Seven mechanisms already
running: state-driven control flow, ReAct, LLM-Modulo, corrective
retrieval, model routing, LLM-as-judge, and a 1987 database paper that
turns out to be the best agent-safety paper in the set.

**Part III — Papers, created.** Five mechanisms landed in the repo
first, then documented: self-consistency, Reflexion, multi-agent debate,
workflow memory, and mixture-of-agents.

**The closer.** Tree of Thoughts — the paper the machine refuses, and
why that refusal is the sharpest available description of what
declarative machines are for.

Chapters run as articles first at
[Mesh Intelligence](https://meshintelligence.substack.com?utm_source=github&utm_campaign=agentic-applications-book) and consolidate here as they stabilize.
Expect stubs and seams; the book is written in the open.
