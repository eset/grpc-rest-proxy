// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package jsonencoder

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"

	jErrors "github.com/juju/errors"
)

type Config struct {
	EmitUnpopulated   bool `mapstructure:"emitUnpopulated"`
	EmitDefaultValues bool `mapstructure:"emitDefaultValues"`
}

type Encoder struct {
	opts protojson.MarshalOptions
}

// New creates a new JSON encoder.
// Type resolver is used to resolve types of messages and can be nil in which case the default resolver is used.
func New(cfg *Config, typeResolver *protoregistry.Types) Encoder {
	return Encoder{
		opts: protojson.MarshalOptions{
			EmitUnpopulated:   cfg.EmitUnpopulated,
			EmitDefaultValues: cfg.EmitDefaultValues,
			Resolver:          typeResolver,
		},
	}
}

func (e Encoder) Encode(m proto.Message) ([]byte, error) {
	response, err := e.opts.Marshal(m)
	if err != nil {
		return nil, jErrors.Trace(err)
	}

	return response, nil
}
