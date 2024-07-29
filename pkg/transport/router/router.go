// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package router

import (
	"github.com/eset/grpc-rest-proxy/pkg/service/transformer"
	routePattern "github.com/eset/grpc-rest-proxy/pkg/transport/router/pattern"

	jErrors "github.com/juju/errors"
)

type Match struct {
	Params   []transformer.Variable
	GrpcSpec *GrpcSpec
	Pattern  string
	BodyRule transformer.HTTPBodyRule
}

type Router struct {
	routesByMethod map[MethodType][]routeMatcher
}

func NewRouter() *Router {
	return &Router{
		routesByMethod: make(map[MethodType][]routeMatcher),
	}
}

type routeMatcher struct {
	matcher  *routePattern.Matcher
	grpcSpec *GrpcSpec
	pattern  string
	bodyRule transformer.HTTPBodyRule
}

func (r *Router) Find(method MethodType, path string) (result *Match) {
	routes, ok := r.routesByMethod[method]
	if !ok {
		return nil
	}

	for _, route := range routes {
		matchRes := route.matcher.Match(path)
		if matchRes.Matched {
			return &Match{
				GrpcSpec: route.grpcSpec,
				Pattern:  route.pattern,
				BodyRule: route.bodyRule,
				Params:   matchRes.Vars,
			}
		}
	}

	return nil
}

func (r *Router) Push(route *Route) error {
	matcher, err := routePattern.Parse(route.pattern)
	if err != nil {
		return jErrors.Trace(err)
	}

	if route.grpcSpec.RequestDesc == nil {
		return jErrors.Errorf("request descriptor is required")
	}

	variablePaths := matcher.GetAllVariablePaths()
	for _, variablePath := range variablePaths {
		err = transformer.ValidateFieldPath(route.grpcSpec.RequestDesc, variablePath)
		if err != nil {
			return jErrors.Trace(err)
		}
	}

	routes := r.routesByMethod[route.method]

	for _, addedRoute := range routes {
		if addedRoute.pattern == route.pattern {
			return jErrors.Errorf("duplicate route: %s", route.pattern)
		}
	}

	bodyRule := transformer.GetHTTPBodyRule(route.body)
	if bodyRule.RuleType == transformer.FieldPathRule {
		err = transformer.ValidateFieldPath(route.grpcSpec.RequestDesc, bodyRule.FieldPath)
		if err != nil {
			return jErrors.Trace(err)
		}
	}

	routes = append(routes, routeMatcher{
		matcher:  matcher,
		pattern:  route.pattern,
		bodyRule: bodyRule,
		grpcSpec: route.grpcSpec,
	})

	r.routesByMethod[route.method] = routes
	return nil
}
