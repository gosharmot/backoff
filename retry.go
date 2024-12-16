package backoff

import (
	"context"
	"errors"
	"time"
)

// Retry function stops retrying if the total time exceeds the maximum elapsed time.
const DefaultMaxElapsedTime = 15 * time.Minute

// An Operation is a function that is to be retried.
type Operation[T any] func() (T, error)

// Notify is a notify-on-error function. It receives an operation error and
// backoff delay if the operation failed (with an error).
//
// NOTE that if the backoff policy stated to stop retrying,
// the notify function isn't called.
type Notify func(error, time.Duration)

type retryOptions struct {
	BackOff        BackOff
	Timer          Timer
	Notify         Notify
	MaxElapsedTime time.Duration
	MaxTries       uint
}

type RetryOption func(*retryOptions)

func WithBackOff(b BackOff) RetryOption {
	return func(args *retryOptions) {
		args.BackOff = b
	}
}

func WithTimer(t Timer) RetryOption {
	return func(args *retryOptions) {
		args.Timer = t
	}
}

func WithNotify(n Notify) RetryOption {
	return func(args *retryOptions) {
		args.Notify = n
	}
}

// WithMaxElapsedTime sets the maximum total time for retries.
func WithMaxElapsedTime(d time.Duration) RetryOption {
	return func(args *retryOptions) {
		args.MaxElapsedTime = d
	}
}

func WithMaxTries(n uint) RetryOption {
	return func(args *retryOptions) {
		args.MaxTries = n
	}
}

// Retry the operation o until it does not return error or BackOff stops.
// o is guaranteed to be run at least once.
//
// If o returns a *PermanentError, the operation is not retried, and the
// wrapped error is returned.
//
// Retry sleeps the goroutine for the duration returned by BackOff after a
// failed operation returns.
func Retry[T any](ctx context.Context, operation Operation[T], opts ...RetryOption) (T, error) {
	// Default options
	args := &retryOptions{
		BackOff:        NewExponentialBackOff(),
		Timer:          &defaultTimer{},
		MaxElapsedTime: DefaultMaxElapsedTime,
	}

	for _, opt := range opts {
		opt(args)
	}

	defer args.Timer.Stop()

	startedAt := time.Now()
	args.BackOff.Reset()
	for numTries := uint(1); ; numTries++ {
		res, err := operation()
		if err == nil {
			return res, nil
		}

		if args.MaxTries > 0 && numTries >= args.MaxTries {
			return res, err
		}

		if time.Since(startedAt) > args.MaxElapsedTime {
			return res, err
		}

		var permanent *PermanentError
		if errors.As(err, &permanent) {
			return res, err
		}

		next := args.BackOff.NextBackOff()
		if next == Stop {
			return res, err
		}

		if cerr := ctx.Err(); cerr != nil {
			return res, cerr
		}

		if args.Notify != nil {
			args.Notify(err, next)
		}

		args.Timer.Start(next)
		select {
		case <-args.Timer.C():
		case <-ctx.Done():
			return res, ctx.Err()
		}
	}
}
