// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	policydomain "github.com/gostafa/reusability/internal/features/policy/domain"
	reporting "github.com/gostafa/reusability/internal/features/reporting/application"
	reportingdomain "github.com/gostafa/reusability/internal/features/reporting/domain"
	"github.com/gostafa/reusability/internal/features/reporting/ports/outbound"
	"github.com/gostafa/reusability/internal/infrastructure/browser"
	"github.com/gostafa/reusability/internal/infrastructure/profiling"
	"github.com/gostafa/reusability/internal/infrastructure/sinks"
	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/internal/shared/version"
	"github.com/gostafa/reusability/reusability"
)

// Run executes reusability with args and returns its process exit code.
func Run(args []string) int { return execute(args) }

func asConfig(set *flag.FlagSet, vals *flagValues, wts *reusability.Weights) reusability.Config {
	return reusability.Config{
		Patterns:           set.Args(),
		IncludeTests:       *vals.includeTests,
		IncludeGenerated:   *vals.generated,
		BuildTags:          splitList(*vals.buildTags),
		Workers:            *vals.workers,
		DependencyScope:    *vals.dependencyScope,
		FieldUsageMode:     *vals.fieldUsage,
		ContinueOnError:    *vals.continueOnError,
		ReusabilityWeights: *wts,
	}
}

func analyzeAndReportWith(cfg *runtimeConfig, analyze analyzeFunc) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()

	return runPipeline(ctx, cfg, analyze)
}

func analyzeThenWrite(ctx context.Context, cfg *runtimeConfig, analyze analyzeFunc) int {
	report, code := runAnalyze(ctx, cfg, analyze)

	if code != exitOK {
		return code
	}

	heapCode := writeHeapIfNeeded(ctx, cfg)

	if heapCode != exitOK {
		return heapCode
	}

	return writeAndGate(ctx, cfg, &report)
}

func applyWebFormat(args *buildArgs) (name string, code int, ok bool) {
	name = *args.vals.format

	if !*args.vals.webReport {
		return name, exitOK, true
	}

	if formatWasSet(args.flagSet) && name != string(reportingdomain.FormatWeb) {
		args.logger.ErrorContext(
			context.Background(),
			msgConflictingWebFormat,
			slog.String(formatFlag, name),
		)

		return emptyString, exitUsage, false
	}

	return string(reportingdomain.FormatWeb), exitOK, true
}

func resolveOutputPath(output, format string) string {
	if output == emptyString && format == string(reportingdomain.FormatWeb) {
		return defaultWebReportName
	}

	return output
}

func buildRuntime(args *buildArgs) parseResult {
	formatOutcome := resolveReportFormat(args)

	if !formatOutcome.ok {
		return doneParse(formatOutcome.code)
	}

	gatingOutcome := resolveGating(args.vals, args.logger)

	if !gatingOutcome.ok {
		return doneParse(gatingOutcome.code)
	}

	return finishRuntime(args, &formatOutcome, &gatingOutcome)
}

func buildWriteRequest(cfg *runtimeConfig, report *reusability.Report) *reporting.WriteRequest {
	return &reporting.WriteRequest{
		Report: *report,
		Format: reportingdomain.Format(cfg.format),
		Sink:   reportSink(cfg.output),
		Options: reportingdomain.TextOptions{
			Color: cfg.output == emptyString && os.Getenv("NO_COLOR") == emptyString &&
				stdoutIsTerminal(),
			Explain: cfg.explain,
		},
	}
}

func closeOSFile(file *os.File) error {
	err := file.Close()
	if err != nil {
		return fmt.Errorf("close file: %w", err)
	}

	return nil
}

func doneParse(code int) parseResult {
	return parseResult{code: code, done: true}
}

func enforcePolicy(ctx context.Context, cfg *runtimeConfig, report *reusability.Report) int {
	if !cfg.gating {
		return exitOK
	}

	violations := policydomain.Evaluate(report, policyRules(cfg))

	if len(violations) == exitOK {
		return policyPassed(ctx, cfg)
	}

	return policyFailed(ctx, cfg, violations)
}

