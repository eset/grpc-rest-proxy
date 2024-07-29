// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package http

import "time"

type ServerTLSConfig struct {
	Cert string `mapstructure:"cert" validate:"required,file"`
	Key  string `mapstructure:"key" validate:"required,file"`
}

type ServerConfig struct {
	Addr              string           `mapstructure:"addr" validate:"required,hostname_port"`
	GracefulTimeout   time.Duration    `mapstructure:"gracefulTimeout" validate:"required"`
	ReadTimeout       time.Duration    `mapstructure:"readTimeout" validate:"gte=0"`
	ReadHeaderTimeout time.Duration    `mapstructure:"readHeaderTimeout" validate:"gte=0"`
	TLS               *ServerTLSConfig `mapstructure:"tls"`
}
