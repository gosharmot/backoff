package backoff

import (
	"context"
	"fmt"
	"log"
)

func ExampleRetry() {
	// An operation that may fail.
	operation := func() (string, error) {
		return "hello", nil
	}

	val, err := Retry(context.TODO(), operation, WithBackOff(NewExponentialBackOff()))
	if err != nil {
		// Handle error.
		return
	}

	// Operation is successful.

	fmt.Println(val)
	// Output: hello
}

func ExampleTicker() {
	// An operation that may fail.
	operation := func() error {
		return nil // or an error
	}

	ticker := NewTicker(NewExponentialBackOff())

	var err error

	// Ticks will continue to arrive when the previous operation is still running,
	// so operations that take a while to fail could run in quick succession.
	for range ticker.C {
		if err = operation(); err != nil {
			log.Println(err, "will retry...")
			continue
		}

		ticker.Stop()
		break
	}

	if err != nil {
		// Operation has failed.
		return
	}

	// Operation is successful.
}
