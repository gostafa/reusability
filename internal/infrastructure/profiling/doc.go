// Package profiling wraps runtime/pprof for the CLI's CPU and heap profile flags.
//
// Callers start a CPU profile before analysis and write a heap profile after it
// finishes; failures are returned to the CLI rather than handled here.
package profiling
