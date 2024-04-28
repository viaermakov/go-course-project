package errgrp

import (
	"context"
	"sync"
	"sync/atomic"
)

// ErrGr is a type that represents an error group. It allows you to run multiple
// asynchronous tasks concurrently, and wait for all tasks to complete or for an
// error to occur. If any task returns an error, the Wait() function will return that error.
type ErrGr struct {
	errc   chan error
	count  atomic.Int64
	cancel func(error)
	sync.Mutex
}

// WithContext takes a context and returns a new ErrGr and a modified context.
// The cancel function inside ErrGr is used to signal cancellation to all goroutines.
func WithContext(ctx context.Context) (*ErrGr, context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	gr := &ErrGr{cancel: cancel}
	return gr, ctx
}

// Go runs the specified callback function in a separate goroutine within the ErrGr object.
// it launches a new goroutine to execute the callback function and sends the result to the error channel.
func (errGr *ErrGr) Go(clb func() error) {
	errGr.count.Add(1)

	// there is no constructor, we need to initialize on the first run of the Go method
	if errGr.errc == nil {
		errGr.errc = make(chan error, 1)
	}

	go func() {
		errGr.errc <- clb()
		errGr.count.Add(-1)
		if errGr.count.Load() == 0 {
			close(errGr.errc)
		}
	}()
}

// Wait waits for all the goroutines launched with the Go method to finish and returns the first non-nil error encountered.
// It closes the error channel to signal that no more errors will be received.
func (errGr *ErrGr) Wait() error {
	var firstErr error

	for curErr := range errGr.errc {
		if curErr != nil && firstErr == nil {
			firstErr = curErr

			if errGr.cancel != nil {
				// cancel ctx to finish all other goroutines
				errGr.cancel(curErr)
			}
		}
	}

	return firstErr
}
