// Package eventbus implements a lightweight publish/subscribe bus used to
// decouple drift-detection components from alert and reporting consumers.
//
// Producers call Publish with an Event; consumers register via Subscribe.
// Each subscription receives its own copy of the handler slice at publish
// time, so subscribing or unsubscribing during dispatch is safe.
package eventbus
