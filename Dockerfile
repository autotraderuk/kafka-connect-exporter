FROM debian:9.3-slim

ADD kafka-connect-exporter /usr/bin

ENTRYPOINT ["kafka-connect-exporter"]
