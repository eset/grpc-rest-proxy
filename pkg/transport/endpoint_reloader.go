// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.
package transport

import (
	"net/http"
	"sync"
)

// EndpointReloader is wrapper for Handle method to allow dynamic endpoint reloading
type EndpointReloader struct {
	mtx   sync.RWMutex
	proxy *ProxyEndpoint
}

func NewEndpointReloader(proxy *ProxyEndpoint) *EndpointReloader {
	return &EndpointReloader{
		proxy: proxy,
	}
}

func (e *EndpointReloader) Handle(w http.ResponseWriter, r *http.Request) {
	e.mtx.RLock()
	defer e.mtx.RUnlock()
	e.proxy.Handle(w, r)
}

func (e *EndpointReloader) Set(endpoint *ProxyEndpoint) {
	e.mtx.Lock()
	e.proxy = endpoint
	e.mtx.Unlock()
}

func (e *EndpointReloader) Endpoint() *ProxyEndpoint {
	e.mtx.RLock()
	defer e.mtx.RUnlock()
	return e.proxy
}
