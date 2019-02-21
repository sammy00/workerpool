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

	testCases := []struct {
		orDone *orDone
		expect <-chan struct{}
	}{
		{&orDone{Context: ctx, done: make(chan struct{})}, ctx.Done()},
		{&orDone{context.TODO(), done}, done},
	}

	for i, c := range testCases {
		if got := c.orDone.Done(); got != c.expect {
			t.Fatalf("#%d unexpected done channel", i)
		}
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
