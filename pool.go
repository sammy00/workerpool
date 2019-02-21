package workerpool

import (
	"context"
	"runtime"
	"sync"
)

type pool struct {
	pendings chan *poolAction

	quit     chan struct{}
	quitOnce sync.Once
	execWG   sync.WaitGroup
}

func (p *pool) Close() error {
	err := ErrClosed

	p.quitOnce.Do(func() {
		//atomic.StoreInt32(&p.status, Stopped)
		close(p.quit)
		p.execWG.Wait() // wait for exit of all active actions

		close(p.pendings)
		// drain the pendings channel the invoke the callback
		// to achieve a graceful quit
		for pending := range p.pendings {
			pending.doneCallback(ErrClosed)
		}

		err = nil
	})

	return err
}

// Execute enqueues all Actions on the worker pool, failing closed on the
// first error or if ctx is cancelled. This method blocks until all enqueued
// Actions have returned. In the event of an error, not all Actions may be
// executed.
//func (p pool) Execute(ctx context.Context, actions ...Action) error {
// delegate the error handling to the caller who is responsible of
// implementing any cancellation mechanism as she/he wants
func (p *pool) Execute(ctx context.Context, actions []Action) <-chan error {
	p.execWG.Add(1)
	defer p.execWG.Done()

	select {
	case <-p.quit:
		responses := make(chan error, 1)
		responses <- ErrClosed
		close(responses)

		return responses
	default:
	}

	nPending := int32(len(actions))
	if nPending == 0 {
		return nil
	}

	// unblocked in case of cancellation or quit
	orDone := &orDone{ctx, p.quit}

	responses := make(chan error, int(nPending)+1)
	var oncer sync.Once
	closer := func(err ...error) {
		oncer.Do(func() {
			if len(err) > 0 {
				responses <- err[0]
			}

			close(responses)
		})
	}

enqueue:
	for _, action := range actions {
		pending := &poolAction{orDone, action, responses, &nPending, closer}
		select {
		case <-orDone.Done():
			responses <- orDone.Err()
			break enqueue
		case p.pendings <- pending: // enqueue action
		}
	}

	return responses
}

// fork a worker responsible of taking job from p.in to do
func (p *pool) fork() {
	defer p.execWG.Done()

	for {
		select {
		case <-p.quit:
			return
		case a := <-p.pendings:
			a.Execute()
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
		pendings: make(chan *poolAction, n),
		quit:     make(chan struct{}),
	}

	for i := 0; i < n; i++ {
		p.execWG.Add(1)
		go p.fork()
	}

	return p
}

/*
func newResponseStream(size int) (chan error, func()) {

	responses := make(chan error, size)

	var oncer sync.Once
	closer := func() {
		oncer.Do(func() { close(responses) })
	}

	return responses, closer
}
*/
