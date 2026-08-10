# Part I — The Substrate

Every machine in this book runs on the same runtime, so the paper
chapters can talk about mechanisms instead of plumbing. This part covers
what a declarative machine is and what a tool contract has to declare.

The runtime is [declarative-agents](https://github.com/Nokia-Bell-Labs/declarative-agents) — a Go implementation released
as open source by Nokia Bell Labs, with a worked
orchestrator/generator example and a white paper of eleven design
patterns. An agent is configuration: states, transitions, signals,
budgets, and tools, declared in YAML and checked before anything runs.
The language model is one declared component the loop calls, not the
center of the system.

Readers already fluent in the runtime can go straight to the map.