func execute(args []string) int {
	return executeWithAnalyze(args, reusability.Analyze)
}

func executeWithAnalyze(args []string, analyze analyzeFunc) int {
	parsed := parseCLI(args)

	if parsed.done {
		return parsed.code
	}

	return analyzeAndReportWith(&parsed.cfg, analyze)
}

func failAnalyze(ctx context.Context, logger *slog.Logger, err error) int {
	logger.ErrorContext(
		ctx,
		msgAnalysisFailed,
		slog.Any(keyError, err),
	)

	if errors.Is(err, context.Canceled) {
		return exitInterrupted
	}

	return exitFail
}

func finalizeHelpDocs(file *os.File, path string) (string, error) {
	err := closeOSFile(file)
	if err != nil {
		return emptyString, fmt.Errorf("close help temp: %w", err)
	}

	docsPath, err := renderHelpDocs(path)
	if err != nil {
		return emptyString, fmt.Errorf("render help docs: %w", err)
	}

	return docsPath, nil
}

func finishRuntime(args *buildArgs, format *formatResult, gating *gatingResult) parseResult {
	weightsOutcome := resolveWeights(args.vals, args.logger)

	if !weightsOutcome.ok {
		return doneParse(weightsOutcome.code)
	}

	args.format = format
	args.gating = gating

	return parseResult{cfg: runtimeConfiguration(args, &weightsOutcome.weights)}
}

func runtimeConfiguration(args *buildArgs, weights *reusability.Weights) runtimeConfig {
	vals := args.vals
	patterns, mins := ruleSlices(args.gating.rules)

	return runtimeConfig{
		logger:        args.logger,
		format:        args.format.format,
		output:        resolveOutputPath(*vals.output, args.format.format),
		explain:       *vals.explain,
		analysis:      asConfig(args.flagSet, vals, weights),
		gating:        *vals.check,
		rulePatterns:  patterns,
		ruleMins:      mins,
		policySource:  args.gating.source,
		cpuProfile:    *vals.cpuProfile,
		memoryProfile: *vals.memoryProfile,
		webToDefault: *vals.output == emptyString &&
			args.format.format == string(reportingdomain.FormatWeb),
	}
}

func formatWasSet(flagSet *flag.FlagSet) bool {
	set := false

	flagSet.Visit(func(item *flag.Flag) { set = set || item.Name == formatFlag })

	return set
}

func isTruthyWebFlag(arg string) bool {
	if arg == "-web" || arg == webFlagLong {
		return true
	}

	return webEqualsTruthy(arg)
}

func loggerFor(vals *flagValues) *slog.Logger {
	level := slog.LevelInfo

	if *vals.verbose {
		level = slog.LevelDebug
	}

	return newLogger(level)
}

func logWebHelpWritten(logger *slog.Logger, path string) {
	logger.InfoContext(
		context.Background(),
		msgMetricsGuideWritten,
		slog.String(keyPath, path),
	)
	openHelpIfTerminal(logger, path)
}

func maybeStartCPU(ctx context.Context, cfg *runtimeConfig) (stop func() error, code int) {
	if cfg.cpuProfile == emptyString {
		return nil, exitOK
	}

	stopProfile, err := profiling.StartCPU(cfg.cpuProfile)
	if err != nil {
		cfg.logger.ErrorContext(
			ctx,
			msgCPUProfilingFailed,
			slog.Any(keyError, err),
		)

		return nil, exitFail
	}

	return stopProfile, exitOK
}

func newFlagSet() (flagSet *flag.FlagSet, vals *flagValues) {
	flagSet = flag.NewFlagSet("reusability", flag.ContinueOnError)
	flagSet.SetOutput(os.Stderr)

	flagSet.Usage = func() { printUsage(flagSet) }

	vals = &flagValues{}
	registerReportFlags(flagSet, vals)
	registerWeightFlags(flagSet, vals)
	registerAnalysisFlags(flagSet, vals)
	registerMiscFlags(flagSet, vals)

	return flagSet, vals
}

