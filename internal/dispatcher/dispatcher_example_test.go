package dispatcher_test

import (
	"fmt"

	"github.com/yourorg/driftwatch/internal/dispatcher"
	"github.com/yourorg/driftwatch/internal/watcher"
)

func ExampleDispatcher_Dispatch() {
	d := dispatcher.New(func(e watcher.Event) {
		fmt.Println("fallback:", e.Path)
	})

	d.Register("/etc", func(e watcher.Event) {
		fmt.Println("etc handler:", e.Path)
	})

	d.Dispatch(watcher.Event{Path: "/etc/nginx/nginx.conf"})
	d.Dispatch(watcher.Event{Path: "/tmp/scratch"})

	// Output:
	// etc handler: /etc/nginx/nginx.conf
	// fallback: /tmp/scratch
}
