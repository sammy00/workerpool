package workerpool

import (
	"context"
	"testing"
)

func TestPoolAction_Execute_ok(t *testing.T) {
	dummyJob := func(context.Context) error {
		return nil
	}
	response := make(chan error, 1)
	var nPendingPeers int32 = 123

	var called bool
	doneSpy := func(err ...error) { called = true }

	action := &poolAction{
		context.TODO(),
		ActionFunc(dummyJob),
		response,
		&nPendingPeers,
		doneSpy,
	}

	action.Execute()

	// expectation checking
	if 0 != len(action.response) {
		t.Fatalf("unexpected #(response): got %d, expect %d",
			len(action.response), 0)
	}

	if 122 != nPendingPeers {
		t.Fatalf("unexpected nPendingPeers: got %d, expect %d", nPendingPeers, 122)
	}

	if called {
		t.Fatalf("done callback shouldn't be called")
	}
}

func TestPoolAction_Execute_error(t *testing.T) {
	dummyJob := func(context.Context) error {
		return context.Canceled
	}

	response := make(chan error, 1)
	var nPendingPeers int32 = 123

	doneSpy := func(err ...error) {}

	action := &poolAction{
		context.TODO(),
		ActionFunc(dummyJob),
		response,
		&nPendingPeers,
		doneSpy,
	}

	action.Execute()

	// expectation checking
	select {
	case err := <-response:
		if err != context.Canceled {
			t.Fatalf("unexpected error: got %v, expect %v", err, context.Canceled)
		}
	default:
		t.Fatalf("missing error")
	}
}

func TestPoolAction_Execute_last(t *testing.T) {
	dummyJob := func(context.Context) error {
		return nil
	}
	response := make(chan error, 1)
	var nPendingPeers int32 = 1

	var called bool
	doneSpy := func(err ...error) { called = true }

	action := &poolAction{
		context.TODO(),
		ActionFunc(dummyJob),
		response,
		&nPendingPeers,
		doneSpy,
	}

	action.Execute()

	// expectation checking
	if 0 != len(action.response) {
		t.Fatalf("unexpected #(response): got %d, expect %d",
			len(action.response), 0)
	}

	if !called {
		t.Fatalf("done callback should be called")
	}
}
