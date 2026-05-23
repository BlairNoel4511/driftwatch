// Package deadletter implements a bounded dead-letter queue for drift events
// that could not be processed after all retry attempts have been exhausted.
//
// Events are stored in insertion order. When the queue reaches capacity the
// oldest entry is evicted automatically (FIFO). All operations are safe for
// concurrent use.
package deadletter
