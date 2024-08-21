// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package protoparser

import (
	"strings"

	"github.com/eset/grpc-rest-proxy/pkg/service/router"

	"google.golang.org/protobuf/reflect/protoregistry"
)

type ParseResult struct {
	FileRegistry *protoregistry.Files
	TypeResolver *protoregistry.Types
	Routes       []*router.Route
	Errors       []error
}

func (r *ParseResult) Ok() bool {
	return len(r.Errors) == 0
}

func (r *ParseResult) AddError(err error) {
	r.Errors = append(r.Errors, err)
}

func (r *ParseResult) AddRoute(route *router.Route) {
	r.Routes = append(r.Routes, route)
}

func (r *ParseResult) ErrorsString() string {
	var sb strings.Builder

	for _, err := range r.Errors {
		sb.WriteString(err.Error())
		sb.WriteString("; ")
	}

	return sb.String()
}
