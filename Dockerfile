FROM debian:jessie

RUN mkdir -p /opt/bin

ADD kafka-connect-exporter /opt/bin

ENTRYPOINT ["/opt/bin/kafka-connect-exporter"]
