// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package router_test

import (
	"testing"

	"github.com/eset/grpc-rest-proxy/pkg/service/router"

	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/api/annotations"
)

type routeTest struct {
	method  router.MethodType
	path    string
	found   bool
	resPath string
}

var routeTests = []routeTest{
	{
		method:  router.GET,
		path:    "/api/v1/rules/1234",
		found:   true,
		resPath: "t1",
	},
	{
		method:  router.GET,
		path:    "/api/v2/rules/1234/body/1/2/4/",
		found:   true,
		resPath: "t3",
	},
}

func TestRouter(t *testing.T) {
	tree := router.NewRouter()
	msgDesc := (&annotations.HttpRule{}).ProtoReflect().Descriptor()

	routes := []*router.Route{
		router.NewRoute("/api/v1/rules/{selector}", "", router.GET, &router.GrpcSpec{Service: "t1", Method: "m1", RequestDesc: msgDesc}),
		router.NewRoute("/api/v1/rules/{selector}", "", router.POST, &router.GrpcSpec{Service: "t1", Method: "m2", RequestDesc: msgDesc}),
		router.NewRoute("/api/v1/rules/{selector}/{get=*}", "", router.POST, &router.GrpcSpec{Service: "t2", Method: "m2", RequestDesc: msgDesc}),
		router.NewRoute("/api/v2/rules/{selector}/body/{body=**}", "", router.GET, &router.GrpcSpec{Service: "t3", Method: "m3", RequestDesc: msgDesc}),
		router.NewRoute("/api/v2/rules/body/test1", "", router.GET, &router.GrpcSpec{Service: "t3", Method: "m3", RequestDesc: msgDesc}),
		router.NewRoute("/api/v2/rules/body/test2", "*", router.GET, &router.GrpcSpec{Service: "t3", Method: "m3", RequestDesc: msgDesc}),
		router.NewRoute("/api/v2/rules/body/test3", "custom.path", router.GET, &router.GrpcSpec{Service: "t3", Method: "m3", RequestDesc: msgDesc}),
	}

	for _, route := range routes {
		require.NoError(t, tree.Push(route))
	}

	redudantRoute := router.NewRoute("/api/v1/users/{id}", "", router.POST, &router.GrpcSpec{Service: "t3", Method: "m3"})
	require.Error(t, tree.Push(redudantRoute))

	incorrectRoute := router.NewRoute("/v1/package/{id/other", "", router.HEAD, &router.GrpcSpec{Service: "t3", Method: "m3"})
	require.Error(t, tree.Push(incorrectRoute))

	incorrectBodyPath := router.NewRoute("/api/v2/rules/body/test4", "-", router.GET, &router.GrpcSpec{Service: "t3", Method: "m3", RequestDesc: msgDesc})
	require.Error(t, tree.Push(incorrectBodyPath))

	incorrectBodyPath2 := router.NewRoute("/api/v2/rules/body/test4", "test", router.GET, &router.GrpcSpec{Service: "t3", Method: "m3", RequestDesc: msgDesc})
	require.Error(t, tree.Push(incorrectBodyPath2))

	for _, routeTest := range routeTests {
		res := tree.Find(routeTest.method, routeTest.path)

		if routeTest.found {
			require.NotNil(t, res)
		} else {
			require.Nil(t, res)
		}

		require.Equal(t, routeTest.resPath, res.GrpcSpec.Service)
	}
}
