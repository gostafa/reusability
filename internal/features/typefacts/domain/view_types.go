// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"fmt"
)

type (
	// FactView is a debug stringer for fact aggregates.
	FactView interface {
		fmt.Stringer
	}

	// ProjectView is a debug view of project facts.
	ProjectView = FactView

	// TypeView is a debug view of type facts.
	TypeView = FactView

	// MethodView is a debug view of method facts.
	MethodView = FactView
)
