# Bind Exporter
[![GoDoc](https://godoc.org/github.com/prometheus-community/bind_exporter?status.svg)](https://godoc.org/github.com/prometheus-community/bind_exporter)
[![Build Status](https://circleci.com/gh/prometheus-community/bind_exporter.svg?style=svg)](https://circleci.com/gh/prometheus-community/bind_exporter)
[![Go Report Card](https://goreportcard.com/badge/prometheus-community/bind_exporter)](https://goreportcard.com/report/prometheus-community/bind_exporter)

Export BIND(named/dns) v9+ service metrics to Prometheus.

## Getting started

```bash
go get github.com/prometheus-community/bind_exporter
cd $GOPATH/src/github.com/prometheus-community/bind_exporter
make
./bind_exporter [flags]
```

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
