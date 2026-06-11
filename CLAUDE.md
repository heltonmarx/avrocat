# avrocat

CLI tool that consumes Kafka messages and deserializes them from Avro binary format to colored JSON, using a local `.avsc` schema file.

## Build & test

```bash
make          # test + build
make build    # build only (produces ./avrocat binary)
make test     # go test -v -cover ./...
make clean    # go clean
```

The binary embeds version and git commit via ldflags; always build through `make` or the CI workflow to get correct version metadata.

## Running

```bash
./avrocat -b localhost:9092 -s path/to/schema.avsc -t my-topic
```

Topic can be omitted: it is derived automatically from the schema filename (strips the `.avsc` extension and the leading path).

Flag reference:
- `-b` / `--broker(s)` — Bootstrap broker(s), comma-separated
- `-t` / `--topic` — Kafka topic (optional, inferred from schema filename)
- `-s` / `--schema` — Path to `.avsc` file
- `-p` / `--partitions` — `all` (default) or comma-separated partition numbers
- `-o` / `--offset` — `newest` (default) or `oldest`
- `-V` / `--Version` — Kafka protocol version (default: `0.8.2.0`)
- `-d` / `--debug` — Enable debug logging (routes sarama logs to logrus)
- `--strip-schema-id` — Strip the 5-byte Confluent Schema Registry header (magic byte + schema ID) before decoding (default: `true`); set to `false` for raw Avro messages
- `-S` / `--sasl` — Enable SASL authentication
- `-U` / `--username` — SASL username
- `-P` / `--password` — SASL password
- `-M` / `--mechanism` — SASL mechanism: `PLAIN` (default), `SCRAM-SHA-256`, `SCRAM-SHA-512`, `OAUTHBEARER`, `GSSAPI`

## Architecture

| File | Responsibility |
|------|---------------|
| `main.go` | CLI setup (flags, signal handling), `validate()`, `parseBrokers()` |
| `consumer.go` | `KafkaConsumer`: wraps sarama, consumes partitions concurrently |
| `sasl.go` | `SASL` struct, `validate()`, `ParseSASLMechanism()` |
| `transformer.go` | `Transform()`: flattens multi-schema Avro arrays into a single root schema |
| `decoder.go` | `AvroDecoder`: binary → textual JSON via goavro |
| `processor.go` | `Processor`: decodes + colorizes output via colorjson |

**Data flow:** Kafka binary message → `AvroDecoder.Decode()` → `Processor.format()` → colored JSON printed to stdout.

**Schema handling:** `Transform()` accepts both single-schema objects and JSON arrays of schemas. When given an array, it resolves cross-schema references and returns the root record (identified by namespace suffix matching the name). This allows multi-record `.avsc` files to be passed directly.

## Known issues / open bugs

- **`-s` flag collision**: `-s` is registered for both `--schema` and `--sasl`, causing a startup panic. Renaming one of them is required before SASL flags can be used.
- **SCRAM requires `SCRAMClientGeneratorFunc`**: sarama's `Config.Validate()` rejects SCRAM-SHA-256/512 if `Net.SASL.SCRAMClientGeneratorFunc` is nil. The current code never sets it, so SCRAM auth cannot be used yet.
- **`ParseSASLMechanism` silent fallback**: invalid mechanism strings silently become `PLAIN`; `SASL.validate()` does not reject unknown mechanism values.
- **Context cancel not wired to consumer**: the `cancel` func in `p.Before` is called on an orphaned child context; `consumer.Consume` receives the original context, so `ctx.Done()` is never triggered by the signal handler (process exits via `os.Exit` instead).
- **`Consume()` references global `topic`**: `KafkaConsumer` stores no `topic` field; the goroutines read the package-level `topic` variable directly.

## Tests

Tests live in `transformer_test.go` and cover `Transform`, `isJSONArray`, and `trimSuffix`. Run with `make test` or `go test -v -cover ./...`.

There are no integration tests; a running Kafka broker is required for manual end-to-end testing.
