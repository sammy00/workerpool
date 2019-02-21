package workerpool

import (
	"context"
	"fmt"
)

type orDone struct {
	context.Context
	done <-chan struct{}
}

func (ctx *orDone) Done() <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)

		select {
		case <-ctx.Context.Done(): // closed by caller through context
			return
		case <-ctx.done: // closed by caller through channel
			return
		}
	}()

	return done
}

func (ctx *orDone) Err() error {
	select {
	case <-ctx.Context.Done(): // ctx is closed by caller
		fmt.Println("111111")
		return ctx.Context.Err()
	case <-ctx.done:
		return ErrClosed
	default:
	}

	return nil
}
