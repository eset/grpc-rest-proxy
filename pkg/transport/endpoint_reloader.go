// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.
package transport

import (
	"net/http"
	"sync"
)

// EndpointReloader is wrapper for Handle method to allow dynamic endpoint reloading
type EndpointReloader struct {
	mtx     sync.RWMutex
	handler http.Handler
}

func NewEndpointReloader(handler http.Handler) *EndpointReloader {
	return &EndpointReloader{
		handler: handler,
	}
}

func (e *EndpointReloader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.mtx.RLock()
	defer e.mtx.RUnlock()
	e.handler.ServeHTTP(w, r)
}

func (e *EndpointReloader) Set(endpoint *ProxyEndpoint) {
	e.mtx.Lock()
	e.handler = endpoint
	e.mtx.Unlock()
}

func (e *EndpointReloader) Endpoint() http.Handler {
	e.mtx.RLock()
	defer e.mtx.RUnlock()
	return e.handler
}
