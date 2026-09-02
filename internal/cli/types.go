// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package cli

import (
	"context"
	"flag"
	"log/slog"
	"os"

	policydomain "github.com/gostafa/reusability/internal/features/policy/domain"
	reportingdomain "github.com/gostafa/reusability/internal/features/reporting/domain"
	"github.com/gostafa/reusability/internal/features/reporting/ports/outbound"
	"github.com/gostafa/reusability/reusability"
)

type (
	cliDeps struct {
		analyze        func(context.Context, *reusability.Config) (reusability.Report, error)
		isTerminal     func() bool
		createHelpTemp func(dir, pattern string) (*os.File, error)
		closeHelpFile  func(*os.File) error
		writeDocs      func(outbound.Sink, string) error
		openBrowser    func(string) error
		startCPU       func(string) (func() error, error)
		writeHeap      func(string) error
	}

	ruleSpec struct {
		pattern string
		minimum float64
	}

	ruleList struct {
		items []ruleSpec
	}

	flagValues struct {
		format                         *string
		webReport                      *bool
		output                         *string
		explain                        *bool
		workers                        *int
		fieldUsage                     *string
		dependencyScope                *string
		reusabilityWeightCohesion      *float64
		reusabilityWeightCoupling      *float64
		reusabilityWeightTestability   *float64
		reusabilityWeightDocumentation *float64
		buildTags                      *string
		includeTests                   *bool
		generated                      *bool
		continueOnError                *bool
		cpuProfile                     *string
		memoryProfile                  *string
		showVersion                    *bool
		verbose                        *bool
		check                          *bool
		rules                          ruleList
	}

	runtimeConfig struct {
		logger        *slog.Logger
		deps          *cliDeps
		format        reportingdomain.Format
		output        string
		policySource  string
		cpuProfile    string
		memoryProfile string
		rules         []policydomain.Rule
		analysis      reusability.Config
		explain       bool
		gating        bool
		webToDefault  bool
	}

	parseResult struct {
		cfg  runtimeConfig
		code int
		done bool
	}

	gatingResult struct {
		source string
		rules  []policydomain.Rule
		code   int
		ok     bool
	}

	formatResult struct {
		format reportingdomain.Format
		code   int
		ok     bool
	}

	weightsResult struct {
		weights reusability.Weights
		code    int
		ok      bool
	}

	buildArgs struct {
		flagSet *flag.FlagSet
		vals    *flagValues
		logger  *slog.Logger
	}

	assembleArgs struct {
		in      buildArgs
		format  reportingdomain.Format
		gating  gatingResult
		weights reusability.Weights
	}
)
