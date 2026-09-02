// Package workerpool runs indexed tasks on a bounded number of goroutines.
//
// Tasks write to their own result slots so callers can merge results
// deterministically after Run returns.
package workerpool
