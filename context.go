package workerpool

import (
	"context"
)

type orDone struct {
	context.Context
	done <-chan struct{}
}

func (ctx *orDone) Done() <-chan struct{} {
	select {
	case <-ctx.Context.Done(): // closed by caller through context
		return ctx.Context.Done()
	case <-ctx.done: // closed by caller through channel
		return ctx.done
	}
}

func (ctx *orDone) Err() error {
	select {
	case <-ctx.Context.Done(): // ctx is closed by caller
		return ctx.Context.Err()
	case <-ctx.done:
		return ErrClosed
	default:
	}

	return nil
}
