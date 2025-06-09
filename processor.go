package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/TylerBrock/colorjson"
	"github.com/sirupsen/logrus"
)

// ProcessorOption allows to configure various aspects of Processor
type ProcessorOption func(*Processor)

// WithTasr sets the tasr decoder (enable/disable)
func WithTasr(tasr bool) ProcessorOption {
	return func(p *Processor) {
		p.tasr = tasr
	}
}

// Processor holds the avro decoder.
type Processor struct {
	decoder *Decoder
	tasr    bool
}

// NewProcessor initialize decoder returing processor.
func NewProcessor(schema []byte, opts ...ProcessorOption) (*Processor, error) {
	decoder, err := NewDecoder(string(schema))
	if err != nil {
		return nil, err
	}
	processor := &Processor{decoder: decoder}
	for _, opt := range opts {
		opt(processor)
	}
	return processor, nil
}

// Process decodes the avro buffer colorizing the message printing in the stdout.
func (p *Processor) Process(ctx context.Context, topic string, buf []byte) error {
	if len(buf) != 0 {
		msg, err := p.decode(buf)
		if err != nil {
			logrus.WithError(err).Error("Parsing failed")
			return err
		}
		v, err := p.format(msg)
		if err != nil {
			logrus.WithError(err).Error("could not format incoming message")
			return err
		}
		fmt.Println(string(v))
	}
	return nil
}

func (p *Processor) decode(buf []byte) ([]byte, error) {
	if p.tasr && len(buf) > 3 {
		logrus.Debugf("TASR ID Type: %v\n", buf[0])
		logrus.Debugf("TASR Version: %v\n", buf[2])
		buf = buf[3:]
	}
	return p.decoder.Decode(buf)
}

func (p *Processor) format(src []byte) ([]byte, error) {
	var obj map[string]any
	if err := json.Unmarshal(src, &obj); err != nil {
		return nil, err
	}
	return colorjson.Marshal(obj)
}
