# Kafka Connect Exporter Example

To run this example, you will need the following installed:

- go 1.9 (some older versions may work)
- docker
- docker-compose
- cURL (Any http client should work, but examples will be given using cURL)
- psql (client that comes bundled with postgresql install on most systems)

This example will set up a local docker-compose cluster consisting of:

- A single node kafka cluster
- A single node zookeeper cluster
- A single kafka connect worker
- A postgresql server
- The kafka connect exporter
- A prometheus server

This example uses the jdbc connector that comes bundled with kafka connect to push sql table updates to kafka.

First, build the exporter executable:

```
cd /path/to/repo/root && go build .

```

Next, change to the example directory and bring up the docker compose cluster.

```
docker compose up -d
```

It may take a few seconds for the kafka broker to register and come online. If you open `localhost:9090/targets` and you see a single target with an `UP` status, you're good to go.

Now we need to create the SQL table:

```
psql --host localhost -U postgres
postgres=# CREATE DATABASE connect_test;
postgres=# \connect connect_test;
postgres=# CREATE TABLE people(id SERIAL PRIMARY KEY, name VARCHAR(255));
```

Then, configure the connector:

```
curl -XPUT http://localhost:8083/connectors/test-connector/config -H "Content-Type: application/json" -H "Accept: application/json" --data-binary @connector-config.json
```

It may take a moment for the connector to launch its task, but you should be able to go to `localhost:9090/graph` in a browser, and see a value of 1 for the `kafka_connect_tasks` metric.
