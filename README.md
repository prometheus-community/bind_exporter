# Bind Exporter
[![GoDoc](https://godoc.org/github.com/prometheus-community/bind_exporter?status.svg)](https://godoc.org/github.com/prometheus-community/bind_exporter)
[![Build Status](https://circleci.com/gh/prometheus-community/bind_exporter.svg?style=svg)](https://circleci.com/gh/prometheus-community/bind_exporter)
[![Go Report Card](https://goreportcard.com/badge/prometheus-community/bind_exporter)](https://goreportcard.com/report/prometheus-community/bind_exporter)

Export BIND (named/dns) v9+ service metrics to Prometheus.

## Getting started

### Build and run from source
```bash
go get github.com/prometheus-community/bind_exporter
cd $GOPATH/src/github.com/prometheus-community/bind_exporter
make
./bind_exporter [flags]
```

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
docker run -d prometheuscommunity/bind-exporter:v0.3.0 -bind.stats-url http://<IP/hostname>:8053
```

## Other resources

Grafana Dashboard: https://grafana.com/dashboards/1666

## Troubleshooting

Make sure BIND was built with libxml2 support. You can check with the following
command: `named -V | grep libxml2`.

Configure BIND to open a statistics channel. It's recommended to run the
bind\_exporter next to BIND, so it's only necessary to open a port locally.

```
statistics-channels {
  inet 127.0.0.1 port 8053 allow { 127.0.0.1; };
};
```

---

Copyright @ 2016 DigitalOcean™ Inc.
