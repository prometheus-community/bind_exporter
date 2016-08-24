# Bind Exporter

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

## Performance

The default on bind before 9.10 is to dump the full socket / task queue. This
is an expensive operation that can freeze your server
- On bind 9.8 expect a stall of a few milliseconds
- On bind 9.9 expect a stall of 10+ms up to >100ms
- On bind 9.10 only the server stats are loaded by default

*Try to avoid bind 9.9 with this exporter*.
