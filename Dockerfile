FROM debian:jessie

ADD kafka-connect-exporter /usr/bin

ENTRYPOINT ["kafka-connect-exporter"]
