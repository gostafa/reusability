// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package cli

import (
	"flag"
	"log/slog"

	"github.com/gostafa/reusability/reusability"
)

type (
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

	// runtimeConfig holds parsed CLI runtime state. Named module types are
	// limited to analysis config so coupling stays low.
	runtimeConfig struct {
		logger        *slog.Logger
		format        string
		output        string
		policySource  string
		cpuProfile    string
		memoryProfile string
		rulePatterns  []string
		ruleMins      []float64
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
		rules  []ruleSpec
		code   int
		ok     bool
	}

	formatResult struct {
		format string
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
)
