package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

type Offset string

const (
	Oldest Offset = "oldest"
	Newest Offset = "newest"
)

// Consume it is a blocking function who dispatch an incoming event to processor.
func Consume(ctx context.Context,
	brokers []string,
	topic string,
	partitions string,
	offset Offset,
	debug bool,
	kafkaVersion string,
	processor *Processor,
) (err error) {
	var hwm int64
	switch offset {
	case Oldest:
		hwm = sarama.OffsetOldest
	case Newest:
		hwm = sarama.OffsetNewest
	default:
		return fmt.Errorf("invalid offset (%s) sould be `oldest` or `newest`", offset)
	}
	if debug {
		sarama.Logger = logrus.StandardLogger()
	}
	config := sarama.NewConfig()

	config.Version, err = sarama.ParseKafkaVersion(kafkaVersion)
	if err != nil {
		config.Version = sarama.MinVersion
	}

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return err
	}
	defer consumer.Close()

	partitionList, err := getPartitions(consumer, topic, partitions)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"brokers":    brokers,
		"topic":      topic,
		"partitions": partitionList,
		"offset":     offset,
	}).Debugf("starting consumer")

	var wg sync.WaitGroup
	for _, partition := range partitionList {
		pc, err := consumer.ConsumePartition(topic, partition, hwm)
		if err != nil {
			return err
		}
		wg.Add(1)
		go func(pc sarama.PartitionConsumer) {
			defer wg.Done()
			for {
				select {
				case msg, ok := <-pc.Messages():
					if !ok {
						return
					}
					out, err := processor.Process(ctx, topic, msg.Value)
					if err != nil {
						logrus.WithError(err).Error("failed to process incoming message")
						continue
					}
					logrus.Infoln(out)
				case err := <-pc.Errors():
					logrus.WithError(err).Error("partition consumer error")
				case <-ctx.Done():
					pc.AsyncClose()
					return
				}
			}
		}(pc)
	}
	wg.Wait()
	return nil
}

// getPartitions returns a slice of partition IDs for the given topic.
// If partitions is "all", it returns all partitions for the topic.
// Otherwise, it parses a comma-separated list of partition numbers.
// Parameters:
//
//	consumer: sarama.Consumer instance used to fetch partition information.
//	topic: Kafka topic name.
//	partitions: "all" or comma-separated partition numbers.
//
// Returns:
//
//	[]int32: slice of partition IDs.
//	error: error if parsing fails or partitions cannot be retrieved.
func getPartitions(consumer sarama.Consumer, topic, partitions string) ([]int32, error) {
	if partitions == "all" {
		return consumer.Partitions(topic)
	}
	v := strings.Split(partitions, ",")

	list := make([]int32, 0, len(v))
	for _, p := range v {
		val, err := strconv.ParseInt(p, 10, 32)
		if err != nil {
			return nil, err
		}
		list = append(list, int32(val))
	}
	return list, nil
}
