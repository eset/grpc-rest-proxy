// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package remote

import (
	"strings"
	"time"
)

type Config struct {
	Timeout               time.Duration `mapstructure:"timeout" validate:"required"`
	ReflectionServiceName string        `mapstructure:"reflectionServiceName" validate:"required"`
	Exclude               []string      `mapstructure:"exclude"`
}

func (cfg *Config) getReflectionServicePath() string {
	if strings.HasPrefix(cfg.ReflectionServiceName, "/") {
		return cfg.ReflectionServiceName
	}

	return "/" + cfg.ReflectionServiceName
}
