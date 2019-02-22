package workerpool

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
)

type pool struct {
	todos chan *todo

	quit     chan struct{}
	quitOnce sync.Once
	execWG   sync.WaitGroup
	workerWG sync.WaitGroup
}

func (p *pool) Close() error {
	err := ErrClosed

	p.quitOnce.Do(func() {
		close(p.quit)
		// TODO: order of waiting should be taken more serious consideration later
		p.workerWG.Wait() // wait for exit of workers
		p.execWG.Wait()   // wait for exit of all active actions

		close(p.todos)
		// drain the todos channel and invoke its callback
		// to achieve a graceful quit
		for pending := range p.todos {
			pending.done(ErrClosed)
		}

		err = nil
	})

	return err
}

// Execute delivers all input Actions on the worker pool, and get the a
// stream for reading out responses. The close of the response steam is
// handled by the pool internally, and it signals a done (no error, or
// error out) of all actions submit in this batch.
// The error handling is delegated to the caller who is responsible of
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

	// one more slot for early quit or cancellation
	responses := make(chan error, int(nPending)+1)
	// callback handling error responsed from any todo job
	doneCallback := func(err error) {
		if nil != err {
			// TODO: more serious tackling of the case of closed responses later
			responses <- err
		}

		if 0 == atomic.AddInt32(&nPending, -1) {
			close(responses)
		}
	}

	// a composite done channel at the expense of a out-of-pool goroutine
	done := orDone.Done()
enqueue:
	for i, action := range actions {
		pending := &todo{orDone, action, doneCallback}
		select {
		case <-done:
			// drain the pending action list and error out,
			// which will also close the responses stream eventually to
			// signal a truly done of this Execute
			for ; i < len(actions); i++ {
				doneCallback(orDone.Err())
			}

			break enqueue
		case p.todos <- pending: // enqueue action
		}
	}

	return responses
}

// fork a worker responsible of taking job from p.todos to do
func (p *pool) fork() {
	defer p.workerWG.Done()

	for {

		// favor quit checking
		select {
		case <-p.quit:
			return
		default:
		}

		select {
		case <-p.quit:
			return
		case todo := <-p.todos:
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
		todos: make(chan *todo, n),
		quit:  make(chan struct{}),
	}

	for i := 0; i < n; i++ {
		p.workerWG.Add(1)
		go p.fork()
	}

	return p
}
