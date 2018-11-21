package workerpool

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
)

type poolAction struct {
	ctx      context.Context
	action   Action
	response chan<- error
}

type pool struct {
	pendings chan poolAction

	quit     chan struct{}
	quitOnce sync.Once
	status   Status
	execWG   sync.WaitGroup
}

func (p *pool) Close() error {
	err := ErrClosed

	p.quitOnce.Do(func() {
		atomic.StoreInt32(&p.status, Stopped)
		close(p.quit)
		p.execWG.Wait() // wait for exit of all active actions
		close(p.pendings)

		err = nil
	})

	return err
}

// Execute enqueues all Actions on the worker pool, failing closed on the
// first error or if ctx is cancelled. This method blocks until all enqueued
// Actions have returned. In the event of an error, not all Actions may be
// executed.
//func (p pool) Execute(ctx context.Context, actions ...Action) error {
func (p *pool) Execute(ctx context.Context, actions []Action,
	failFast ...bool) error {
	p.execWG.Add(1)
	defer p.execWG.Done()

	if Runnable != atomic.LoadInt32(&p.status) {
		return ErrClosed
	}

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
		case <-p.quit: // pool is closed
			return ErrClosed
		case <-ctx.Done(): // ctx is closed by caller
			err = ctx.Err()
			break enqueue
		case p.pendings <- pa: // enqueue action
			queued++
		}
	}

	for ; queued > 0; queued-- {
		if r := <-res; r != nil && nil == err {
			err = r

			if len(failFast) > 0 && failFast[0] {
				// should cancel all running jobs if fail fast is required
				cancel()
			}
		}
	}

	return err
}

// fork a worker responsible of taking job from p.in to do
func (p *pool) fork() {
	defer p.execWG.Done()

	for {
		select {
		case <-p.quit:
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
func Pool(n int) Executor {
	if n <= 0 {
		n = runtime.NumCPU()
	}

	p := &pool{
		pendings: make(chan poolAction, n),
		quit:     make(chan struct{}),
		status:   Runnable,
	}

	for i := 0; i < n; i++ {
		p.execWG.Add(1)
		go p.fork()
	}

	return p
}
