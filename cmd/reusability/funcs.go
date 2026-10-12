// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"os"
)

func main() {
	defaultRuntime().start(os.Args[1:])
}

func defaultRuntime() mainRuntime {
	return mainRuntime{run: run, exit: exit}
}

func (runtime mainRuntime) start(args []string) {
	runtime.exit(runtime.run(args))
}
