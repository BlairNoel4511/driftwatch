package filter_test

import (
	"testing"

	"github.com/yourusername/driftwatch/internal/filter"
)

func TestAllow_NoPatterns_AllowsAll(t *testing.T) {
	f := filter.New(nil, nil)
	for _, path := range []string{"/etc/hosts", "/tmp/foo.txt", "bar.yaml"} {
		ok, err := f.Allow(path)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", path, err)
		}
		if !ok {
			t.Errorf("expected %q to be allowed with no patterns", path)
		}
	}
}

func TestAllow_IncludeGlob(t *testing.T) {
	f := filter.New([]string{"*.yaml"}, nil)

	allowed, err := f.Allow("/etc/driftwatch.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Error("expected .yaml file to be allowed")
	}

	denied, err := f.Allow("/etc/hosts")
	if err != nil {
		t.Fatal(err)
	}
	if denied {
		t.Error("expected non-.yaml file to be denied")
	}
}

func TestAllow_ExcludeGlob(t *testing.T) {
	f := filter.New(nil, []string{"*.tmp"})

	allowed, err := f.Allow("/var/run/app.conf")
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Error("expected .conf file to be allowed")
	}

	denied, err := f.Allow("/var/run/scratch.tmp")
	if err != nil {
		t.Fatal(err)
	}
	if denied {
		t.Error("expected .tmp file to be excluded")
	}
}

func TestAllow_ExcludeTakesPrecedence(t *testing.T) {
	// A path that matches both include and exclude should be denied.
	f := filter.New([]string{"*.yaml"}, []string{"secret*.yaml"})

	ok, err := f.Allow("secret_keys.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("exclude should take precedence over include")
	}

	ok, err = f.Allow("config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("non-excluded .yaml file should be allowed")
	}
}

func TestAllow_InvalidPattern_ReturnsError(t *testing.T) {
	f := filter.New([]string{"[invalid"}, nil)
	_, err := f.Allow("anything.yaml")
	if err == nil {
		t.Error("expected error for invalid glob pattern")
	}
}
