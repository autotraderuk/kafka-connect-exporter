[![Build Status](https://travis-ci.org/zenreach/kafka-connect-exporter.svg?branch=master)](https://travis-ci.org/zenreach/kafka-connect-exporter)

# Kafka Connect Exporter

This is a service for monitoring kafka connect tasks via prometheus. It exposes a single guage that tracks the number of tasks deployed to a kafka connect cluster. The Guage has three labels associated:

- connector: The name of the connector the task belongs to.
- state: The state (RUNNING, FAILED, etc...) of the task.
- worker: The kafka connect worker (host:port) the task is deployed to.

# Configuration

The task monitor can be configured via environment variables, yaml file, or remote config via consul. If using one of the latter two methods, the `CONFIG_FILE_PATH` or `CONFIG_CONSUL_HOST` and `CONFIG_CONSUL_PATH` environment variables will have to be set, respectively. For all other configuration options, please see the config.yaml.example file. The corresponding environment variable names will be in all caps, with dots and hyphens replaced with underscores. For example, using the given config.yaml.example, the `logging.logentries.token` path could correspond to the `LOGGING_LOGENTRIES_TOKEN` environment variable.

# Example

For an example on how to use the task monitor, see the example directory.

# Docker

Images are hosted on docker hub:

```
docker pull zenreach/kafka-connect-exporter
```

Version tags match up with the github releases.
