# avrocat

Consume events from kafka deserializing using avro schemas and print as a json message

## Usage

```console
Usage: avrocat <command>

Flags:

  -b, --broker      Bootstrap broker (host[:port]) (default: <none>)
  -d                enable debug logging (default: false)
  -o, --offset      The offset to start with. Can be `oldest` or `newest` (default: newest)
  -p, --partitions  The partitions to consume, can be 'all' or comma-separated numbers (default: all)
  -s, --schema      Path to avro schema file (default: <none>)
  -t, --topic       Topic to consume from (default: <none>)
  --tasr            Enable the tasr decoder (removing event header) (default: false)

Commands:

  version  Show the version information.

```

## Docker

### Build

It's not necessary install [Golang](https://golang.org/dl/) to run the script, just build the container
using [docker](https://www.docker.com/) with following command:

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

