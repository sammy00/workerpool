package workerpool

import "errors"

var (
	// ErrClosed means the pool has been shutdown
	ErrClosed = errors.New("pool already closed")
)
