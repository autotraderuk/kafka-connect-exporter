# Prometheus

This package implements a prometheus metrics backend for monitoring kafka connect tasks. It exposes a single guage that tracks the number of tasks deployed to a kafka connect cluster. The Guage has three labels associated:

- connector: The name of the connector the task belongs to.
- state: The state (RUNNING, FAILED, etc...) of the task.
- worker: The kafka connect worker (host:port) the task is deployed to.
