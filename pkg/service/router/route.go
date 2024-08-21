// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package router

type Route struct {
	pattern string
	body    string
	method  MethodType

	grpcSpec *GrpcSpec
}

func (r *Route) Path() string {
	return r.pattern
}

func (r *Route) Method() MethodType {
	return r.method
}

func (r *Route) GrpcSpec() *GrpcSpec {
	return r.grpcSpec
}

func NewRoute(pattern, body string, method MethodType, spec *GrpcSpec) *Route {
	return &Route{
		pattern:  pattern,
		body:     body,
		method:   method,
		grpcSpec: spec,
	}
}
