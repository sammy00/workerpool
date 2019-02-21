package workerpool

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
)

type pool struct {
	//pendings chan *poolAction
	pendings chan *todo

	quit     chan struct{}
	quitOnce sync.Once
	execWG   sync.WaitGroup
	workerWG sync.WaitGroup
}

func (p *pool) Close() error {
	err := ErrClosed

	p.quitOnce.Do(func() {
		close(p.quit)
		// order of waiting should be taken more serious consideration later
		p.workerWG.Wait() // wait for exit of workers
		p.execWG.Wait()   // wait for exit of all active actions

		close(p.pendings)
		// drain the pendings channel the invoke the callback
		// to achieve a graceful quit
		for pending := range p.pendings {
			//pending.doneCallback(ErrClosed)
			pending.done(ErrClosed)
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
		responses := make(chan error)
		close(responses)
		return responses
	}

	// unblocked in case of cancellation or quit
	orDone := &orDone{ctx, p.quit}

	responses := make(chan error, int(nPending)+1)
	doneCallback := func(err error) {
		if nil != err {
			responses <- err
		}

		if 0 == atomic.AddInt32(&nPending, -1) {
			close(responses)
		}
	}

	done := orDone.Done()
enqueue:
	for i, action := range actions {
		pending := &todo{orDone, action, doneCallback}
		select {
		case <-done:
			// drain the pending action list and error out
			for ; i < len(actions); i++ {
				doneCallback(orDone.Err())
			}

			break enqueue
		case p.pendings <- pending: // enqueue action
		}
	}

	return responses
}

// fork a worker responsible of taking job from p.in to do
func (p *pool) fork() {
	//defer p.execWG.Done()
	defer p.workerWG.Done()

	for {

		// favor quit checking
		select {
		case <-p.quit:
		default:
		}

		select {
		case <-p.quit:
			return
		case todo := <-p.pendings:
			//a.Execute()
			todo.done(todo.action.Execute(todo.ctx))
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
		//pendings: make(chan *poolAction, n),
		pendings: make(chan *todo, n),
		quit:     make(chan struct{}),
	}

	for i := 0; i < n; i++ {
		//p.execWG.Add(1)
		p.workerWG.Add(1)
		go p.fork()
	}

	return p
}
