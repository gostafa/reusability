// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

// Package embedding exercises embedded and promoted members.
package embedding

// Base counts things.
type Base struct {
	Count int
}

// Wrapper embeds Base; promoted members stay Base's.
type Wrapper struct {
	Base
	Name string
}
