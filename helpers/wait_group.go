package helpers

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrWaitgroupNilProvided        = errors.New("Invalid wait group")
	ErrWaitgroupContextDoneNoError = errors.New("Context is in done state, no error provided is return")
)

func WaitWithContext(ctx context.Context, wg *sync.WaitGroup) error {
	if wg == nil {
		return ErrWaitgroupNilProvided
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
	}()

	select {
	case <-done:
	case <-ctx.Done():
		ctxErr := ctx.Err()
		if ctxErr != nil {
			return ctxErr
		}
		return ErrWaitgroupContextDoneNoError
	}
	return nil
}
