# Bind Exporter
[![GoDoc](https://godoc.org/github.com/digitalocean/bind_exporter?status.svg)](https://godoc.org/github.com/digitalocean/bind_exporter)
[![Build Status](https://travis-ci.org/digitalocean/bind_exporter.svg)](https://travis-ci.org/digitalocean/bind_exporter)
[![Go Report Card](https://goreportcard.com/badge/digitalocean/bind_exporter)](https://goreportcard.com/report/digitalocean/bind_exporter)

Export BIND(named/dns) v9+ service metrics to Prometheus.

## Getting started

```bash
make
./bind_exporter [flags]
```

## Troubleshooting

Make sure BIND was built with libxml2 support. You can check with the following
command: `named -V | grep libxml2`.

Configure BIND to open a statistics channel. It's recommended to run the
bind_exporter next to BIND, so it's only necessary to open a port locally.

```
statistics-channels {
  inet 127.0.0.1 port 8053 allow { 127.0.0.1; };
};
```

---

Copyright @ 2016 DigitalOcean™ Inc.
