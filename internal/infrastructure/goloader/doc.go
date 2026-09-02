// Package goloader adapts the Go compiler toolchain into type facts.
//
// It is the infrastructure boundary that imports go/packages, go/types, and
// go/ast, producing extracted facts instead of metric results.
package goloader
