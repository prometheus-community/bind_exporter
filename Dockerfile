FROM        quay.io/prometheus/busybox:latest
MAINTAINER  DigitalOcean Engineering <>

COPY bind_exporter /bin/bind_exporter

EXPOSE      9119
ENTRYPOINT  [ "/bin/bind_exporter" ]