func newLogger(level slog.Leveler) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if len(groups) == exitOK && (attr.Key == slog.TimeKey || attr.Key == slog.LevelKey) {
				return slog.Attr{}
			}

			return attr
		},
	}

	return slog.New(slog.NewTextHandler(os.Stderr, opts))
}

func openHelpIfTerminal(logger *slog.Logger, path string) {
	if !stdoutIsTerminal() {
		return
	}

	err := browser.Open(path)
	if err != nil {
		logger.WarnContext(
			context.Background(),
			msgOpeningGuideFailed,
			slog.Any(keyError, err),
		)
	}
}

func openWebBrowser(ctx context.Context, cfg *runtimeConfig) {
	if !stdoutIsTerminal() {
		return
	}

	err := browser.Open(cfg.output)
	if err != nil {
		cfg.logger.WarnContext(
			ctx,
			msgOpeningReportFailed,
			slog.Any(keyError, err),
		)
	}
}

func openWebIfNeeded(ctx context.Context, cfg *runtimeConfig) {
	if !cfg.webToDefault {
		return
	}

	cfg.logger.InfoContext(
		ctx,
		msgReportWritten,
		slog.String(keyPath, cfg.output),
	)
	openWebBrowser(ctx, cfg)
}

func parseCLI(args []string) parseResult {
	flagSet, vals := newFlagSet()

	err := flagSet.Parse(args)
	if err != nil {
		return parseHelpOrUsage(err, args)
	}

	if *vals.showVersion {
		return parseVersion()
	}

	return buildRuntime(&buildArgs{flagSet: flagSet, vals: vals, logger: loggerFor(vals)})
}

func parseHelpOrUsage(err error, args []string) parseResult {
	if errors.Is(err, flag.ErrHelp) && wantsWebHelp(args) {
		return parseResult{code: runWebHelp(), done: true}
	}

	return parseResult{code: exitUsage, done: true}
}

func parseRuleSpec(value string) (ruleSpec, error) {
	pattern, number, ok := strings.Cut(value, ":")

	if !ok {
		return ruleSpec{}, fmt.Errorf("%w: got %q", errExpectedPatternMin, value)
	}

	pattern = strings.TrimSpace(pattern)

	if pattern == emptyString {
		return ruleSpec{}, fmt.Errorf("%w: %q", errEmptyPattern, value)
	}

	parsed, err := strconv.ParseFloat(strings.TrimSpace(number), floatBits)
	if err != nil {
		return ruleSpec{}, fmt.Errorf("invalid number in %q: %w", value, err)
	}

	return ruleSpec{pattern: pattern, minimum: parsed}, nil
}

func parseVersion() parseResult {
	err := printTo(os.Stdout, "reusability "+version.Version()+"\n")
	if err != nil {
		return parseResult{code: exitFail, done: true}
	}

	return parseResult{code: exitOK, done: true}
}

func policyFailed(ctx context.Context, cfg *runtimeConfig, items []policydomain.Violation) int {
	cfg.logger.ErrorContext(
		ctx, msgPolicyCheckFailed,
		slog.String(keyPolicySource, cfg.policySource),
		slog.Int(keyViolations, len(items)),
	)
	writeViolations(ctx, cfg.logger, items)

	return exitPolicy
}

func policyPassed(ctx context.Context, cfg *runtimeConfig) int {
	cfg.logger.InfoContext(
		ctx,
		msgPolicyCheckSucceeded,
		slog.String(keyPolicySource, cfg.policySource),
	)

	return exitOK
}

func policyRules(cfg *runtimeConfig) []policydomain.Rule {
	out := make([]policydomain.Rule, exitOK, len(cfg.rulePatterns))

	for i := range cfg.rulePatterns {
		out = append(out, policydomain.Rule{
			Pattern: cfg.rulePatterns[i],
			Min:     cfg.ruleMins[i],
		})
	}

	return out
}

