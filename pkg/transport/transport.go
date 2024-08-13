// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package transport

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	grpcClient "github.com/eset/grpc-rest-proxy/pkg/gateway/grpc"
	"github.com/eset/grpc-rest-proxy/pkg/service/transformer"
	routerPkg "github.com/eset/grpc-rest-proxy/pkg/transport/router"

	logging "log/slog"

	"github.com/go-chi/chi/v5"
	jErrors "github.com/juju/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/dynamicpb"
)

type Context struct {
	Router     *routerPkg.ReloadableRouter
	GrcpClient grpcClient.ClientInterface
}

type Logger interface {
	// Log error during request processing
	ErrorContext(ctx context.Context, msg string, args ...any)
}

func NewHandler(routerContext *Context, logger Logger) http.Handler {
	routes := chi.NewRouter()
	if logger == nil {
		logger = logging.Default()
	}
	routes.HandleFunc("/*", createRoutingEndpoint(routerContext, logger))
	routes.Get("/status", statusJSON)
	return routes
}

func getQueryVariables(queryValues url.Values) []transformer.Variable {
	var queryVariables []transformer.Variable
	for name, values := range queryValues {
		fieldPath := strings.Split(name, ".")
		for _, value := range values {
			queryVariables = append(queryVariables, transformer.Variable{FieldPath: fieldPath, Value: value})
		}
	}
	return queryVariables
}

func convertRequestToGRPC(route *routerPkg.Match, r *http.Request) (req *dynamicpb.Message, resp *dynamicpb.Message, err error) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, jErrors.Trace(err)
	}
	r.Body.Close()

	queryVariables := getQueryVariables(r.URL.Query())
	route.Params = append(route.Params, queryVariables...)

	req, err = transformer.GetRPCRequest(reqBody, route.GrpcSpec.RequestDesc, route.Params, route.BodyRule)
	if err != nil {
		return nil, nil, jErrors.Trace(err)
	}
	resp = transformer.GetRPCResponse(route.GrpcSpec.ResponseDesc)

	return req, resp, nil
}

func createRoutingEndpoint(rc *Context, logger Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		method, err := routerPkg.StringToMethod(r.Method)
		if err != nil {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		routeMatch := rc.Router.Find(method, r.URL.Path)
		if routeMatch == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		rpcRequest, rpcResponse, err := convertRequestToGRPC(routeMatch, r)
		if err != nil {
			logger.ErrorContext(r.Context(), jErrors.Details(jErrors.Trace(err)))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var header, trailer metadata.MD
		err = rc.GrcpClient.Invoke(
			transformer.GetRPCRequestContext(r),
			routeMatch.GrpcSpec.FullPath(),
			rpcRequest,
			rpcResponse,
			grpc.Header(&header),
			grpc.Trailer(&trailer),
		)
		if err != nil {
			if e, ok := status.FromError(err); ok {
				transformer.SetRESTHeaders(r.ProtoMajor, w.Header(), header, trailer)
				w.WriteHeader(transformer.GetHTTPStatusCode(e.Code()))
			}
			logger.ErrorContext(r.Context(), jErrors.Details(jErrors.Trace(err)))
			return
		}

		transformer.SetRESTHeaders(r.ProtoMajor, w.Header(), header, trailer)

		response, err := protojson.Marshal(rpcResponse)
		if err != nil {
			logger.ErrorContext(r.Context(), jErrors.Details(jErrors.Trace(err)))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = w.Write(response)
		if err != nil {
			logger.ErrorContext(r.Context(), jErrors.Details(jErrors.Trace(err)))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func statusJSON(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, err := io.WriteString(w, `{"status":"OK"}`)
	if err != nil {
		logging.Error(fmt.Sprintf("write http response body error for /status: %s", err))
	}
}
