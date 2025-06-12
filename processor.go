package main

import (
	"context"
	"encoding/json"

	"github.com/TylerBrock/colorjson"
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
func (p *Processor) Process(ctx context.Context, topic string, buf []byte) ([]byte, error) {
	if len(buf) == 0 {
		return buf, nil
	}
	msg, err := p.decode(buf)
	if err != nil {
		return nil, err
	}
	return p.format(msg)
}

func (p *Processor) decode(buf []byte) ([]byte, error) {
	if p.tasr && len(buf) > 3 {
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
