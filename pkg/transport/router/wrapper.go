// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package router

import (
	"sync"
)

type ReloadableRouter struct {
	router *Router
	mtx    sync.RWMutex
}

func WithWrapper(router *Router) *ReloadableRouter {
	return &ReloadableRouter{
		router: router,
		mtx:    sync.RWMutex{},
	}
}

func (w *ReloadableRouter) Find(method MethodType, path string) (result *Match) {
	w.mtx.RLock()
	defer w.mtx.RUnlock()
	return w.router.Find(method, path)
}

func (w *ReloadableRouter) SetRouter(router *Router) {
	w.mtx.Lock()
	w.router = router
	w.mtx.Unlock()
}
