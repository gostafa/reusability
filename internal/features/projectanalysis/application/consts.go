// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

const (
	needComplexity metricNeeds = 1 << 0
	needCohesion   metricNeeds = 1 << 1

	zero = 0

	opAnalyze  = "Analyze"
	opAssemble = "assemble"
	opNewCalc  = "newReusabilityCalculator"

	errFmtOp = "%s: %w"
)
