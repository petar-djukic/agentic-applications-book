// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

// The examples mage module. Run from the examples directory
// (mage -d examples from the book root): Audit validates the manifest
// and its constraints, Test validates every application's declarations,
// Demo runs each application's canned demo.
package main

import (
	"bytes"
	"io"

	"gopkg.in/yaml.v3"
)

// newBytesReader keeps the strict decoder construction in one place.
func newBytesReader(content []byte) io.Reader {
	return bytes.NewReader(content)
}

// yamlUnmarshalLenient parses YAML without strict field checking, for
// documents whose full shape belongs to another owner (example SRDs,
// application declarations).
func yamlUnmarshalLenient(content []byte, out any) error {
	return yaml.Unmarshal(content, out)
}
