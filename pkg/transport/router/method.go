// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package router

import (
	"net/http"
	"strings"

	jErrors "github.com/juju/errors"
)

type MethodType uint

const MethodNotFound = jErrors.ConstError("method not found")

const (
	UnknownMethod MethodType = iota
	CONNECT
	DELETE
	GET
	HEAD
	OPTIONS
	PATCH
	POST
	PUT
	TRACE
)

var methodToEnum = map[string]MethodType{
	http.MethodConnect: CONNECT,
	http.MethodDelete:  DELETE,
	http.MethodGet:     GET,
	http.MethodHead:    HEAD,
	http.MethodOptions: OPTIONS,
	http.MethodPatch:   PATCH,
	http.MethodPost:    POST,
	http.MethodPut:     PUT,
	http.MethodTrace:   TRACE,
}

var enumToMethod = map[MethodType]string{
	CONNECT: http.MethodConnect,
	DELETE:  http.MethodDelete,
	GET:     http.MethodGet,
	HEAD:    http.MethodHead,
	OPTIONS: http.MethodOptions,
	PATCH:   http.MethodPatch,
	POST:    http.MethodPost,
	PUT:     http.MethodPut,
	TRACE:   http.MethodTrace,
}

func StringToMethod(method string) (MethodType, error) {
	method = strings.ToUpper(method)
	m, ok := methodToEnum[method]
	if !ok {
		return m, jErrors.Trace(MethodNotFound)
	}

	return m, nil
}

func MethodToString(method MethodType) string {
	m, ok := enumToMethod[method]
	if !ok {
		return "UNKNOWN"
	}

	return m
}
