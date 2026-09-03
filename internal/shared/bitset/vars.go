// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package bitset

var (
	_ Setter      = FieldSet{}
	_ Intersecter = SmallFieldSet{}
	_ cloner      = FieldSet{}
	_ container   = FieldSet{}
	_ counter     = FieldSet{}
)
