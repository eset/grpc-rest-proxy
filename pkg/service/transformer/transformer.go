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

	headerAccept        = "accept"
	headerContentType   = "content-type"
	headerContentLength = "content-length"

	// connection specific headers http1.0/1.1
	headerConnection       = "connection"
	headerProxyConnection  = "proxy-connection"
	headerKeepAlive        = "keep-alive"
	headerTransferEncoding = "transfer-encoding"
	headerUpgrade          = "upgrade"
)

func isConnectionSpecificHeader(name string) bool {
	switch name {
	case headerConnection, headerProxyConnection, headerKeepAlive, headerTransferEncoding, headerUpgrade:
		return true
	}
	return false
}

func GetRPCRequestContext(request *http.Request) context.Context {
	grpcMetadata := metadata.Pairs()

	for name, values := range request.Header {
		name = strings.ToLower(name)

		// in case the client sends a content-length header it will be removed before proceeding
		if name == headerContentLength {
			continue
		}
		// RFC 9113 8.2.2.: endpoint MUST NOT generate an HTTP/2 message containing connection-specific header fields
		if request.ProtoMajor > 1 && isConnectionSpecificHeader(name) {
			continue
		}
		grpcMetadata.Append(name, values...)
	}

	grpcMetadata.Set(headerAccept, "application/protobuf")
	grpcMetadata.Set(headerContentType, "application/protobuf")

	return metadata.NewOutgoingContext(request.Context(), grpcMetadata)
}

func setHeader(headers http.Header, protoMajor int, name string, values []string) {
	// RFC 9113 8.2.2.: endpoint MUST NOT generate an HTTP/2 message containing connection-specific header fields
	if protoMajor > 1 && isConnectionSpecificHeader(name) {
		return
	}
	for _, value := range values {
		headers.Add(name, value)
	}
}

func SetRESTHeaders(protoMajor int, headers http.Header, gRPCheader metadata.MD, gRPCTrailer metadata.MD) {
	// set headers
	for name, values := range gRPCheader {
		setHeader(headers, protoMajor, name, values)
	}
	// append trailers as headers
	for name, values := range gRPCTrailer {
		setHeader(headers, protoMajor, name, values)
	}

	headers.Set(headerContentType, "application/json")
}

func GetRPCResponse(responseDesc protoreflect.MessageDescriptor) *dynamicpb.Message {
	return dynamicpb.NewMessage(responseDesc)
}

// https://chromium.googlesource.com/external/github.com/grpc/grpc/+/refs/tags/v1.21.4-pre1/doc/statuscodes.md
//
//nolint:gocyclo,mnd
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
