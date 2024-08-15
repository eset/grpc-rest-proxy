// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package protoparser

import (
	"errors"
	"strings"

	"github.com/eset/grpc-rest-proxy/pkg/service/router"

	jErrors "github.com/juju/errors"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ParseFileDescSets(fdSets []*descriptorpb.FileDescriptorSet) ParseResult {
	result := ParseResult{
		FileRegistry: &protoregistry.Files{},
		TypeResolver: &protoregistry.Types{},
	}

	// Register default types.
	registerFile(anypb.File_google_protobuf_any_proto, &result)
	registerFile(timestamppb.File_google_protobuf_timestamp_proto, &result)

	fileDescriptors := SortByDependencies(fdSets)

	for _, fileDesc := range fileDescriptors {
		parseFileDescSet(fileDesc, &result)
	}

	return result
}

func parseFileDescSet(file *descriptorpb.FileDescriptorProto, result *ParseResult) {
	fd, err := protodesc.NewFile(file, result.FileRegistry)
	if err != nil {
		result.AddError(jErrors.Annotatef(err, "error while parsing %s", file.GetName()))
		return
	}

	registerFile(fd, result)
	ParseFileDesc(fd, result)
}

func registerFile(fd protoreflect.FileDescriptor, result *ParseResult) {
	_, err := result.FileRegistry.FindFileByPath(fd.Path())
	if err == nil {
		return
	}

	if errors.Is(err, protoregistry.NotFound) {
		err = result.FileRegistry.RegisterFile(fd)
		if err != nil {
			result.AddError(jErrors.Trace(err))
		}
		return
	}

	result.AddError(jErrors.Trace(err))
}

func ParseFileDesc(fd protoreflect.FileDescriptor, result *ParseResult) {
	err := registerTypes(fd, result)
	if err != nil {
		result.AddError(jErrors.Trace(err))
		return
	}

	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		parseServiceDesc(services.Get(i), result)
	}
}

func ParseServiceNameAndMethod(fullname string) (service string, method string, err error) {
	// add leading slash
	if fullname != "" && fullname[0] != '/' {
		fullname = "/" + fullname
	}

	// check position of method name
	methodPos := strings.LastIndex(fullname, ".")
	if methodPos < 0 || (methodPos >= len(fullname)) {
		return service, method, jErrors.Errorf("error while parsing service name and method from %s", fullname)
	}

	service = fullname[:methodPos]
	method = fullname[methodPos+1:]

	return service, method, nil
}

func getAdditionalBindings(httpOption *annotations.HttpRule) []*annotations.HttpRule {
	var bindings []*annotations.HttpRule
	additionalBindings := httpOption.GetAdditionalBindings()

	if len(additionalBindings) == 0 {
		return bindings
	}

	bindings = append(bindings, additionalBindings...)
	for _, binding := range additionalBindings {
		bindings = append(bindings, getAdditionalBindings(binding)...)
	}

	return bindings
}

func createRoute(rule *annotations.HttpRule, fullname string, method protoreflect.MethodDescriptor) (*router.Route, error) {
	methodType, pattern, err := getPattern(rule)
	if err != nil {
		return nil, jErrors.Trace(err)
	}

	rpcService, rpcMethod, err := ParseServiceNameAndMethod(fullname)
	if err != nil {
		return nil, jErrors.Trace(err)
	}

	route := router.NewRoute(pattern, rule.GetBody(), methodType, &router.GrpcSpec{
		RequestDesc:  method.Input(),
		ResponseDesc: method.Output(),
		Service:      rpcService,
		Method:       rpcMethod,
	})
	return route, nil
}

func parseServiceDesc(service protoreflect.ServiceDescriptor, result *ParseResult) {
	methods := service.Methods()

	for m := 0; m < methods.Len(); m++ {
		method := methods.Get(m)
		fullname := string(method.FullName())
		methodOpts, ok := method.Options().(*descriptorpb.MethodOptions)
		if !ok {
			result.AddError(jErrors.Trace(jErrors.New("cannot convert method options to Method Options")))
			continue
		}

		if !proto.HasExtension(methodOpts, annotations.E_Http) {
			continue
		}

		httpOption, ok := proto.GetExtension(methodOpts, annotations.E_Http).(*annotations.HttpRule)
		if !ok {
			result.AddError(jErrors.Trace(jErrors.New("cannot convert extension to HttpRule")))
			continue
		}

		httpRules := []*annotations.HttpRule{httpOption}

		additionalBindings := getAdditionalBindings(httpOption)
		httpRules = append(httpRules, additionalBindings...)

		for _, rule := range httpRules {
			route, err := createRoute(rule, fullname, method)
			if err != nil {
				result.AddError(jErrors.Trace(err))
				continue
			}
			result.AddRoute(route)
		}
	}
}

func getPattern(rule *annotations.HttpRule) (router.MethodType, string, error) {
	switch pattern := rule.GetPattern().(type) {
	case *annotations.HttpRule_Get:
		return router.GET, pattern.Get, nil
	case *annotations.HttpRule_Put:
		return router.PUT, pattern.Put, nil
	case *annotations.HttpRule_Post:
		return router.POST, pattern.Post, nil
	case *annotations.HttpRule_Delete:
		return router.DELETE, pattern.Delete, nil
	case *annotations.HttpRule_Patch:
		return router.PATCH, pattern.Patch, nil
	case *annotations.HttpRule_Custom:
		return router.UnknownMethod, "", jErrors.Errorf("unsupported custom rule: %s", pattern.Custom.GetKind())
	}

	return router.UnknownMethod, "", jErrors.Errorf("unknown method")
}

func registerTypes(fd protoreflect.FileDescriptor, result *ParseResult) error {
	msgs := fd.Messages()
	for idx := 0; idx < msgs.Len(); idx++ {
		msg := msgs.Get(idx)
		_, err := result.TypeResolver.FindMessageByName(msg.FullName())
		if err == nil {
			continue
		}

		err = result.TypeResolver.RegisterMessage(dynamicpb.NewMessageType(msg))
		if err != nil {
			return jErrors.Trace(err)
		}
	}

	enums := fd.Enums()
	for idx := 0; idx < enums.Len(); idx++ {
		enum := enums.Get(idx)
		_, err := result.TypeResolver.FindEnumByName(enum.FullName())
		if err == nil {
			continue
		}

		err = result.TypeResolver.RegisterEnum(dynamicpb.NewEnumType(enum))
		if err != nil {
			return jErrors.Trace(err)
		}
	}

	exts := fd.Extensions()
	for idx := 0; idx < exts.Len(); idx++ {
		ext := exts.Get(idx)
		_, err := result.TypeResolver.FindExtensionByName(ext.FullName())
		if err == nil {
			continue
		}

		err = result.TypeResolver.RegisterExtension(dynamicpb.NewExtensionType(ext))
		if err != nil {
			return jErrors.Trace(err)
		}
	}

	return nil
}
