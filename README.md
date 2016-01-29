# Bind Exporter

Export BIND(named/dns) v9+ service metrics to Prometheus.

## Getting started

```bash
make
./bind_exporter [flags]
```

## Troubeshooting

Make sure BIND was built with libxml2 support. You can check with the following
command: `named -V | grep libxml2`.

Configure BIND to open a statistics channel. It's recommended to run the
bind_exporter next to BIND, so it's only necessary to open a port locally.

```
statistics-channels {
  inet 127.0.0.1 port 8080 allow { 127.0.0.1; };
};
```
