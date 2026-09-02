// Command reusability reports a type-level reusability index for named types
// in a Go module.
package main

import (
	"os"

	"github.com/gostafa/reusability/internal/cli"
)

var (
	run  = cli.Run
	exit = os.Exit
)

func main() {
	exit(run(os.Args[1:]))
}
