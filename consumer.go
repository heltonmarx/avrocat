package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/IBM/sarama"
	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
)

// Consume it is a blocking function who dispatch an incoming event to processor.
func Consume(ctx context.Context, brokers []string, topic string, partitions string, offset string, processor *Processor) error {
	var hwm int64
	switch offset {
	case "oldest":
		hwm = sarama.OffsetOldest
	case "newest":
		hwm = sarama.OffsetNewest
	default:
		return fmt.Errorf("invalid offset (%s) sould be `oldest` or `newest`", offset)
	}

	consumer, err := sarama.NewConsumer(brokers, nil)
	if err != nil {
		return err
	}

	partitionList, err := getPartitions(consumer, topic, partitions)
	if err != nil {
		return err
	}
	logrus.Debugf("connecting to %v -- partitions: %d\n", brokers, len(partitionList))
	logrus.Debugf("topic: %s -- offset: %s\n", topic, offset)

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
						continue
					}
					logrus.Debugf("incoming message: %v\n", spew.Sdump(msg))
					processor.Process(ctx, topic, msg.Value)
				case err := <-pc.Errors():
					logrus.WithError(err).Errorf("partition consumer error")
				case <-ctx.Done():
					pc.AsyncClose()
					return
				}
			}
		}(pc)
	}
	wg.Wait()
	return consumer.Close()
}

func getPartitions(consumer sarama.Consumer, topic, partitions string) ([]int32, error) {
	if partitions == "all" {
		return consumer.Partitions(topic)
	}
	v := strings.Split(partitions, ",")
	var list []int32
	for _, p := range v {
		val, err := strconv.ParseInt(p, 10, 32)
		if err != nil {
			return nil, err
		}
		list = append(list, int32(val))
	}
	return list, nil
}
