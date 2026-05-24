// Package correlator groups related drift events by a caller-supplied key
// function, making it easy to surface clusters of related infrastructure
// changes — for example all files under the same service directory or owned
// by the same team — rather than surfacing every file change in isolation.
//
// Typical usage:
//
//	c := correlator.New(func(e watcher.Event) string {
//		return filepath.Dir(e.Path)
//	})
//	c.Record(event)
//	for _, g := range c.Groups() {
//		fmt.Println(g.Key, len(g.Events))
//	}
package correlator
