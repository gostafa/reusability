// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"github.com/gostafa/reusability/reusability"
)

type (
	// Violation records one type that fell below a matching rule's minimum.
	Violation struct {
		Package   string
		Type      string
		Rule      string
		Value     float64
		Threshold float64
	}

	// Rule is a package-path glob paired with a minimum reusability.
	Rule struct {
		// Pattern is a glob against the full import path; * matches one segment,
		// ** matches zero or more segments, using / as the separator.
		Pattern string
		// Min is the minimum acceptable type-level reusability in [0, 1].
		Min float64
	}

	matchPos struct {
		pi, si int
	}

	typeGate struct {
		typ       *reusability.TypeReport
		pkgPath   string
		pattern   string
		threshold float64
	}

	ruleCandidate struct {
		importPath string
		pattern    string
		rule       Rule
		threshold  float64
		matched    bool
	}
)
