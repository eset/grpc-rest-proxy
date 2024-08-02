// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	logging "log/slog"

	jErrors "github.com/juju/errors"
)

type Server struct {
	conf       *ServerConfig
	handler    http.Handler
	httpServer *http.Server
	logger     *logging.Logger
}

func NewServer(serverConf *ServerConfig, handler http.Handler) *Server {
	return &Server{
		conf:       serverConf,
		handler:    handler,
		httpServer: nil,
		logger:     logging.With("addr", serverConf.Addr),
	}
}

func (server *Server) Close() {
	if server == nil || server.httpServer == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), server.conf.GracefulTimeout)
	defer cancel()

	err := server.httpServer.Shutdown(ctx)
	if err != nil {
		server.logger.Error(fmt.Sprintf("shutdown http server error: %s", err))
		return
	}
	server.logger.Info("graceful shutdown successful")
}

func (server *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", server.conf.Addr)
	if err != nil {
		return jErrors.Trace(err)
	}

	err = jErrors.Trace(server.Serve(ln))
	if !errors.Is(err, http.ErrServerClosed) {
		return jErrors.Trace(err)
	}
	return nil
}

func (server *Server) Serve(listener net.Listener) error {
	server.httpServer = &http.Server{
		Addr:              server.conf.Addr,
		Handler:           server.handler,
		ReadTimeout:       server.conf.ReadTimeout,
		ReadHeaderTimeout: server.conf.ReadHeaderTimeout,
	}

	if server.conf.TLS != nil {
		server.logger.Info("starting HTTPS server with TLS")

		err := server.httpServer.ServeTLS(listener, server.conf.TLS.Cert, server.conf.TLS.Key)
		if !errors.Is(err, http.ErrServerClosed) {
			return jErrors.Trace(err)
		}
		return nil
	}

	server.logger.Info("starting HTTP server")

	err := server.httpServer.Serve(listener)
	if !errors.Is(err, http.ErrServerClosed) {
		return jErrors.Trace(err)
	}
	return nil
}

func ListenAndServe(ctx context.Context, s *Server) error {
	errs := make(chan error)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		errs <- jErrors.Trace(s.ListenAndServe())
		cancel()
	}()

	select {
	case <-ctx.Done():
		logging.Info("received exit signal from context")
		return nil
	case err := <-errs:
		return jErrors.Trace(err)
	}
}
