package workerpool

import (
	"context"
	"runtime"
	"testing"
)

func TestPool(t *testing.T) {
	numCPU := runtime.NumCPU()

	testCases := []struct {
		n      int
		expect int
	}{
		{-2, numCPU},
		{-1, numCPU},
		{0, numCPU},
		{1, 1},
		{2, 2},
	}

	for i, c := range testCases {
		exec := Pool(c.n)

		pool := exec.(*pool)
		if cap(pool.pendings) != c.expect {
			t.Fatalf("#%d failed: got pool size as %d, expect %d",
				i, cap(pool.pendings), c.expect)
		}

		exec.Close()
	}
}

// this test demonstrate the scenario when all workers have quit,
// followed by the closing operation without a non-empty pending queue
func TestPool_Close_drainTodos(t *testing.T) {
	dummyJob := func(context.Context) error {
		return nil
	}

	var cbErr error
	doneSpy := func(err error) {
		cbErr = err
	}

	pool := Pool(2).(*pool)
	pool.execWG.Add(1)

	done := make(chan struct{})
	go func() {
		pool.Close()
		close(done)
	}()

	pool.workerWG.Wait()
	pool.pendings <- &todo{
		context.TODO(),
		ActionFunc(dummyJob),
		doneSpy,
	}
	pool.execWG.Done()

	<-done

	if cbErr != ErrClosed {
		t.Fatalf("unexpected error: got %v, expect %v", cbErr, ErrClosed)
	}
}
