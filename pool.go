package workerpool

import (
	"context"
	"errors"
	"runtime"
)

type poolAction struct {
	ctx      context.Context
	action   Action
	response chan<- error
}

type pool struct {
	ctx      context.Context
	cancel   context.CancelFunc
	pendings chan poolAction
}

func (p pool) Close() error {
	if nil == p.cancel {
		return errors.New("pool isn't active")
	}

	p.cancel()
	return nil
}

// Execute enqueues all Actions on the worker pool, failing closed on the
// first error or if ctx is cancelled. This method blocks until all enqueued
// Actions have returned. In the event of an error, not all Actions may be
// executed.
func (p pool) Execute(ctx context.Context, actions ...Action) error {
	qty := len(actions)
	if qty == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	res := make(chan error, qty)

	var err error
	var queued uint64

enqueue:
	for _, action := range actions {
		pa := poolAction{ctx: ctx, action: action, response: res}
		select {
		//case <-p.done: // pool is closed
		case <-p.ctx.Done(): // pool is closed
			cancel()
			return errors.New("pool is closed")
		case <-ctx.Done(): // ctx is closed by caller
			err = ctx.Err()
			break enqueue
		case p.pendings <- pa: // enqueue action
			queued++ // double-check if thread-safe needed
		}
	}

	// fail fast
	for ; queued > 0; queued-- {
		if r := <-res; r != nil {
			if err == nil {
				err = r
				cancel()
			}
		}
	}

	return err
}

// fork a worker responsible of taking job from p.in to do
func (p pool) fork() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case a := <-p.pendings:
			a.response <- a.action.Execute(a.ctx)
		}
	}
}

// Pool creates an Executor backed by a concurrent worker pool. Up to n Actions
// can be in-flight simultaneously; if n is less than or equal to zero,
// runtime.NumCPU is used. The done channel should be closed to release
// resources held by the Executor.
//func Pool(n int, done <-chan struct{}) Executor {
//func Pool(n int) (Executor, context.CancelFunc) {
func Pool(n int) Executor {
	if n <= 0 {
		n = runtime.NumCPU()
	}

	ctx, cancel := context.WithCancel(context.Background())
	p := pool{ctx: ctx, cancel: cancel, pendings: make(chan poolAction, n)}

	for i := 0; i < n; i++ {
		go p.fork()
	}

	return p
}
