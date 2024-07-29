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
	"github.com/eset/grpc-rest-proxy/pkg/service/protoparser"
	"github.com/eset/grpc-rest-proxy/pkg/transport"
	"github.com/eset/grpc-rest-proxy/pkg/transport/http"
	routerPkg "github.com/eset/grpc-rest-proxy/pkg/transport/router"
)

type App struct {
	conf            *Config
	serverHTTP      *http.Server
	router          *routerPkg.ReloadableRouter
	descriptorsRepo descriptors.Descriptors
	gateways        *gateways
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

	router, err := app.createRouter(ctx)
	if err != nil {
		return nil, jErrors.Annotate(jErrors.Trace(err), "failed to create router")
	}
	app.router = routerPkg.WithWrapper(router)

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
	routerContext := &transport.Context{
		Router:     app.router,
		GrcpClient: app.gateways.grpcClient,
	}
	handler := transport.NewHandler(routerContext, logging.Default())
	app.serverHTTP = http.NewServer(app.conf.Transport.HTTP.Server, handler)
}

func (app *App) createRouter(ctx context.Context) (*routerPkg.Router, error) {
	router, err := app.getRouterWithRoutes(ctx)
	if err != nil {
		return nil, jErrors.Trace(err)
	}
	return router, nil
}

func (app *App) reloadRouter(ctx context.Context, r *routerPkg.ReloadableRouter) error {
	routerRoutes, err := app.getRouterWithRoutes(ctx)
	if err != nil {
		return jErrors.Trace(err)
	}
	r.SetRouter(routerRoutes)
	return nil
}

func (app *App) getRouterWithRoutes(ctx context.Context) (*routerPkg.Router, error) {
	fileDescriptorSet, err := app.descriptorsRepo.GetProtoFileDescriptorSet(ctx)
	if err != nil {
		return nil, jErrors.Annotate(jErrors.Trace(err), "failed to retrieve proto descriptors from source")
	}

	parseResult := protoparser.ParseFileDescSets(fileDescriptorSet)
	if !parseResult.Ok() {
		return nil, jErrors.Trace(jErrors.New(parseResult.ErrorsString()))
	}

	routerRoutes := routerPkg.NewRouter()

	for _, route := range parseResult.Routes {
		err = routerRoutes.Push(route)
		if err != nil {
			return nil, jErrors.Trace(err)
		}
		logging.Info(fmt.Sprintf("Added route: [%s] %s", routerPkg.MethodToString(route.Method()), route.Path()))
	}
	return routerRoutes, nil
}

func (app *App) listenForSignal(ctx context.Context, sigUsr1 <-chan os.Signal) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-sigUsr1:
			logging.Info("reload signal received")

			err := app.reloadRouter(ctx, app.router)
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
