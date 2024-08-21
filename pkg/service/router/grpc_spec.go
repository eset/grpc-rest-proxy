// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package router

import (
	"path"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type GrpcSpec struct {
	RequestDesc  protoreflect.MessageDescriptor
	ResponseDesc protoreflect.MessageDescriptor
	Service      string
	Method       string
}

func (g *GrpcSpec) FullPath() string {
	return path.Join(g.Service, g.Method)
}
