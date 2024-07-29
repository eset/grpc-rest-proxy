// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package remote

import (
	"context"
	"errors"
	"io"
	logging "log/slog"
	"slices"

	grpcClient "github.com/eset/grpc-rest-proxy/pkg/gateway/grpc"

	jErrors "github.com/juju/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection/grpc_reflection_v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	FileDescriptorsNotFound = jErrors.ConstError("descriptors on remote servers not found")
)

type Remote struct {
	cfg        *Config
	grpcClient grpcClient.ClientInterface
}

func New(cfg *Config, client grpcClient.ClientInterface) (*Remote, error) {
	return &Remote{cfg: cfg, grpcClient: client}, nil
}

func (r *Remote) GetProtoFileDescriptorSet(ctx context.Context) ([]*descriptorpb.FileDescriptorSet, error) {
	var fdSets []*descriptorpb.FileDescriptorSet

	desc := &grpc.StreamDesc{
		ServerStreams: true,
		ClientStreams: true,
	}

	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeout)
	defer cancel()
	fdSet, err := r.readDescSetFromClient(ctx, desc)
	if err != nil {
		return nil, jErrors.Trace(err)
	}
	if fdSet.GetFile() != nil {
		fdSets = append(fdSets, fdSet)
	}

	if len(fdSets) == 0 {
		return nil, jErrors.Trace(FileDescriptorsNotFound)
	}

	logging.Info("descriptors successfully fetched from remote servers")
	return fdSets, nil
}

func (r *Remote) readDescSetFromClient(ctx context.Context, desc *grpc.StreamDesc) (
	*descriptorpb.FileDescriptorSet, error) {
	stream, err := r.grpcClient.NewStream(ctx, desc, r.cfg.getReflectionServicePath(), grpc.WaitForReady(true))
	if err != nil {
		return nil, jErrors.Trace(err)
	}
	defer closeStream(stream)

	fdSet, err := r.readDescriptorFile(stream)
	if err != nil {
		return nil, jErrors.Trace(err)
	}
	return fdSet, nil
}

func closeStream(stream grpc.ClientStream) {
	err := stream.CloseSend()
	if err != nil {
		logging.Error(jErrors.Details(jErrors.Trace(err)))
	}
}

func getListServices(stream grpc.ClientStream) ([]*grpc_reflection_v1.ServiceResponse, error) {
	req := &grpc_reflection_v1.ServerReflectionRequest{
		MessageRequest: &grpc_reflection_v1.ServerReflectionRequest_ListServices{},
	}

	err := stream.SendMsg(req)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, jErrors.Trace(err)
	}

	response := &grpc_reflection_v1.ServerReflectionResponse{}
	err = stream.RecvMsg(response)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, jErrors.Trace(err)
	}

	return response.GetListServicesResponse().GetService(), nil
}

func getFileDescriptorProto(stream grpc.ClientStream, serviceName string) ([][]byte, error) {
	req := &grpc_reflection_v1.ServerReflectionRequest{}
	req.MessageRequest = &grpc_reflection_v1.ServerReflectionRequest_FileContainingSymbol{FileContainingSymbol: serviceName}

	err := stream.SendMsg(req)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, jErrors.Trace(err)
	}

	response := &grpc_reflection_v1.ServerReflectionResponse{}
	err = stream.RecvMsg(response)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, jErrors.Trace(err)
	}

	return response.GetFileDescriptorResponse().GetFileDescriptorProto(), nil
}

func getFileDescriptorSet(fdSet *descriptorpb.FileDescriptorSet, stream grpc.ClientStream, serviceName string) error {
	fdsProto, err := getFileDescriptorProto(stream, serviceName)
	if err != nil {
		return jErrors.Trace(err)
	}

	for _, protoFile := range fdsProto {
		fdProto := new(descriptorpb.FileDescriptorProto)
		if err = proto.Unmarshal(protoFile, fdProto); err != nil {
			return jErrors.Trace(err)
		}
		fdSet.File = append(fdSet.GetFile(), fdProto)
	}

	return nil
}

func (r *Remote) readDescriptorFile(stream grpc.ClientStream) (*descriptorpb.FileDescriptorSet, error) {
	fdSet := new(descriptorpb.FileDescriptorSet)

	services, err := getListServices(stream)
	if err != nil {
		return nil, jErrors.Trace(err)
	}

	for _, service := range services {
		if slices.Contains(r.cfg.Exclude, service.GetName()) {
			continue
		}
		err = getFileDescriptorSet(fdSet, stream, service.GetName())
		if err != nil {
			return nil, jErrors.Trace(err)
		}
	}

	return fdSet, nil
}
