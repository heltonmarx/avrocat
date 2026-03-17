package main

import (
	"github.com/linkedin/goavro/v2"
	"github.com/sirupsen/logrus"
)

// Decoder wraps an Avro codec for binary-to-textual conversion.
type Decoder struct {
	codec *goavro.Codec
}

// NewDecoder initializes the avro codec and returns a Decoder.
func NewDecoder(schema string) (*Decoder, error) {
	codec, err := goavro.NewCodec(schema)
	if err != nil {
		logrus.WithError(err).Error("could not create new codec")
		return nil, err
	}
	return &Decoder{codec}, nil
}

// Decode attempts to decode the bytes in buf using the Decoder's codec.
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
