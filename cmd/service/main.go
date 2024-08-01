// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package main

import (
	"context"
	"fmt"
	logging "log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	jErrors "github.com/juju/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultRequestTimeout          = 5 * time.Second
	defaultReadTimeout             = 10 * time.Second
	descriptorTimeout              = time.Minute
	maxRequestSize                 = 10024
	httpServerAddr                 = "0.0.0.0:8080"
	defaultDescriptorsFetchingType = "remote"
	reflectionServiceName          = "grpc.reflection.v1.ServerReflection/ServerReflectionInfo"
	excludedDescriptors            = "grpc.health.v1.Health,grpc.reflection.v1.ServerReflection"
	grpcServerAddr                 = "0.0.0.0:50051"
	tls                            = false
	tlsSkipverify                  = false
)

var (
	Commit  string
	Version string
	Build   string
)

func main() {
	conf := &Config{}

	pflag.Uint("transport.http.maxRequestSizeKB", maxRequestSize, "maximum size of requests in KB")
	pflag.Duration("transport.http.requestTimeout", defaultRequestTimeout, "request timeout")
	pflag.String("transport.http.server.addr", httpServerAddr, "address and port of the HTTP server")
	pflag.Duration("transport.http.server.gracefulTimeout", defaultRequestTimeout, "graceful timeout")
	pflag.Duration("transport.http.server.readTimeout", defaultReadTimeout, "read timeout")
	pflag.Duration("transport.http.server.readHeaderTimeout", defaultRequestTimeout, "read header timeout")

	pflag.String("descriptors.kind", defaultDescriptorsFetchingType, "type of descriptors fetching")
	pflag.Duration("descriptors.remote.timeout", descriptorTimeout, "request timeout for remote descriptors")
	pflag.String("descriptors.remote.reflectionServiceName", reflectionServiceName, "reflection service name")
	pflag.StringArray("descriptors.remote.exclude", strings.Split(excludedDescriptors, ","), "remote descriptors to exclude")

	pflag.Duration("gateways.grpc.requestTimeout", defaultRequestTimeout, "client request timeout")
	pflag.String("gateways.grpc.client.targetAddr", grpcServerAddr, "address and port of the gRPC server")
	pflag.Duration("gateways.grpc.client.requestTimeout", defaultRequestTimeout, "requests timeout")
	pflag.Bool("gateways.grpc.client.tls", tls, "use TLS for gRPC connection")
	pflag.Bool("gateways.grpc.client.tlsSkipverify", tlsSkipverify, "skip TLS verification")
	pflag.BoolP("version", "v", false, "print version")
	configFile := pflag.StringP("config", "c", "", "path to config file")

	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		logging.Error(jErrors.Details(jErrors.Trace(err)))
		os.Exit(1)
	}

	if viper.GetBool("version") {
		fmt.Println("Version:\t", Version)
		fmt.Println("Build:  \t", Build)
		fmt.Println("Commit: \t", Commit)
		os.Exit(0)
	}

	err = viper.Unmarshal(&conf)
	if err != nil {
		logging.Error(jErrors.Details(jErrors.Annotate(err, "decode configuration to struct")))
		os.Exit(1)
	}

	if configFile != nil && *configFile != "" {
		err := loadConfigFromFile(*configFile, conf)
		if err != nil {
			logging.Error(jErrors.Details(jErrors.Trace(err)))
			os.Exit(1)
		}
	}

	err = conf.validate()
	if err != nil {
		logging.Error(jErrors.Details(jErrors.Annotate(err, "configuration validation error")))
		os.Exit(1)
	}

	os.Exit(run(conf))
}

func run(conf *Config) int {
	logging.SetDefault(logging.New(logging.NewJSONHandler(os.Stderr, &logging.HandlerOptions{Level: logging.LevelDebug})))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app, err := New(ctx, conf)
	if err != nil {
		logging.Error(jErrors.Details(jErrors.Trace(err)))
		return 1
	}

	err = app.Run(ctx)
	if err != nil {
		logging.Error(jErrors.Details(jErrors.Trace(err)))
		return 1
	}

	return 0
}
