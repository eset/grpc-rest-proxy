package status

import (
	"net/http"

	"github.com/eset/grpc-rest-proxy/pkg/service/transformer"

	grpcStatus "google.golang.org/grpc/status"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

func FromHTTPCode(code int) *Error {
	return &Error{
		Code:    int32(code), //nolint:gosec
		Message: http.StatusText(code),
	}
}

func FromGRPC(status *grpcStatus.Status) *Error {
	httpStatus := transformer.GetHTTPStatusCode(status.Code())

	msg := status.Message()
	if msg == "" {
		msg = http.StatusText(httpStatus)
	}

	var details []*anypb.Any
	if status.Proto() != nil {
		details = status.Proto().GetDetails()
	}

	return &Error{
		Code:    int32(httpStatus), //nolint:gosec
		Message: msg,
		Details: details,
	}
}
