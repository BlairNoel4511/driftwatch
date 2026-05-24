package retrier_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"driftwatch/internal/retrier"
)

// ExampleRetrier_Do demonstrates retrying a fallible operation.
func ExampleRetrier_Do() {
	r := retrier.New(retrier.Config{
		MaxAttempts: 3,
		BaseDelay:   time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
	})

	attempts := 0
	err := r.Do(context.Background(), func(_ context.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("not ready yet")
		}
		return nil
	})

	fmt.Println(err)        // <nil>
	fmt.Println(attempts)   // 3
	// Output:
	// <nil>
	// 3
}
