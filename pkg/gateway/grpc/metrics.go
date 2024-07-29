// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package grpc

import (
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

func createClientMetricsInterceptor() grpc.UnaryClientInterceptor {
	clMetrics := grpcprom.NewClientMetrics(grpcprom.WithClientHandlingTimeHistogram())
	prometheus.MustRegister(clMetrics)
	return clMetrics.UnaryClientInterceptor()
}
