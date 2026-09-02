// Package store defines persistence contracts.
package store

// Store persists values by key.
type Store interface {
	Put(id int, v any) error
}
