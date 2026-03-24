package main

import (
	"fmt"

	"github.com/linkedin/goavro/v2"
)

type Decoder interface {
	Decode(buf []byte) ([]byte, error)
}

// Decoder wraps an Avro codec for binary-to-textual conversion.
type AvroDecoder struct {
	codec *goavro.Codec
}

// NewAvroDecoder initializes the avro codec and returns a Decoder.
func NewAvroDecoder(codec *goavro.Codec) *AvroDecoder {
	return &AvroDecoder{codec}
}

// Decode attempts to decode the bytes in buf using the Decoder's codec.
func (d *AvroDecoder) Decode(buf []byte) ([]byte, error) {
	datum, _, err := d.codec.NativeFromBinary(buf)
	if err != nil {
		return nil, fmt.Errorf("could mot parse native from binary: %w", err)
	}
	textual, err := d.codec.TextualFromNative(nil, datum)
	if err != nil {
		return nil, fmt.Errorf("could not parse textual from native: %w", err)
	}
	return textual, nil
}
