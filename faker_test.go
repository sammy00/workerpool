package workerpool_test

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/sammyne/workerpool"
)

var errFailure = errors.New("failure")

type JobSpy struct {
	Busy       time.Duration
	ErrOut     bool
	Done, Wait chan struct{}
	Progress   *int32
}

func (spy *JobSpy) Execute(ctx context.Context) error {
	if nil != spy.Wait {
		<-spy.Wait
	}

	defer func() {
		if nil != spy.Done {
			close(spy.Done)
		}
	}()

	select {
	case <-ctx.Done():
	case <-time.After(spy.Busy):
		atomic.AddInt32(spy.Progress, 1)
	}

	var err error
	if spy.ErrOut {
		err = errFailure
	}

	return err
}

func MockErrPinepline(progress *int32) []workerpool.Action {
	starter := &JobSpy{
		Done:     make(chan struct{}, 1),
		Progress: progress,
	}

	errJob := &JobSpy{
		Busy:     time.Second * 2,
		Done:     make(chan struct{}, 1),
		ErrOut:   true,
		Progress: progress,
		Wait:     starter.Done,
	}

	okJob := &JobSpy{
		Busy:     time.Minute,
		Done:     make(chan struct{}, 1),
		Progress: progress,
		Wait:     errJob.Done,
	}

	end := &JobSpy{
		Busy:     time.Minute,
		Progress: progress,
		Wait:     okJob.Done,
	}

	return []workerpool.Action{starter, errJob, okJob, end}
}

func MockErrFanOut(progress *int32) []workerpool.Action {
	starter := &JobSpy{
		Done:     make(chan struct{}, 1),
		Progress: progress,
	}

	errJob := &JobSpy{
		Done:     make(chan struct{}, 1),
		ErrOut:   true,
		Progress: progress,
		Wait:     starter.Done,
	}

	okJob := &JobSpy{
		Busy:     time.Second,
		Done:     make(chan struct{}, 1),
		Progress: progress,
		Wait:     starter.Done,
	}

	end := &JobSpy{
		Busy:     time.Second,
		Progress: progress,
		Wait:     starter.Done,
	}

	return []workerpool.Action{starter, errJob, okJob, end}
}
