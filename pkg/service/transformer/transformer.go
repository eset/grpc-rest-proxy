// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package transformer

import (
	"context"
	"net/http"
	"strings"

	jErrors "github.com/juju/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

const (
	InvalidGrpcMethodName = jErrors.ConstError("gRPC method name is invalid")
)

func GetRPCRequestContext(request *http.Request) context.Context {
	grpcMetadata := metadata.Pairs()

	for name, values := range request.Header {
		name = strings.ToLower(name)

		// in case the client sends a content-length header it will be removed before proceeding
		if name == "content-length" {
			continue
		}
		grpcMetadata.Append(name, values...)
	}

	grpcMetadata.Set("accept", "application/protobuf")
	grpcMetadata.Set("content-type", "application/protobuf")

	return metadata.NewOutgoingContext(request.Context(), grpcMetadata)
}

func SetRESTHeaders(headers http.Header, gRPCheader metadata.MD, gRPCTrailer metadata.MD) {
	// set headers
	for name, values := range gRPCheader {
		for _, value := range values {
			headers.Add(name, value)
		}
	}
	// append trailers as headers
	for name, values := range gRPCTrailer {
		for _, value := range values {
			headers.Add(name, value)
		}
	}

	headers.Set("content-type", "application/json")
}

func GetRPCResponse(responseDesc protoreflect.MessageDescriptor) *dynamicpb.Message {
	return dynamicpb.NewMessage(responseDesc)
}

// https://chromium.googlesource.com/external/github.com/grpc/grpc/+/refs/tags/v1.21.4-pre1/doc/statuscodes.md
//
//nolint:gocyclo,gomnd
func GetHTTPStatusCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return 499
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
