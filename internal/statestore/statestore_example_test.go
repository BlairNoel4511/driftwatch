package statestore_test

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"driftwatch/internal/statestore"
)

func ExampleStore_Set() {
	dir, _ := os.MkdirTemp("", "statestore-example-*")
	defer os.RemoveAll(dir)

	st, _ := statestore.New(filepath.Join(dir, "state.json"))

	entry := statestore.Entry{
		Path:       "/etc/hosts",
		Size:       512,
		Mode:       0o644,
		ModTime:    time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		Checksum:   "d41d8cd98f00b204e9800998ecf8427e",
		RecordedAt: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
	}
	_ = st.Set(entry)

	got, ok := st.Get("/etc/hosts")
	fmt.Println(ok, got.Checksum)
	// Output: true d41d8cd98f00b204e9800998ecf8427e
}