func printTo(writer io.Writer, text string) error {
	written, err := fmt.Fprint(writer, text)
	if err != nil {
		return fmt.Errorf("fprint: %w", err)
	}

	if written != len(text) {
		return fmt.Errorf("%w: wrote %d of %d", errShortWrite, written, len(text))
	}

	return nil
}

func printUsage(flagSet *flag.FlagSet) {
	err := printTo(os.Stderr, usageHeader)
	if err != nil {
		return
	}

	flagSet.PrintDefaults()

	err = printTo(os.Stderr, usageFooter)
	if err != nil {
		return
	}
}

func registerAnalysisFlags(flagSet *flag.FlagSet, vals *flagValues) {
	vals.workers = flagSet.Int(
		"workers", exitOK, "concurrent package workers (0 = min(GOMAXPROCS, packages))",
	)
	vals.fieldUsage = flagSet.String(
		"field-usage",
		"direct",
		"field usage resolution: direct or transitive",
	)
	vals.dependencyScope = flagSet.String(
		"dependency-scope", "module", "dependency scope: project, module, or all",
	)
	vals.buildTags = flagSet.String("build-tags", emptyString, "comma-separated build tags")
	vals.includeTests = flagSet.Bool("tests", false, "include test files and test packages")
	vals.generated = flagSet.Bool("generated", false, "include generated files")
	vals.continueOnError = flagSet.Bool(
		"continue-on-error", false, "skip packages that fail to load or type-check",
	)
}

func registerMiscFlags(flagSet *flag.FlagSet, vals *flagValues) {
	vals.cpuProfile = flagSet.String("cpu-profile", emptyString, "write a CPU profile to this file")
	vals.memoryProfile = flagSet.String(
		"memory-profile",
		emptyString,
		"write a memory profile to this file",
	)
	vals.showVersion = flagSet.Bool("version", false, "print the version and exit")
	vals.verbose = flagSet.Bool("verbose", false, "verbose logging to stderr")
	vals.check = flagSet.Bool("check", false, "enforce -rule thresholds and exit 3 on violations")
	flagSet.Func(
		"rule",
		"policy rule pattern:min (repeatable; e.g. '**/internal/**':0.8; requires -check)",
		func(value string) error {
			return appendRule(&vals.rules, value)
		},
	)
}

func registerReportFlags(flagSet *flag.FlagSet, vals *flagValues) {
	vals.format = flagSet.String(formatFlag, "text", "report format: text, json, csv, or web")
	vals.webReport = flagSet.Bool(
		"web",
		false,
		"shorthand for -format=web: write a self-contained HTML report and open it",
	)
	vals.output = flagSet.String(
		"output", emptyString, "write the report to this file instead of stdout",
	)
	vals.explain = flagSet.Bool(
		"explain",
		false,
		"include reasons for n/a and dropped-component metrics in the text report",
	)
}

func registerWeightFlags(flagSet *flag.FlagSet, vals *flagValues) {
	defaults := metrics.DefaultReusabilityWeights()

	vals.reusabilityWeightCohesion = flagSet.Float64(
		"reusability-weight-cohesion", defaults.Cohesion, "reusability cohesion component weight",
	)
	vals.reusabilityWeightCoupling = flagSet.Float64(
		"reusability-weight-coupling", defaults.Coupling, "reusability coupling component weight",
	)
	vals.reusabilityWeightTestability = flagSet.Float64(
		"reusability-weight-testability",
		defaults.Testability,
		"reusability testability component weight",
	)
	vals.reusabilityWeightDocumentation = flagSet.Float64(
		"reusability-weight-documentation",
		defaults.Documentation,
		"reusability documentation component weight",
	)
}

func renderHelpDocs(path string) (string, error) {
	fileSink := sinks.FileSink{Path: path}

	err := reporting.WriteDocs(outbound.NewSink(func() (*outbound.Stream, error) {
		return sinks.OpenFile(fileSink)
	}), version.Version())
	if err != nil {
		return emptyString, fmt.Errorf("write help docs: %w", err)
	}

	return path, nil
}

