package main

import (
	"github.com/linkedin/goavro/v2"
	"github.com/sirupsen/logrus"
)

// Decoder holds the avro codec schema.
type Decoder struct {
	codec *goavro.Codec
}

// NewDecoder initialize the avro codec, returning the decoder.
func NewDecoder(schema string) (*Decoder, error) {
	codec, err := goavro.NewCodec(schema)
	if err != nil {
		logrus.WithError(err).Error("could not create new codec")
		return nil, err
	}
	return &Decoder{codec}, nil
}

// Decode attempts to decode the bytes in buf using one of the Code instances
// in codex.
func (d *Decoder) Decode(buf []byte) ([]byte, error) {
	datum, _, err := d.codec.NativeFromBinary(buf)
	if err != nil {
		logrus.WithError(err).Error("could not parse native from binary")
		return nil, err
	}
	textual, err := d.codec.TextualFromNative(nil, datum)
	if err != nil {
		logrus.WithError(err).Error("could not parse textual from native")
		return nil, err
	}
	return textual, nil
}
