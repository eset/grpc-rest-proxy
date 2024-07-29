// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package transformer

import (
	jErrors "github.com/juju/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func GetRPCRequest(
	body []byte,
	requestDesc protoreflect.MessageDescriptor,
	params []Variable,
	httpBodyRule HTTPBodyRule,
) (*dynamicpb.Message, error) {
	protoRequest := dynamicpb.NewMessage(requestDesc)

	var err error
	if len(body) > 0 {
		params, err = processRequestBody(httpBodyRule, body, protoRequest, params)
		if err != nil {
			return nil, jErrors.Trace(err)
		}
	}

	err = setVariables(protoRequest, params)
	if err != nil {
		return nil, jErrors.Trace(err)
	}

	return protoRequest, nil
}

func processRequestBody(bodyRule HTTPBodyRule, body []byte, protoRequest proto.Message, params []Variable) ([]Variable, error) {
	switch bodyRule.RuleType {
	case NoBodyRule:
		return params, nil
	case MapRootRule:
		err := protojson.Unmarshal(body, protoRequest)
		if err != nil {
			return nil, jErrors.Trace(err)
		}
		return params, nil
	case FieldPathRule:
		params = append(params, Variable{FieldPath: bodyRule.FieldPath, Value: string(body)})
		return params, nil
	default:
		return nil, jErrors.New("unsupported body rules type")
	}
}
