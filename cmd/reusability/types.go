// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

// Package main provides reusable analysis utilities.
package main

type (
	mainRuntime struct {
		run  func([]string) int
		exit func(int)
	}
)
