// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package cli

import (
	"errors"
	"os"

	reporting "github.com/gostafa/reusability/internal/features/reporting/application"
	"github.com/gostafa/reusability/internal/infrastructure/browser"
	"github.com/gostafa/reusability/internal/infrastructure/profiling"
	"github.com/gostafa/reusability/reusability"
)

var (
	errEmptyPattern       = errors.New("empty pattern in rule spec")
	errExpectedPatternMin = errors.New("expected pattern:min")
	errNoPolicyRules      = errors.New(
		"no policy rules configured; pass at least one -rule=pattern:min with -check",
	)
	errShortWrite = errors.New("fprint: short write")

	// Test hooks — override in tests; production uses the defaults below.
	analyzeFn        = reusability.Analyze
	isTerminalFn     = stdoutIsTerminal
	createHelpTempFn = os.CreateTemp
	closeHelpFileFn  = closeOSFile
	writeDocsFn      = reporting.WriteDocs
	openBrowserFn    = browser.Open
	startCPUFn       = profiling.StartCPU
	writeHeapFn      = profiling.WriteHeap
)
