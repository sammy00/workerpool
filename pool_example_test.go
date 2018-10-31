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
	//pool, cancel := workerpool.Pool(2)
	//defer cancel()
	pool := workerpool.Pool(2)
	defer pool.Close()

	ctx := context.TODO()
	actions := []workerpool.Action{
		workerpool.ActionFunc(SayHello),
		workerpool.ActionFunc(SayWorld),
		workerpool.ActionFunc(SayWorld),
		workerpool.ActionFunc(SayHello),
	}

	if err := pool.Execute(ctx, actions, false); nil != err {
		fmt.Println(err)
		return
	}

	// Unordered output:
	// hello
	// world
	// world
	// hello
}
