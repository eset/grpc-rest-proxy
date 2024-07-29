// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package protoparser_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/eset/grpc-rest-proxy/pkg/service/protoparser"

	jErrors "github.com/juju/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	_ "google.golang.org/protobuf/types/known/anypb"
	_ "google.golang.org/protobuf/types/known/timestamppb"
)

func TestProtoParser(t *testing.T) {
	protoFile, err := os.ReadFile("../../../cmd/examples/grpcserver/gen/user/v1/user.desc")
	require.NoError(t, err)

	pbSet := new(descriptorpb.FileDescriptorSet)
	require.NoError(t, err)
	require.NoError(t, proto.Unmarshal(protoFile, pbSet))

	result := protoparser.ParseFileDescSets([]*descriptorpb.FileDescriptorSet{pbSet})

	for _, route := range result.Routes {
		fmt.Println(route.Path())
	}

	for _, err := range result.Errors {
		fmt.Println(jErrors.Details(err))
	}

	require.True(t, result.Ok())
}

type gRPCServiceName struct {
	FullName string
	Service  string
	Method   string
}

func TestParseServiceNameAndMethod(t *testing.T) {
	routes := []gRPCServiceName{
		{FullName: "eset.dps.enumservice.service.grpcEnumListService.getApplications",
			Method:  "getApplications",
			Service: "/eset.dps.enumservice.service.grpcEnumListService"},
		{FullName: "eset.dps.enumservice.service.grpcEnumListService.getEsetIpAddresses",
			Method:  "getEsetIpAddresses",
			Service: "/eset.dps.enumservice.service.grpcEnumListService"},
		{FullName: "eset.dotnod.isp_management.v1.Isps.GetIsp",
			Method:  "GetIsp",
			Service: "/eset.dotnod.isp_management.v1.Isps"},
		{FullName: "eset.Method",
			Method:  "Method",
			Service: "/eset"},
		{FullName: "eset.Method",
			Method:  "Method",
			Service: "/eset"},
	}

	for _, route := range routes {
		service, method, err := protoparser.ParseServiceNameAndMethod(route.FullName)

		require.NoError(t, err)
		require.Equal(t, route.Service, service)
		require.Equal(t, route.Method, method)
	}
}
