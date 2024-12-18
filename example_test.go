package backoff

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

func ExampleRetry() {
	// Define an operation function that returns a value and an error.
	// The value can be any type.
	// We'll pass this operation to Retry function.
	operation := func() (string, error) {
		// An example request that may fail.
		resp, err := http.Get("http://httpbin.org/get")
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		// In case on non-retriable error, return Permanent error to stop retrying.
		// For this HTTP example, client errors are non-retriable.
		if resp.StatusCode == 400 {
			return "", Permanent(errors.New("bad request"))
		}

		// If we are being rate limited, return a RetryAfter to specify how long to wait.
		// This will also reset the backoff policy.
		if resp.StatusCode == 429 {
			seconds, err := strconv.ParseInt(resp.Header.Get("Retry-After"), 10, 64)
			if err == nil {
				return "", RetryAfter(int(seconds))
			}
		}

		// Return successful response.
		return "hello", nil
	}

	result, err := Retry(context.TODO(), operation, WithBackOff(NewExponentialBackOff()))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Operation is successful.

	fmt.Println(result)
	// Output: hello
}
