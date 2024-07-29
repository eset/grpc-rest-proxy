// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package grpc

import (
	"time"
)

type ClientConfig struct {
	RequestTimeout time.Duration `mapstructure:"requestTimeout" validate:"gt=100ms"`
	Config         *Config       `mapstructure:"client" validate:"required"`
}

type Config struct {
	TargetAddr     string        `mapstructure:"targetAddr" validate:"required"`
	RequestTimeout time.Duration `mapstructure:"requestTimeout" validate:"gt=100ms"`
	TLS            bool          `mapstructure:"tls"`
	TLSSkipVerify  bool          `mapstructure:"tlsSkipverify"`
}
