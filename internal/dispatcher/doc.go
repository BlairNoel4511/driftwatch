// Package dispatcher provides prefix-based routing of drift events to
// registered Handler functions.
//
// A Dispatcher holds an ordered list of Routes. When Dispatch is called with
// a watcher.Event, every route whose Prefix is a prefix of the event's Path
// receives the event. If no route matches and a fallback Handler was supplied
// to New, the fallback is invoked instead.
//
// Routes are matched in registration order; multiple routes may fire for a
// single event. Thread-safe for concurrent Dispatch and Register calls.
package dispatcher
