package backoff

import (
	"fmt"
	"time"
)

// PermanentError signals that the operation should not be retried.
type PermanentError struct {
	Err error
}

// Permanent wraps the given err in a *PermanentError.
func Permanent(err error) error {
	if err == nil {
		return nil
	}
	return &PermanentError{
		Err: err,
	}
}

func (e *PermanentError) Error() string {
	return e.Err.Error()
}

func (e *PermanentError) Unwrap() error {
	return e.Err
}

type RetryAfterError struct {
	Duration time.Duration
}

func RetryAfter(seconds int) error {
	return &RetryAfterError{Duration: time.Duration(seconds) * time.Second}
}

func (e *RetryAfterError) Error() string {
	return fmt.Sprintf("retry after %s", e.Duration)
}
