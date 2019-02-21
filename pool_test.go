package workerpool_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/sammyne/workerpool"
)

func SayErr(ctx context.Context) error {
	fmt.Println("error")
	return errors.New("error occurs")
}

func Sleep(ctx context.Context) error {
	var err error

	select {
	case <-ctx.Done():
		err = ctx.Err()
		fmt.Println("action is cancelled")
	case <-time.After(time.Second * 2):
		fmt.Println("sleep is enough")
	}

	return err
}

func TestPool_Close(t *testing.T) {
	testCases := []struct {
		worker workerpool.Executor
		nClose int
		expect error
	}{
		{workerpool.Pool(2), 1, nil},
		{workerpool.Pool(2), 2, workerpool.ErrClosed},
	}

	for i, c := range testCases {
		var got error
		for ; c.nClose > 0; c.nClose-- {
			got = c.worker.Close()
		}

		if got != c.expect {
			t.Fatalf("#%d unexpected error: got %v, expect %v", i, got, c.expect)
		}
	}
}
func TestPool_Execute_afterQuit(t *testing.T) {
	pool := workerpool.Pool(3)
	pool.Close()

	dummyJob := workerpool.ActionFunc(
		func(ctx context.Context) error {
			time.Sleep(time.Millisecond * 500)
			return nil
		})
	jobs := []workerpool.Action{dummyJob, dummyJob, dummyJob, dummyJob}

	for response := range pool.Execute(context.TODO(), jobs) {
		if response != workerpool.ErrClosed {
			t.Fatalf("unexpected error: got %v, expect %v", response,
				workerpool.ErrClosed)
		}
	}
}

func TestPool_Execute_earlyQuit(t *testing.T) {
	pool := workerpool.Pool(1)
	//pool.Close()

	//var doing int32
	doing := make(chan struct{}, 3)
	dummyJob := workerpool.ActionFunc(
		func(ctx context.Context) error {
			//atomic.AddInt32(&doing, 1)
			doing <- struct{}{}
			<-ctx.Done()
			return ctx.Err()
		})
	jobs := []workerpool.Action{dummyJob, dummyJob, dummyJob}

	done := make(chan struct{})
	go func() {
		defer close(done)

		//for 0 == atomic.LoadInt32(&doing) {
		//}
		<-doing

		pool.Close()
	}()

	for response := range pool.Execute(context.TODO(), jobs) {
		if response != workerpool.ErrClosed {
			t.Fatalf("unexpected error: got %v, expect %v", response,
				workerpool.ErrClosed)
		}
	}

	<-done
}

func TestPool_Execute_noJob(t *testing.T) {
	pool := workerpool.Pool(3)
	defer pool.Close()

	<-pool.Execute(context.TODO(), nil)
}

func TestPool_Execute_ok(t *testing.T) {
	dummyJob := workerpool.ActionFunc(
		func(ctx context.Context) error {
			time.Sleep(time.Millisecond * 500)
			return nil
		})
	jobs := []workerpool.Action{dummyJob, dummyJob, dummyJob, dummyJob}

	pool := workerpool.Pool(3)
	defer pool.Close()

	responses := pool.Execute(context.TODO(), jobs)
	for response := range responses {
		if nil != response {
			t.Fatalf("unexpected error: %v", response)
		}
	}
}
