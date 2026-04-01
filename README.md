# avrocat

A CLI tool for consuming events from Kafka, deserializing them with Avro schemas, and printing the output as JSON.

## Usage

```console
Usage: avrocat <command>

Flags:

  -V, --Version     Kafka version (default: 0.8.2.0)
  -b, --broker(s)   Bootstrap broker(s) (host[:port]), comma-separated (default: <none>)
  -d, --debug       enable debug logging (default: false)
  -o, --offset      The offset to start with. Can be `oldest` or `newest` (default: newest)
  -p, --partitions  The partitions to consume, can be 'all' or comma-separated numbers (default: all)
  -s, --schema      Path to avro schema file (default: <none>)
  -t, --topic       Topic to consume from (default: <none>)

Commands:

  version  Show the version information.

```

## Docker

### Build

It's not necessary to install [Golang](https://golang.org/dl/) to run the script, just build the container
using [docker](https://www.docker.com/) with the following command:

```console
docker build -t avrocat .

```

### Running

```console
$ docker run --rm -ti --name avrocat \
    --privileged --network=host --pid=host \
    avrocat -b kafka.hmarques.dev.use1.com:9092 \
      -t <topic> \
      -s schema.avsc \
      -d

```

