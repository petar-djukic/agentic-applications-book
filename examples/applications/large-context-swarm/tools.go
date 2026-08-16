// Copyright (c) 2026 Petar Djukic
// SPDX-License-Identifier: MIT

//go:build tools

// Package tools pins the runtime binary the demo runs. The blank import
// is what keeps the agent-core requirement in go.mod: nothing here
// imports the runtime as a library, and without this file `go mod tidy`
// drops the requirement and the audit's single-release check has
// nothing to check.
package tools

import _ "github.com/Nokia-Bell-Labs/declarative-agents/agent-core/cmd/agent"