func reportSink(output string) outbound.Sink {
	if output != emptyString {
		fileSink := sinks.FileSink{Path: output}

		return outbound.NewSink(func() (*outbound.Stream, error) {
			return sinks.OpenFile(fileSink)
		})
	}

	return outbound.NewSink(sinks.OpenStdout)
}

func resolveGating(vals *flagValues, logger *slog.Logger) gatingResult {
	if !*vals.check {
		return gatingResult{ok: true}
	}

	resolved, source, err := resolvePolicy(vals.rules)
	if err != nil {
		logger.ErrorContext(
			context.Background(),
			msgPolicyConfigurationFailed,
			slog.Any(keyError, err),
		)

		return gatingResult{code: exitUsage}
	}

	return gatingResult{rules: resolved, source: source, ok: true}
}

func resolvePolicy(rules []ruleSpec) (specs []ruleSpec, source string, err error) {
	if len(rules) == exitOK {
		return nil, emptyString, errNoPolicyRules
	}

	specs = append([]ruleSpec(nil), rules...)

	domainRules := policyDomainRules(specs)

	err = policydomain.Validate(domainRules)
	if err != nil {
		return nil, emptyString, fmt.Errorf("validate policy: %w", err)
	}

	return specs, policySourceFlagRules, nil
}

func policyDomainRules(specs []ruleSpec) []policydomain.Rule {
	rules := make([]policydomain.Rule, exitOK, len(specs))

	for i := range specs {
		rules = append(rules, policydomain.Rule{Pattern: specs[i].pattern, Min: specs[i].minimum})
	}

	return rules
}

func resolveReportFormat(args *buildArgs) formatResult {
	name, code, ok := applyWebFormat(args)

	if !ok {
		return formatResult{code: code}
	}

	format, ok := reportingdomain.ParseFormat(name)

	if !ok {
		args.logger.ErrorContext(
			context.Background(),
			msgInvalidFormat,
			slog.String(formatFlag, name),
			slog.String(keyWant, wantFormatValues),
		)

		return formatResult{code: exitUsage}
	}

	return formatResult{format: string(format), ok: true}
}

func resolveWeights(vals *flagValues, logger *slog.Logger) weightsResult {
	weights := reusability.Weights{
		Cohesion:      *vals.reusabilityWeightCohesion,
		Coupling:      *vals.reusabilityWeightCoupling,
		Testability:   *vals.reusabilityWeightTestability,
		Documentation: *vals.reusabilityWeightDocumentation,
	}

	err := weights.Validate()
	if err != nil {
		logger.ErrorContext(
			context.Background(),
			msgAnalysisFailed,
			slog.Any(keyError, err),
		)

		return weightsResult{code: exitFail}
	}

	return weightsResult{weights: weights, ok: true}
}

func ruleSlices(rules []ruleSpec) (patterns []string, mins []float64) {
	patterns = make([]string, exitOK, len(rules))
	mins = make([]float64, exitOK, len(rules))

	for i := range rules {
		patterns = append(patterns, rules[i].pattern)
		mins = append(mins, rules[i].minimum)
	}

	return patterns, mins
}

func runAnalyze(
	ctx context.Context,
	cfg *runtimeConfig,
	analyze analyzeFunc,
) (report reusability.Report, code int) {
	start := time.Now()

	report, err := analyze(ctx, &cfg.analysis)
	if err != nil {
		return reusability.Report{}, failAnalyze(ctx, cfg.logger, err)
	}

	cfg.logger.DebugContext(
		ctx,
		msgAnalysisComplete,
		slog.Int(keyPackages, len(report.Packages)),
		slog.Duration(keyDuration, time.Since(start)),
	)

	return report, exitOK
}

