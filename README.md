# driftwatch

A daemon that monitors infrastructure config files and alerts when live state diverges from declared state.

---

## Installation

```bash
go install github.com/yourusername/driftwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/driftwatch.git && cd driftwatch && go build -o driftwatch .
```

---

## Usage

Create a `driftwatch.yaml` config file pointing to your infrastructure declarations:

```yaml
watch_interval: 60s
alert_channel: slack
configs:
  - path: ./infra/network.tf
    provider: terraform
  - path: ./k8s/deployment.yaml
    provider: kubernetes
```

Then start the daemon:

```bash
driftwatch --config driftwatch.yaml
```

When live state diverges from declared state, driftwatch will emit an alert with a diff summary:

```
[DRIFT DETECTED] k8s/deployment.yaml
  expected replicas: 3
  live replicas:     1
  detected at: 2024-11-02T14:32:10Z
```

Run `driftwatch --help` for a full list of flags and options.

---

## Configuration Options

| Field            | Description                          | Default |
|------------------|--------------------------------------|---------|
| `watch_interval` | How often to poll live state         | `30s`   |
| `alert_channel`  | Alert destination (`slack`, `pagerduty`, `log`) | `log` |
| `configs`        | List of config files to monitor      | —       |

---

## License

MIT © 2024 driftwatch contributors