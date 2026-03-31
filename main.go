package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"avrocat/version"

	"github.com/IBM/sarama"
	"github.com/genuinetools/pkg/cli"
	"github.com/linkedin/goavro/v2"
	"github.com/sirupsen/logrus"
)

var (
	broker       string
	topic        string
	schema       string
	partitions   string
	offset       string
	kafkaVersion string
	debug        bool
)

func main() {
	p := cli.NewProgram()
	p.Name = "avrocat"
	p.Description = "consumes event from kafka deserializing the avro message"
	// set gitcommit and version
	p.GitCommit = version.GITCOMMIT
	p.Version = version.VERSION

	p.FlagSet = flag.NewFlagSet("global", flag.ExitOnError)

	p.FlagSet.StringVar(&broker, "broker(s)", "", "Bootstrap broker(s) (host[:port]), comma-separated")
	p.FlagSet.StringVar(&broker, "b", "", "Bootstrap broker(s) (host[:port]), comma-separated")

	p.FlagSet.StringVar(&topic, "topic", "", "Topic to consume from")
	p.FlagSet.StringVar(&topic, "t", "", "Topic to consume from")

	p.FlagSet.StringVar(&schema, "schema", "", "Path to avro schema file")
	p.FlagSet.StringVar(&schema, "s", "", "Path to avro schema file")

	p.FlagSet.StringVar(&partitions, "partitions", "all", "The partitions to consume, can be 'all' or comma-separated numbers")
	p.FlagSet.StringVar(&partitions, "p", "all", "The partitions to consume, can be 'all' or comma-separated numbers")

	p.FlagSet.StringVar(&offset, "offset", "newest", "The offset to start with. Can be `oldest` or `newest`")
	p.FlagSet.StringVar(&offset, "o", "newest", "The offset to start with. Can be `oldest` or `newest`")

	p.FlagSet.BoolVar(&debug, "d", false, "enable debug logging")
	p.FlagSet.BoolVar(&debug, "debug", false, "enable debug logging")

	p.FlagSet.StringVar(&kafkaVersion, "V", sarama.MinVersion.String(), "Kafka version")
	p.FlagSet.StringVar(&kafkaVersion, "Version", sarama.MinVersion.String(), "Kafka version")

	// Set the before function.
	p.Before = func(ctx context.Context) error {
		// On ^C, or SIGTERM handle exit.
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)
		signal.Notify(signals, syscall.SIGTERM)
		_, cancel := context.WithCancel(ctx)
		go func() {
			for sig := range signals {
				cancel()
				logrus.Infof("Received %s, exiting.", sig.String())
				os.Exit(0)
			}
		}()

		// Set the log level.
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		} else {
			logrus.SetLevel(logrus.ErrorLevel)
		}

		return nil
	}

	p.Action = func(ctx context.Context, args []string) error {
		switch {
		case broker == "":
			return errors.New("kafka broker not defined")
		case schema == "":
			return errors.New("schema filename not defined")
		case topic == "":
			v := strings.Split(schema, "/")
			if len(v) == 0 {
				return errors.New("invalid schema path")
			}
			n := len(v) - 1
			topic = strings.TrimSuffix(v[n], ".avsc")
			if topic == "" {
				return errors.New("topic not defined")
			}
		}

		if _, err := os.Stat(schema); os.IsNotExist(err) {
			return fmt.Errorf("no suck file or directory `%s`: %w", schema, err)
		}
		buf, err := os.ReadFile(schema)
		if err != nil {
			return fmt.Errorf("reading filed %s failed: %w", schema, err)
		}
		buf, err = Transform(buf)
		if err != nil {
			return fmt.Errorf("failed to transform `%s`: %w", schema, err)
		}
		codec, err := goavro.NewCodec(string(buf))
		if err != nil {
			return fmt.Errorf("failed to load avro code: %w", err)
		}
		processor := NewProcessor(NewAvroDecoder(codec))

		brokers := parseBrokers(broker)
		consumer, err := NewKafkaConsumer(brokers, topic, partitions, Offset(offset), debug, kafkaVersion, processor)
		if err != nil {
			return fmt.Errorf("failed to initialize kafka consumer: %w", err)
		}
		err = consumer.Consume(ctx)
		if err != nil {
			return fmt.Errorf("consume %s topic and serialize %s schema failed: %w", topic, schema, err)
		}
		return nil
	}
	p.Run()
}

func parseBrokers(broker string) []string {
	broker = strings.ReplaceAll(broker, " ", "")
	if !strings.Contains(broker, ",") {
		return []string{broker}
	}
	return strings.Split(broker, ",")
}
