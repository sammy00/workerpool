package workerpool

import (
	"context"
	"fmt"
	"sync/atomic"
)

type poolAction struct {
	ctx      context.Context
	action   Action
	response chan<- error
	// #(pending jobs) submit in the same batch to Execute
	// zero-value means doneCallback should be invoked
	nPendingPeers *int32
	doneCallback  func(err ...error)
}

func (action *poolAction) Execute() {
	if err := action.action.Execute(action.ctx); nil != err {
		action.response <- err
	}

	// no more peer jobs are pending, the last should be responsible of
	// signal the end of the bounded batch processing
	if n := atomic.AddInt32(action.nPendingPeers, -1); 0 == n {
		action.doneCallback()
	}
	fmt.Println("n=", *action.nPendingPeers)
}
