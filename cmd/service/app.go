// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package main

import (
	"context"
	"fmt"
	logging "log/slog"
	"os"
	"os/signal"
	"syscall"

	jErrors "github.com/juju/errors"

	grpcClient "github.com/eset/grpc-rest-proxy/pkg/gateway/grpc"
	"github.com/eset/grpc-rest-proxy/pkg/repository/descriptors"
	"github.com/eset/grpc-rest-proxy/pkg/service/jsonencoder"
	"github.com/eset/grpc-rest-proxy/pkg/service/protoparser"
	routerPkg "github.com/eset/grpc-rest-proxy/pkg/service/router"
	"github.com/eset/grpc-rest-proxy/pkg/transport"
	"github.com/eset/grpc-rest-proxy/pkg/transport/http"
)

type App struct {
	conf            *Config
	serverHTTP      *http.Server
	descriptorsRepo descriptors.Descriptors
	gateways        *gateways
	reloader        *transport.EndpointReloader
}

type gateways struct {
	grpcClient grpcClient.ClientInterface
}

func New(ctx context.Context, conf *Config) (*App, error) {
	var err error
	app := &App{
		conf: conf,
	}
	app.gateways, err = createGateways(conf)
	if err != nil {
		return nil, jErrors.Trace(err)
	}

	app.descriptorsRepo, err = descriptors.New(conf.Descriptors, app.gateways.grpcClient)
	if err != nil {
		return nil, jErrors.Trace(err)
	}

	endpointProxy, err := app.createProxyEndpoint(ctx)
	if err != nil {
		return nil, jErrors.Annotate(jErrors.Trace(err), "failed to create router")
	}

	app.reloader = transport.NewEndpointReloader(endpointProxy)
	app.createHTTPServer()
	return app, nil
}

func createGateways(conf *Config) (*gateways, error) {
	client, err := grpcClient.NewClient(conf.Gateways.GrpcClientConfig)
	if err != nil {
		return nil, jErrors.Trace(err)
	}

	return &gateways{
		grpcClient: client,
	}, nil
}

func (app *App) createHTTPServer() {
	handler := transport.NewHandler(app.reloader)
	app.serverHTTP = http.NewServer(app.conf.Transport.HTTP.Server, handler)
}

func (app *App) reloadEndpoint(ctx context.Context) error {
	endpoint, err := app.createProxyEndpoint(ctx)
	if err != nil {
		return jErrors.Trace(err)
	}

	app.reloader.Set(endpoint)
	return nil
}

func (app *App) createProxyEndpoint(ctx context.Context) (*transport.ProxyEndpoint, error) {
	fileDescriptorSet, err := app.descriptorsRepo.GetProtoFileDescriptorSet(ctx)
	if err != nil {
		return nil, jErrors.Annotate(jErrors.Trace(err), "failed to retrieve proto descriptors from source")
	}

	parseResult := protoparser.ParseFileDescSets(fileDescriptorSet)
	if !parseResult.Ok() {
		return nil, jErrors.Trace(jErrors.New(parseResult.ErrorsString()))
	}

	router := routerPkg.NewRouter()

	for _, route := range parseResult.Routes {
		err = router.Push(route)
		if err != nil {
			return nil, jErrors.Trace(err)
		}
		logging.Info(fmt.Sprintf("Added route: [%s] %s", routerPkg.MethodToString(route.Method()), route.Path()))
	}

	encoder := jsonencoder.New(app.conf.Service.JSONEncoder, parseResult.TypeResolver)

	return transport.NewProxyEndpoint(
		logging.Default(),
		router,
		app.gateways.grpcClient,
		encoder,
	), nil
}

func (app *App) listenForSignal(ctx context.Context, sigUsr1 <-chan os.Signal) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-sigUsr1:
			logging.Info("reload signal received")

			err := app.reloadEndpoint(ctx)
			if err != nil {
				logging.Error(jErrors.Details(jErrors.Trace(err)))
			}
		}
	}
}

func (app *App) handleSignal(ctx context.Context) {
	sigUsr1 := make(chan os.Signal, 1)
	signal.Notify(sigUsr1, syscall.SIGUSR1)

	go app.listenForSignal(ctx, sigUsr1)
}

func (app *App) Run(ctx context.Context) error {
	defer app.gateways.grpcClient.Close()
	defer app.serverHTTP.Close()

	app.handleSignal(ctx)
	return jErrors.Trace(http.ListenAndServe(ctx, app.serverHTTP))
}
