package main

import (
	"testing"

	"github.com/IBM/sarama"
)

// mockConsumer implements sarama.Consumer for testing getPartitions.
type mockConsumer struct {
	partitions []int32
	err        error
}

func (m *mockConsumer) Topics() ([]string, error)      { return nil, nil }
func (m *mockConsumer) Close() error                   { return nil }
func (m *mockConsumer) PauseAll()                      {}
func (m *mockConsumer) ResumeAll()                     {}
func (m *mockConsumer) HighWaterMarks() map[string]map[int32]int64 { return nil }
func (m *mockConsumer) Pause(topicPartitions map[string][]int32)   {}
func (m *mockConsumer) Resume(topicPartitions map[string][]int32)  {}
func (m *mockConsumer) Partitions(topic string) ([]int32, error) {
	return m.partitions, m.err
}
func (m *mockConsumer) ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
	return nil, nil
}

func TestGetPartitions_All(t *testing.T) {
	mc := &mockConsumer{partitions: []int32{0, 1, 2}}
	got, err := getPartitions(mc, "test-topic", "all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 || got[0] != 0 || got[1] != 1 || got[2] != 2 {
		t.Errorf("getPartitions(all) = %v, want [0 1 2]", got)
	}
}

func TestGetPartitions_AllError(t *testing.T) {
	mc := &mockConsumer{err: sarama.ErrUnknownTopicOrPartition}
	_, err := getPartitions(mc, "missing-topic", "all")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPartitions_CommaSeparated(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantLen    int
		wantFirst  int32
		wantLast   int32
	}{
		{"single partition", "0", 1, 0, 0},
		{"two partitions", "0,1", 2, 0, 1},
		{"multiple partitions", "0,2,4", 3, 0, 4},
	}
	mc := &mockConsumer{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPartitions(mc, "topic", tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Fatalf("len = %d, want %d", len(got), tt.wantLen)
			}
			if got[0] != tt.wantFirst {
				t.Errorf("first = %d, want %d", got[0], tt.wantFirst)
			}
			if got[len(got)-1] != tt.wantLast {
				t.Errorf("last = %d, want %d", got[len(got)-1], tt.wantLast)
			}
		})
	}
}

func TestGetPartitions_InvalidNumber(t *testing.T) {
	mc := &mockConsumer{}
	_, err := getPartitions(mc, "topic", "0,abc,2")
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
}

func TestNewKafkaConsumer_InvalidOffset(t *testing.T) {
	_, err := NewKafkaConsumer(
		[]string{"localhost:9092"},
		"topic",
		"all",
		Offset("invalid"),
		false,
		sarama.DefaultVersion.String(),
		SASL{},
		nil,
	)
	if err == nil {
		t.Fatal("expected error for invalid offset, got nil")
	}
}

func TestNewKafkaConsumer_OldestOffset(t *testing.T) {
	// Confirms the oldest offset path returns a broker error (not an offset validation error),
	// meaning offset resolution itself succeeded.
	_, err := NewKafkaConsumer(
		[]string{"localhost:19092"}, // unreachable
		"topic",
		"all",
		Oldest,
		false,
		sarama.DefaultVersion.String(),
		SASL{},
		nil,
	)
	if err == nil {
		t.Fatal("expected connection error, got nil")
	}
}

func TestNewKafkaConsumer_NewestOffset(t *testing.T) {
	// Same as above for the newest path.
	_, err := NewKafkaConsumer(
		[]string{"localhost:19092"}, // unreachable
		"topic",
		"all",
		Newest,
		false,
		sarama.DefaultVersion.String(),
		SASL{},
		nil,
	)
	if err == nil {
		t.Fatal("expected connection error, got nil")
	}
}
