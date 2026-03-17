package main

import (
	"context"
	"encoding/json"

	"github.com/TylerBrock/colorjson"
)

// Processor holds the avro decoder.
type Processor struct {
	decoder *Decoder
}

// NewProcessor initialize decoder returing processor.
func NewProcessor(schema []byte) (*Processor, error) {
	decoder, err := NewDecoder(string(schema))
	if err != nil {
		return nil, err
	}
	processor := &Processor{decoder: decoder}
	return processor, nil
}

// Process decodes the avro buffer colorizing the message printing in the stdout.
func (p *Processor) Process(ctx context.Context, topic string, buf []byte) ([]byte, error) {
	if len(buf) == 0 {
		return buf, nil
	}
	msg, err := p.decoder.Decode(buf)
	if err != nil {
		return nil, err
	}
	return p.format(msg)
}

func (p *Processor) format(src []byte) ([]byte, error) {
	var obj map[string]any
	if err := json.Unmarshal(src, &obj); err != nil {
		return nil, err
	}
	return colorjson.Marshal(obj)
}
