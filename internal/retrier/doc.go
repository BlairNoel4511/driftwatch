// Package retrier implements retry logic with exponential backoff for use
// throughout driftwatch wherever transient failures may occur — e.g. when
// reading a monitored file, delivering an alert, or persisting a checkpoint.
//
// Usage:
//
//	r := retrier.New(retrier.Config{
//		MaxAttempts: 5,
//		BaseDelay:   100 * time.Millisecond,
//		MaxDelay:    5 * time.Second,
//	})
//	err := r.Do(ctx, func(ctx context.Context) error {
//		return doSomethingFallible(ctx)
//	})
package retrier
