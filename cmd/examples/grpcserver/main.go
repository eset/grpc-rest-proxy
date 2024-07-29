// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package main

import (
	"fmt"
	"net"

	logging "log/slog"

	jErrors "github.com/juju/errors"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/eset/grpc-rest-proxy/cmd/examples/grpcserver/gen/user/v1"
)

const (
	addr = "0.0.0.0"
	port = "50051"
)

func main() {
	addr := pflag.String("addr", addr, "address and port of the gRPC server")
	port := pflag.String("port", port, "port of the gRPC server")

	serverAddr := fmt.Sprintf("%s:%s", *addr, *port)

	grpcServer := newServer()

	err := listenAndServe(serverAddr, grpcServer)
	if err != nil {
		logging.Error("failed to run grpc server", logging.String("grpc-server:", jErrors.Details(err)))
	}
}

func newServer() *grpc.Server {
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, NewUserService())
	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)
	return grpcServer
}

func listenAndServe(addr string, grpcServer *grpc.Server) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return jErrors.Annotatef(err, "gRPC listen '%s'", addr)
	}

	logging.Info("starting gRPC server", logging.String("Addr:", addr))

	err = grpcServer.Serve(lis)
	if err != nil && err != grpc.ErrServerStopped {
		return jErrors.Annotatef(err, "start gRPC server on '%s", addr)
	}
	return nil
}
