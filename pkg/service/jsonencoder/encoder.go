// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package jsonencoder

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	jErrors "github.com/juju/errors"
)

type Config struct {
	EmitUnpopulated   bool `mapstructure:"emitUnpopulated"`
	EmitDefaultValues bool `mapstructure:"emitDefaultValues"`
}

type Encoder struct {
	opts protojson.MarshalOptions
}

func NewOptions(cfg *Config) Encoder {
	return Encoder{
		opts: protojson.MarshalOptions{EmitUnpopulated: cfg.EmitUnpopulated, EmitDefaultValues: cfg.EmitDefaultValues},
	}
}

func (e Encoder) Format(m proto.Message) (string, error) {
	response, err := e.opts.Marshal(m)
	if err != nil {
		return "", jErrors.Trace(err)
	}

	return string(response), nil
}
