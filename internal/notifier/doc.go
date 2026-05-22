// Package notifier provides a fan-out dispatcher that receives drift events
// from a watcher.Poller channel and forwards them to one or more Sink
// implementations (e.g. alert.Alert, webhook, PagerDuty).
//
// Basic usage:
//
//	h := history.New(100)
//	a, _ := alert.New(alert.Config{})
//	sink := notifier.NewAlertSink(a)
//	n := notifier.New(h, nil, sink)
//	go n.Run(ctx, driftCh)
package notifier
