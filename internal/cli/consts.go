// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package cli

const (
	defaultWebReportName = "reusability-report.html"
	helpTempPattern      = "reusability-help-*.html"

	emptyString = ""
	commaSep    = ","
	formatFlag  = "format"
	webFlagLong = "--web"

	keyDuration      = "duration"
	keyError         = "error"
	keyPackages      = "packages"
	keyPath          = "path"
	keyPolicySource  = "policy_source"
	keyViolations    = "violations"
	keyWant          = "want"
	wantFormatValues = "text, json, csv, or web"

	policySourceFlagRules = "flag rules"

	msgAnalysisComplete          = "analysis complete"
	msgAnalysisFailed            = "analysis failed"
	msgConflictingWebFormat      = "conflicting flags: -web implies -format=web"
	msgCPUProfilingFailed        = "cpu profiling failed"
	msgInvalidFormat             = "invalid format"
	msgMemoryProfilingFailed     = "memory profiling failed"
	msgMetricsGuideWritten       = "metrics guide written"
	msgOpeningGuideFailed        = "opening the metrics guide in a browser failed"
	msgOpeningReportFailed       = "opening the report in a browser failed"
	msgPolicyCheckFailed         = "policy check failed"
	msgPolicyCheckSucceeded      = "policy check succeeded"
	msgPolicyConfigurationFailed = "policy configuration failed"
	msgReportWritten             = "report written"
	msgWritingGuideFailed        = "writing the metrics guide failed"
	msgWritingReportFailed       = "writing report failed"
	msgWritingViolationsFailed   = "writing policy violations failed"

	usageHeader = "usage: reusability [flags] [patterns...]\n\n"
	usageFooter = "\nFor an illustrated guide to the reported metric:\n  reusability --help " +
		webFlagLong + "\n"

	floatBits = 64

	exitOK          = 0
	exitFail        = 1
	exitUsage       = 2
	exitPolicy      = 3
	exitInterrupted = 130
)
