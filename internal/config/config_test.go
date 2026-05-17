package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "driftwatch.yaml")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}
	return p
}

func TestLoad_Valid(t *testing.T) {
	raw := `
poll_interval: 10s
log_level: debug
targets:
  - name: nginx
    path: /etc/nginx/nginx.conf
alerting:
  webhook_url: https://hooks.example.com/alert
`
	p := writeTemp(t, raw)
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != 10*time.Second {
		t.Errorf("poll_interval: got %v, want 10s", cfg.PollInterval)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("log_level: got %q, want debug", cfg.LogLevel)
	}
	if len(cfg.Targets) != 1 || cfg.Targets[0].Name != "nginx" {
		t.Errorf("unexpected targets: %+v", cfg.Targets)
	}
}

func TestLoad_Defaults(t *testing.T) {
	raw := `
targets:
  - name: sshd
    path: /etc/ssh/sshd_config
`
	p := writeTemp(t, raw)
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != 30*time.Second {
		t.Errorf("default poll_interval: got %v, want 30s", cfg.PollInterval)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("default log_level: got %q, want info", cfg.LogLevel)
	}
}

func TestLoad_NoTargets(t *testing.T) {
	raw := `log_level: info\n`
	p := writeTemp(t, raw)
	_, err := Load(p)
	if err == nil {
		t.Fatal("expected error for missing targets, got nil")
	}
}

func TestLoad_MissingPath(t *testing.T) {
	_, err := Load("/nonexistent/path/driftwatch.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_TargetMissingName(t *testing.T) {
	raw := `
targets:
  - path: /etc/hosts
`
	p := writeTemp(t, raw)
	_, err := Load(p)
	if err == nil {
		t.Fatal("expected validation error for missing target name")
	}
}
