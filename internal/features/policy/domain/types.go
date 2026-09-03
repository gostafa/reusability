// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"github.com/gostafa/reusability/reusability"
)

type (
	// PathMatcher matches import paths against a policy pattern.
	PathMatcher interface {
		Matches(importPath string) bool
	}

	// RuleGate exposes the minimum reusability for a matching rule.
	RuleGate interface {
		MinReusability() float64
	}

	// Violation records one type that fell below a matching rule's minimum.
	Violation = struct {
		Package   string
		Type      string
		Rule      string
		Value     float64
		Threshold float64
	}

	// Rule is a package-path glob paired with a minimum reusability. When more
	// than one rule matches, the most specific pattern wins; exact ties use the
	// later rule.
	Rule struct {
		// Pattern is a glob against the full import path; * matches one segment,
		// ** matches zero or more segments, using / as the separator.
		Pattern string
		// Min is the minimum acceptable type-level reusability in [0, 1].
		Min float64
	}

	matchPos = struct {
		pi, si int
	}

	typeGate = struct {
		typ       *reusability.TypeReport
		pkgPath   string
		pattern   string
		threshold float64
	}
)
