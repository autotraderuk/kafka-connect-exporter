[![Build Status](https://travis-ci.org/zenreach/kafka-connect-exporter.svg?branch=master)](https://travis-ci.org/zenreach/kafka-connect-exporter) [![GoDoc](https://godoc.org/github.com/zenreach/kafka-connect-exporter?status.svg)](https://godoc.org/github.com/zenreach/kafka-connect-exporter)

# Kafka Connect Exporter

This is a service for monitoring kafka connect tasks via prometheus. It exposes a single guage that tracks the number of tasks deployed to a kafka connect cluster. The Guage has three labels associated:

- connector: The name of the connector the task belongs to.
- state: The state (RUNNING, FAILED, etc...) of the task.
- worker: The kafka connect worker (host:port) the task is deployed to.

Configuration
=============

The following environment variables can be used to configure the exporter.

| Variable                  | Description                   | Required  | Default   |
| ------------------------- | ----------------------------- | --------- | --------- |
| KAFKA\_CONNECT\_HOST      | Kafka connect host to monitor | Yes       | N/A       |
| PORT                      | Port to listen on             | No        | 9400      |       

Example
=======

For an example on how to use the exporter, see [here](example).

Docker
======

Images are hosted on docker hub:

```
docker pull zenreach/kafka-connect-exporter
```

Version tags match up with the github releases.
