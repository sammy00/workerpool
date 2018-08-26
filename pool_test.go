package workerpool_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/sammy00/workerpool"
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

func TestPool_Execute_ErrAction(t *testing.T) {
	pool, cancel := workerpool.Pool(1)
	defer cancel()

	ctx := context.TODO()
	actions := []workerpool.Action{
		workerpool.ActionFunc(Sleep),
		workerpool.ActionFunc(SayErr),
		workerpool.ActionFunc(Sleep),
	}

	start := time.Now()

	expectErr := "error occurs"
	if err := pool.Execute(ctx, actions...); (nil == err) ||
		(err.Error() != expectErr) {
		t.Fatalf("invalid error: got (%s), expect (%s)", err.Error(), expectErr)
	}

	duration := time.Now().Sub(start)

	// it should takes approximately 2 seconds (by the 1st Sleep())
	// or a little more
	if (duration < time.Second*2) || (duration > time.Second*3) {
		t.Fatalf("the time elapased is out of range: %v", duration)
	}
}

func TestPool_Execute_CancelledByCaller(t *testing.T) {
	pool, cancel := workerpool.Pool(1)
	defer cancel()

	ctx, cancel2 := context.WithCancel(context.Background())

	var err error
	done := make(chan bool)

	go func() {
		actions := []workerpool.Action{
			workerpool.ActionFunc(Sleep),
			workerpool.ActionFunc(Sleep),
		}
		err = pool.Execute(ctx, actions...)
		done <- true
	}()

	cancel2()
	<-done

	expect := "context canceled"
	if nil == err {
		t.Fatal("got no error, expect one")
	} else if err.Error() != expect {
		t.Fatalf("got error description as (%s), but expect %s",
			err.Error(), expect)
	}
}

func TestPool_Execute_Close(t *testing.T) {
	pool, cancel := workerpool.Pool(1)
	defer cancel()

	var err error
	done := make(chan bool)

	go func() {
		ctx := context.TODO()
		actions := []workerpool.Action{
			workerpool.ActionFunc(Sleep),
			workerpool.ActionFunc(Sleep),
		}
		err = pool.Execute(ctx, actions...)
		done <- true
	}()

	cancel()
	<-done

	expect := "pool is closed"
	if nil == err {
		t.Fatal("got no error, expect one")
	} else if err.Error() != expect {
		t.Fatalf("got error description as (%s), but expect %s",
			err.Error(), expect)
	}
}

func TestPool_Execute_NoAction(t *testing.T) {
	pool, cancel := workerpool.Pool(1)
	defer cancel()

	ctx := context.TODO()
	if err := pool.Execute(ctx); nil != err {
		t.Fatalf("unexpected error: %v", err)
	}
}
