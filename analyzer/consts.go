// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package analyzer

const (
	// Name is the analyzer and golangci-lint plugin name.
	Name = "reusability"

	// Doc is the analyzer documentation shown by go/analysis drivers.
	Doc = `enforce Go type-reusability policy thresholds

Reports policy violations when a type's reusability index falls below a
configured minimum. Rules match package import paths using glob patterns
(* = one segment, ** = zero or more). Policy rules are configured inline
in the golangci-lint settings block.`

	zero                  = 0
	one                   = 1
	two                   = 2
	floatBits             = 64
	defaultPackagePattern = "./..."

	errFmtUnmarshal = "UnmarshalSettings: %w"
	errFmtValidate  = "validate: %w"
	errFmtRemap     = "remapKebabKeys: %w"
	errFmtQuoted    = "%w: %q"
)
