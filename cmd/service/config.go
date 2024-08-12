// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package main

import (
	"bytes"
	"os"

	"github.com/eset/grpc-rest-proxy/pkg/gateway/grpc"
	"github.com/eset/grpc-rest-proxy/pkg/repository/descriptors"
	"github.com/eset/grpc-rest-proxy/pkg/service/jsonencoder"
	"github.com/eset/grpc-rest-proxy/pkg/transport"

	"github.com/go-playground/validator/v10"
	jErrors "github.com/juju/errors"
	"github.com/spf13/viper"
)

type Config struct {
	Transport   *transport.Config   `mapstructure:"transport" validate:"required"`
	Descriptors *descriptors.Config `mapstructure:"descriptors" validate:"required"`
	Gateways    *Gateway            `mapstructure:"gateways" validate:"required"`
	Service     *Service            `mapstructure:"service" validate:"required"`
}

type Gateway struct {
	GrpcClientConfig *grpc.ClientConfig `mapstructure:"grpc"`
}

type Service struct {
	JSONEncoder *jsonencoder.Config `mapstructure:"jsonencoder"`
}

func (c *Config) validate() error {
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return jErrors.Trace(err)
	}
	return nil
}

func loadConfigFromFile(filepath string, c *Config) error {
	viper.SetConfigType("yml")

	file, err := os.ReadFile(filepath)
	if err != nil {
		return jErrors.Trace(err)
	}

	err = viper.ReadConfig(bytes.NewReader(file))
	if err != nil {
		return jErrors.Annotatef(err, "failed to parse config file %s", filepath)
	}

	err = viper.Unmarshal(&c)
	if err != nil {
		return jErrors.Annotate(err, "decode configuration to struct")
	}
	return nil
}
