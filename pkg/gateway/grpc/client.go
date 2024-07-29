// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package grpc

import (
	"context"
	"crypto/tls"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	jErrors "github.com/juju/errors"
)

type ClientInterface interface {
	Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error
	NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error)
	Close() error
}

type client struct {
	requestTimeout time.Duration
	grpcClient     grpc.ClientConnInterface
}

func NewClient(c *ClientConfig) (ClientInterface, error) {
	var dialOpts []grpc.DialOption
	var transportCreds credentials.TransportCredentials

	if c.Config.TLS {
		transportCreds = credentials.NewTLS(&tls.Config{InsecureSkipVerify: c.Config.TLSSkipVerify}) //nolint:gosec
	} else {
		transportCreds = insecure.NewCredentials()
	}

	metricInterceptor := createClientMetricsInterceptor()

	dialOpts = append(dialOpts,
		grpc.WithTransportCredentials(transportCreds),
		grpc.WithUnaryInterceptor(metricInterceptor))

	grpcClient, err := grpc.NewClient(c.Config.TargetAddr, dialOpts...)
	if err != nil {
		return nil, jErrors.Trace(err)
	}

	return &client{
		requestTimeout: c.Config.RequestTimeout,
		grpcClient:     grpcClient,
	}, nil
}

func (c *client) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	ctx, cancel := context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()
	err := c.grpcClient.Invoke(ctx, method, args, reply, opts...)
	return err
}

func (c *client) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	stream, err := c.grpcClient.NewStream(ctx, desc, method, opts...)
	return stream, err
}

func (c *client) Close() error {
	if closer, ok := c.grpcClient.(io.Closer); ok {
		return jErrors.Trace(closer.Close())
	}
	return nil
}
