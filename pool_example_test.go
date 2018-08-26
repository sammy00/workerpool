package workerpool_test

import (
	"context"
	"fmt"

	"github.com/sammy00/workerpool"
)

func SayHello(ctx context.Context) error {
	fmt.Println("hello")
	return nil
}

func SayWorld(ctx context.Context) error {
	fmt.Println("world")
	return nil
}

func ExamplePool() {
	pool, cancel := workerpool.Pool(2)
	defer cancel()

	ctx := context.TODO()
	actions := []workerpool.Action{
		workerpool.ActionFunc(SayHello),
		workerpool.ActionFunc(SayWorld),
		workerpool.ActionFunc(SayWorld),
		workerpool.ActionFunc(SayHello),
	}

	if err := pool.Execute(ctx, actions...); nil != err {
		fmt.Println(err)
		return
	}

	// Unordered output:
	// hello
	// world
	// world
	// hello
}
