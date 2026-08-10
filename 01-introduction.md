# Introduction

An agentic application is not a chatbot with tools bolted on. It is an
application whose control flow an agent decides at runtime — which makes
the engineering question the opposite of the usual one. You cannot edit
the model, so everything you need to govern lives outside it: the tools
it may call, the states it may occupy, the budgets it may spend, and the
path back when it does something wrong.

This book approaches the problem declaratively. An agent is
configuration: tools, states, transitions, signals, and budgets declared
up front and checked before anything runs. The LLM is one declared
component the loop calls — not the center of the system. The approach
comes from [Declarative Agents](https://github.com/Nokia-Bell-Labs/declarative-agents),
a framework I designed at Nokia Bell Labs, released as open source with a
Go runtime and a white paper of eleven design patterns.

Each chapter runs as an article first at
[Mesh Intelligence](https://meshintelligence.substack.com?utm_source=github&utm_campaign=agentic-applications-book); chapters consolidate here as the live
versions stabilize. The book is written in the open — stubs and seams
included. A companion volume,
[Agentic Coding](https://github.com/petar-djukic/agentic-coding-book),
covers using agents to build software; this one covers building
applications out of agents.
