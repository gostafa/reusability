// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

var (
	_ ProjectView = (*ProjectFacts)(nil)
	_ TypeView    = (*TypeFacts)(nil)
	_ MethodView  = (*MethodFacts)(nil)
	_ FactView    = (*TypeExtract)(nil)
)