func runPipeline(ctx context.Context, cfg *runtimeConfig, analyze analyzeFunc) int {
	stopCPU, code := maybeStartCPU(ctx, cfg)

	if code != exitOK {
		return code
	}

	if stopCPU != nil {
		defer stopCPUProfile(ctx, cfg.logger, stopCPU)
	}

	return analyzeThenWrite(ctx, cfg, analyze)
}

func runWebHelp() int {
	logger := newLogger(nil)

	path, err := writeHelpDocs()
	if err != nil {
		logger.ErrorContext(
			context.Background(),
			msgWritingGuideFailed,
			slog.Any(keyError, err),
		)

		return exitFail
	}

	logWebHelpWritten(logger, path)

	return exitOK
}

func appendRule(rules *[]ruleSpec, value string) error {
	spec, err := parseRuleSpec(value)
	if err != nil {
		return fmt.Errorf("set rule: %w", err)
	}

	*rules = append(*rules, spec)

	return nil
}

func splitList(list string) []string {
	if list == emptyString {
		return nil
	}

	parts := strings.Split(list, commaSep)
	out := make([]string, exitOK, len(parts))

	for i := range parts {
		if part := strings.TrimSpace(parts[i]); part != emptyString {
			out = append(out, part)
		}
	}

	return out
}

func stopCPUProfile(ctx context.Context, logger *slog.Logger, stopProfile func() error) {
	err := stopProfile()
	if err != nil {
		logger.ErrorContext(
			ctx,
			msgCPUProfilingFailed,
			slog.Any(keyError, err),
		)
	}
}

func stdoutIsTerminal() bool {
	info, err := os.Stdout.Stat()

	return err == nil && info.Mode()&os.ModeCharDevice != exitOK
}

func wantsWebHelp(args []string) bool {
	for i := range args {
		if args[i] == "--" {
			return false
		}

		if isTruthyWebFlag(args[i]) {
			return true
		}
	}

	return false
}

func webEqualsTruthy(arg string) bool {
	value, found := strings.CutPrefix(arg, "-web=")

	if !found {
		value, found = strings.CutPrefix(arg, webFlagLong+"=")
	}

	if !found {
		return false
	}

	truthy, err := strconv.ParseBool(value)

	return err == nil && truthy
}

func writeAndGate(ctx context.Context, cfg *runtimeConfig, report *reusability.Report) int {
	if code := writeReport(ctx, cfg, report); code != exitOK {
		return code
	}

	openWebIfNeeded(ctx, cfg)

	return enforcePolicy(ctx, cfg, report)
}

func writeHeapIfNeeded(ctx context.Context, cfg *runtimeConfig) int {
	if cfg.memoryProfile == emptyString {
		return exitOK
	}

	err := profiling.WriteHeap(cfg.memoryProfile)
	if err != nil {
		cfg.logger.ErrorContext(
			ctx,
			msgMemoryProfilingFailed,
			slog.Any(keyError, err),
		)

		return exitFail
	}

	return exitOK
}

func writeHelpDocs() (string, error) {
	file, err := os.CreateTemp(emptyString, helpTempPattern)
	if err != nil {
		return emptyString, fmt.Errorf("create help temp: %w", err)
	}

	path, err := finalizeHelpDocs(file, file.Name())
	if err != nil {
		return emptyString, fmt.Errorf("finalize help docs: %w", err)
	}

	return path, nil
}

func writeReport(ctx context.Context, cfg *runtimeConfig, report *reusability.Report) int {
	err := reporting.Write(buildWriteRequest(cfg, report))
	if err != nil {
		cfg.logger.ErrorContext(
			ctx,
			msgWritingReportFailed,
			slog.Any(keyError, err),
		)

		return exitFail
	}

	return exitOK
}

func writeViolations(ctx context.Context, logger *slog.Logger, items []policydomain.Violation) {
	err := printTo(os.Stderr, policydomain.FormatViolations(items))
	if err != nil {
		logger.ErrorContext(ctx, msgWritingViolationsFailed, slog.Any(keyError, err))
	}
}
