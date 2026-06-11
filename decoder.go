package main

import (
	"fmt"

	"github.com/linkedin/goavro/v2"
)

type Decoder interface {
	Decode(buf []byte) ([]byte, error)
}

// AvroDecoder wraps an Avro codec for binary-to-textual conversion.
type AvroDecoder struct {
	codec        *goavro.Codec
	stripSchemaID bool
}

// NewAvroDecoder initializes the avro codec and returns a Decoder.
// stripSchemaID controls whether the 5-byte Confluent Schema Registry header
// (magic byte + 4-byte schema ID) is removed before decoding.
func NewAvroDecoder(codec *goavro.Codec, stripSchemaID bool) *AvroDecoder {
	return &AvroDecoder{codec: codec, stripSchemaID: stripSchemaID}
}

// Decode attempts to decode the bytes in buf using the Decoder's codec.
func (d *AvroDecoder) Decode(buf []byte) ([]byte, error) {
	if d.stripSchemaID {
		if len(buf) < 5 {
			return nil, fmt.Errorf("message too short to contain schema registry header (%d bytes)", len(buf))
		}
		buf = buf[5:]
	}

	datum, _, err := d.codec.NativeFromBinary(buf)
	if err != nil {
		return nil, fmt.Errorf("could not parse native from binary: %w", err)
	}
	textual, err := d.codec.TextualFromNative(nil, datum)
	if err != nil {
		return nil, fmt.Errorf("could not parse textual from native: %w", err)
	}
	return textual, nil
}
