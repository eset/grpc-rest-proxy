// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package transport

import (
	"time"

	"github.com/eset/grpc-rest-proxy/pkg/transport/http"
)

type Config struct {
	HTTP *ConfigHTTP `mapstructure:"http" validate:"required"`
}

type ConfigHTTP struct {
	MaxRequestSizeKB uint               `mapstructure:"maxRequestSizeKB"`
	RequestTimeout   time.Duration      `mapstructure:"requestTimeout" validate:"gte=0"`
	Server           *http.ServerConfig `mapstructure:"server" validate:"required"`
}
