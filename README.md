# Bind Exporter
[![GoDoc](https://godoc.org/github.com/prometheus-community/bind_exporter?status.svg)](https://godoc.org/github.com/prometheus-community/bind_exporter)
[![Build Status](https://github.com/prometheus-community/bind_exporter/actions/workflows/ci.yml/badge.svg)](https://github.com/prometheus-community/bind_exporter/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/prometheus-community/bind_exporter)](https://goreportcard.com/report/prometheus-community/bind_exporter)

Export BIND (named/dns) v9+ service metrics to Prometheus.

## Getting started

### Build and run from source
```bash
git clone https://github.com/prometheus-community/bind_exporter.git
cd bind_exporter
make build
./bind_exporter [flags]
```

Pre-built binaries are available on the [releases page](https://github.com/prometheus-community/bind_exporter/releases).

### Run in Docker container

1. Pull Docker container using a specific version:
```bash
docker pull prometheuscommunity/bind-exporter:v0.3.0
```
2. Run in a Docker container (as daemon), use `--network host` when communicating with `named` via `localhost`:
```bash
docker run -d --network host prometheuscommunity/bind-exporter:v0.3.0
```

### Examples

Run `bind_exporter` in a Docker container and communicate with `named` on non-default statistics URL:
```bash
docker run -d prometheuscommunity/bind-exporter:v0.3.0 --bind.stats-url http://<IP/hostname>:8053
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--bind.stats-url` | `http://localhost:8053/` | Address of the BIND statistics channel. |
| `--bind.stats-version` | `json` | Statistics channel API to use: `json`, or `xml` (`xml.v3` is an alias). `auto` is accepted and currently uses the JSON API. |
| `--bind.stats-groups` | `server,view` | Comma-separated list of statistics groups to collect: `server`, `view`, `tasks`. |
| `--bind.timeout` | `10s` | Timeout for requests to the statistics channel. |
| `--bind.pid-file` | `/run/named/named.pid` | Path to BIND's pid file, used to export `process_*` metrics for `named`. |
| `--web.listen-address` | `:9119` | Address to expose metrics on. Repeatable. |
| `--web.telemetry-path` | `/metrics` | Path under which to expose metrics. |
| `--web.config.file` | | Path to a TLS/basic auth configuration file (see below). |
| `--web.systemd-socket` | `false` | Use systemd socket activation listeners instead of port listeners (Linux only). |
| `--log.level` | `info` | One of `debug`, `info`, `warn`, `error`. |
| `--log.format` | `logfmt` | One of `logfmt`, `json`. |

Run `./bind_exporter --help` for the full list.

## TLS and basic authentication

The Bind Exporter supports TLS and basic authentication.

To use TLS and/or basic authentication, you need to pass a configuration file
using the `--web.config.file` parameter. The format of the file is described
[in the exporter-toolkit repository](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md).

## Other resources

[Grafana Dashboard #12309](https://grafana.com/grafana/dashboards/12309)

## Troubleshooting

The default `--bind.stats-version=json` needs BIND built with json-c support;
`--bind.stats-version=xml` needs libxml2 support. You can check with the
following command: `named -V | grep -E 'json-c|libxml2'`.

Configure BIND to open a statistics channel. It's recommended to run the
bind\_exporter next to BIND, so it's only necessary to open a port locally.

```
statistics-channels {
  inet 127.0.0.1 port 8053 allow { 127.0.0.1; };
};
```

---

Copyright @ 2016 DigitalOcean™ Inc.
