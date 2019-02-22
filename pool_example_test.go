package workerpool_test

import (
	"context"
	"fmt"

	"github.com/sammyne/workerpool"
)

func ExamplePool() {
	say := func(what string) workerpool.ActionFunc {
		return workerpool.ActionFunc(
			func(context.Context) error {
				fmt.Println(what)
				return nil
			})
	}

	pool := workerpool.Pool(2)

	ctx := context.TODO()
	actions := []workerpool.Action{
		say("how"),
		say("do"),
		say("you"),
		say("do"),
	}

	for response := range pool.Execute(ctx, actions) {
		if nil != response {
			fmt.Println("unexpected response:", response)
		}
	}

	pool.Close()

	for response := range pool.Execute(ctx, actions) {
		if response != workerpool.ErrClosed {
			fmt.Printf("unexpected response: got %v, expect %v", response,
				workerpool.ErrClosed)
		}
	}

	// Unordered output:
	// do
	// you
	// do
	// how
}
