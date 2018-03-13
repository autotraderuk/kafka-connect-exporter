# Kafka Connect Exporter Example

To run this example, you will need the following installed:

- go 1.9 (some older versions may work)
- docker
- docker-compose
- cURL (Any http client should work, but examples will be given using cURL)

This example will set up a local docker-compose cluster consisting of:

- A single node kafka cluster
- A single node zookeeper cluster
- A single kafka connect worker
- A single-member mongo replica set
- The kafka connect exporter
- A prometheus server

This example uses sink and source connectors for mongodb.

First, build the exporter executable:

```
cd /path/to/repo/root && go build .
```

Next, change to the example directory and bring up the docker compose cluster.

```
docker compose up -d
```

It may take a few seconds for the kafka broker to register and come online. If you open `localhost:9090/targets` and you see a single target with an `UP` status, you're good to go.

Configure mongo:

```
docker run -it --rm --network example_default mongo:3.2 mongo mongo:27017 --eval 'rs.initiate({_id: "rs0", members: [{_id: 0, host: "mongo:27017"}]})'
```

Configure the connectors:

```
curl -XPUT "http://localhost:8083/connectors/foo-source/config" -H "Content-Type: application/json" -H "Accept: application/json" --data-binary @source-config.json
curl -XPUT "http://localhost:8083/connectors/foo-sink/config" -H "Content-Type: application/json" -H "Accept: application/json" --data-binary @sink-config.json
```

It may take a moment for the connector to launch its task, but you should be able to go to `localhost:9090/graph` in a browser, and see a value of 1 for the `kafka_connect_tasks` metric.

_BONUS:_ 

You can add some data to mongo and see it get copied by the connectors:

```
docker run -it --rm --net example_default mongo mongo:27017/foo
db.bar.insert({foo: "bar"})
db.bar_copy.find()
quit()
```
