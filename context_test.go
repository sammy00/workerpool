package workerpool

import (
	"context"
	"testing"
)

func TestOrDone_Done(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	close(done)

	// none of below shoule be blocking
	// otherwise, the timeout stemming from blocking will fail the test
	testCases := []*orDone{
		&orDone{Context: ctx, done: make(chan struct{})},
		&orDone{context.TODO(), done},
	}

	for _, c := range testCases {
		<-c.Done()
	}
}

func TestOrDone_Err(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	close(done)

	testCases := []struct {
		orDone *orDone
		expect error
	}{
		{&orDone{ctx, make(chan struct{})}, ctx.Err()},
		{&orDone{context.TODO(), done}, ErrClosed},
		{&orDone{context.TODO(), make(chan struct{})}, nil},
	}

	for i, c := range testCases {
		if got := c.orDone.Err(); got != c.expect {
			t.Fatalf("#%d unexpected error", i)
		}
	}
}
