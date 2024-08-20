// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.
package transport

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"

	grpcClient "github.com/eset/grpc-rest-proxy/pkg/gateway/grpc"
	"github.com/eset/grpc-rest-proxy/pkg/service/jsonencoder"
	routerPkg "github.com/eset/grpc-rest-proxy/pkg/service/router"
	"github.com/eset/grpc-rest-proxy/pkg/service/transformer"
	statusPkg "github.com/eset/grpc-rest-proxy/pkg/transport/status"

	jErrors "github.com/juju/errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	grpcStatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/dynamicpb"
)

type ProxyEndpoint struct {
	logger      Logger
	router      *routerPkg.Router
	client      grpcClient.ClientInterface
	jsonEncoder jsonencoder.Encoder
}

func NewProxyEndpoint(
	logger Logger,
	router *routerPkg.Router,
	client grpcClient.ClientInterface,
	jsonEncoder jsonencoder.Encoder,
) *ProxyEndpoint {
	return &ProxyEndpoint{
		logger:      logger,
		router:      router,
		client:      client,
		jsonEncoder: jsonEncoder,
	}
}

func (e *ProxyEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method, err := routerPkg.StringToMethod(r.Method)
	if err != nil {
		e.respondWithError(r.Context(), w, statusPkg.FromHTTPCode(http.StatusMethodNotAllowed))
		return
	}

	routeMatch := e.router.Find(method, r.URL.Path)
	if routeMatch == nil {
		e.respondWithError(r.Context(), w, statusPkg.FromHTTPCode(http.StatusNotFound))
		return
	}

	rpcRequest, err := convertRequestToGRPC(routeMatch, r)
	if err != nil {
		e.logger.ErrorContext(r.Context(), jErrors.Details(jErrors.Trace(err)))
		e.respondWithError(r.Context(), w, statusPkg.FromHTTPCode(http.StatusBadRequest))
		return
	}
	rpcResponse := transformer.GetRPCResponse(routeMatch.GrpcSpec.ResponseDesc)

	var header, trailer metadata.MD
	err = e.client.Invoke(
		transformer.GetRPCRequestContext(r),
		routeMatch.GrpcSpec.FullPath(),
		rpcRequest,
		rpcResponse,
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)
	if err != nil {
		if errStatus, ok := grpcStatus.FromError(err); ok {
			transformer.SetRESTHeaders(r.ProtoMajor, w.Header(), header, trailer)
			e.respondWithError(r.Context(), w, statusPkg.FromGRPC(errStatus))
			return
		}
		e.logger.ErrorContext(r.Context(), jErrors.Details(jErrors.Trace(err)))
		e.respondWithError(r.Context(), w, statusPkg.FromHTTPCode(http.StatusInternalServerError))
		return
	}

	transformer.SetRESTHeaders(r.ProtoMajor, w.Header(), header, trailer)

	response, err := e.jsonEncoder.Encode(rpcResponse)
	if err != nil {
		e.logger.ErrorContext(r.Context(), jErrors.Details(jErrors.Trace(err)))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(response)
	if err != nil {
		e.logger.ErrorContext(r.Context(), jErrors.Details(jErrors.Trace(err)))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (e *ProxyEndpoint) respondWithError(ctx context.Context, w http.ResponseWriter, status *statusPkg.Error) {
	encodedStatus, err := e.jsonEncoder.Encode(status)
	if err != nil {
		e.logger.ErrorContext(ctx, jErrors.Details(jErrors.Trace(err)))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(int(status.GetCode()))

	_, err = w.Write(encodedStatus)
	if err != nil {
		e.logger.ErrorContext(ctx, jErrors.Details(jErrors.Trace(err)))
	}
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

func convertRequestToGRPC(route *routerPkg.Match, r *http.Request) (req *dynamicpb.Message, err error) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, jErrors.Trace(err)
	}
	r.Body.Close()

	queryVariables := getQueryVariables(r.URL.Query())
	route.Params = append(route.Params, queryVariables...)

	req, err = transformer.GetRPCRequest(reqBody, route.GrpcSpec.RequestDesc, route.Params, route.BodyRule)
	if err != nil {
		return nil, jErrors.Trace(err)
	}

	return req, nil
}
