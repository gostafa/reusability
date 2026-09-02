package domain

import "fmt"

// The view interfaces below expose the fact types as read-only, self-describing
// contracts, so a consumer can depend on a narrow interface instead of the
// concrete fact struct.

// FactView is the shared contract of every fact type: a human-readable debug
// summary of the fact.
type FactView interface {
	fmt.Stringer
}

// ProjectView is a read-only view of a whole project's facts.
type ProjectView interface {
	FactView
}

// TypeView is a read-only view of a single named type's facts.
type TypeView interface {
	FactView
}

// MethodView is a read-only view of a single method's facts.
type MethodView interface {
	FactView
}

// Compile-time proof that the concrete fact types satisfy the view contracts.
var (
	_ ProjectView = (*ProjectFacts)(nil)
	_ TypeView    = (*TypeFacts)(nil)
	_ MethodView  = (*MethodFacts)(nil)
	_ FactView    = (*TypeExtract)(nil)
)
