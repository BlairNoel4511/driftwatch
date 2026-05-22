package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// WriteText writes all counters in a simple key=value text format to w.
// Static labels are prepended as comments.
func (c *Counters) WriteText(w io.Writer) error {
	labels := c.Labels()
	if len(labels) > 0 {
		keys := make([]string, 0, len(labels))
		for k := range labels {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s=%s", k, labels[k]))
		}
		if _, err := fmt.Fprintf(w, "# labels: %s\n", strings.Join(parts, " ")); err != nil {
			return err
		}
	}

	snap := c.Snapshot()
	keys := make([]string, 0, len(snap))
	for k := range snap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if _, err := fmt.Fprintf(w, "%s %d\n", k, snap[k]); err != nil {
			return err
		}
	}
	return nil
}

// WriteJSON writes all counters as a JSON object to w, including labels.
func (c *Counters) WriteJSON(w io.Writer) error {
	type payload struct {
		Labels   map[string]string `json:"labels,omitempty"`
		Counters map[string]int64  `json:"counters"`
	}
	p := payload{
		Labels:   c.Labels(),
		Counters: c.Snapshot(),
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(p)
}
