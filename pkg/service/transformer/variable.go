// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package transformer

import (
	"strconv"

	jErrors "github.com/juju/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

const errUnsupportedFieldType jErrors.ConstError = "unsupported field type"

type Variable struct {
	FieldPath []string
	Value     string
}

func ValidateFieldPath(desc protoreflect.MessageDescriptor, fieldPath []string) error {
	msg := dynamicpb.NewMessage(desc)

	_, fieldDesc, err := findInnerField(msg, fieldPath)
	if err != nil {
		return jErrors.Trace(err)
	}

	return jErrors.Trace(validateFieldType(fieldDesc))
}

func findFieldByName(protoDesc protoreflect.MessageDescriptor, fieldName protoreflect.Name) protoreflect.FieldDescriptor {
	field := protoDesc.Fields().ByName(fieldName)
	if field != nil {
		return field
	}

	oneofs := protoDesc.Oneofs()
	for idx := 0; idx < oneofs.Len(); idx++ {
		field := oneofs.Get(idx).Fields().ByName(fieldName)
		if field != nil {
			return field
		}
	}
	return nil
}

func setVariables(request *dynamicpb.Message, params []Variable) error {
	if len(params) == 0 {
		return nil
	}

	for _, param := range params {
		err := insertValueByPath(request, param.FieldPath, param.Value)
		if err != nil {
			return jErrors.Trace(err)
		}
	}

	return nil
}

func insertValueByPath(msg *dynamicpb.Message, fieldPath []string, value string) error {
	currentMsg, fieldDesc, err := findInnerField(msg, fieldPath)
	if err != nil {
		return jErrors.Trace(err)
	}

	return jErrors.Trace(setValueToField(currentMsg, fieldDesc, value))
}

func findInnerField(msg protoreflect.Message, fieldPath []string) (protoreflect.Message, protoreflect.FieldDescriptor, error) {
	if len(fieldPath) == 0 {
		return nil, nil, jErrors.New("field path is empty")
	}

	var currentMsg protoreflect.Message = msg
	// Iterate over all fields except the last one, which is the field to set.
	// Last field is processed separately because it requires a different handling.
	for _, field := range fieldPath[:len(fieldPath)-1] {
		fieldDescriptor := findFieldByName(currentMsg.Descriptor(), protoreflect.Name(field))
		if fieldDescriptor == nil {
			return nil, nil, jErrors.Errorf("field %s not found", field)
		}

		if fieldDescriptor.Kind() != protoreflect.MessageKind {
			return nil, nil, jErrors.New("field type must be message")
		}

		if fieldDescriptor.Cardinality() == protoreflect.Repeated {
			return nil, nil, jErrors.New("repeated fields are not supported")
		}

		innerMsg := currentMsg.Mutable(fieldDescriptor).Message()
		if innerMsg == nil {
			return nil, nil, jErrors.Errorf("field %s is nil", field)
		}

		currentMsg = innerMsg
	}

	lastField := fieldPath[len(fieldPath)-1]
	lastFieldDescriptor := findFieldByName(currentMsg.Descriptor(), protoreflect.Name(lastField))
	if lastFieldDescriptor == nil {
		return nil, nil, jErrors.Errorf("field %s not found", lastField)
	}

	return currentMsg, lastFieldDescriptor, nil
}

func setValueToField(msg protoreflect.Message, field protoreflect.FieldDescriptor, value string) error {
	fieldValue, err := valueOfFieldType(field, value)
	if err != nil {
		return jErrors.Trace(err)
	}

	if field.Cardinality() != protoreflect.Repeated {
		msg.Set(field, fieldValue)
		return nil
	}

	if field.IsList() {
		list := msg.Mutable(field).List()
		list.Append(fieldValue)
		return nil
	}

	return jErrors.New("only list or non-repeatable types are supported")
}

func valueOfFieldType(field protoreflect.FieldDescriptor, fieldValue string) (protoreflect.Value, error) { //nolint: gocyclo, funlen
	stringValue := protoreflect.ValueOfString(fieldValue)

	switch field.Kind() {
	case protoreflect.BoolKind:
		value, err := strconv.ParseBool(fieldValue)
		if err != nil {
			return stringValue, jErrors.Annotate(err, "parse boolean param")
		}
		return protoreflect.ValueOfBool(value), nil
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		value, err := strconv.ParseInt(fieldValue, 10, 32)
		if err != nil {
			return stringValue, jErrors.Annotate(err, "parse int32 param")
		}
		return protoreflect.ValueOfInt32(int32(value)), nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		value, err := strconv.ParseUint(fieldValue, 10, 32)
		if err != nil {
			return stringValue, jErrors.Annotate(err, "parse uint32 param")
		}
		return protoreflect.ValueOfUint32(uint32(value)), nil
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		value, err := strconv.ParseInt(fieldValue, 10, 64)
		if err != nil {
			return stringValue, jErrors.Annotate(err, "parse int64 param")
		}
		return protoreflect.ValueOfInt64(value), nil
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		value, err := strconv.ParseUint(fieldValue, 10, 64)
		if err != nil {
			return stringValue, jErrors.Annotate(err, "parse uint64 param")
		}
		return protoreflect.ValueOfUint64(value), nil
	case protoreflect.FloatKind:
		value, err := strconv.ParseFloat(fieldValue, 32)
		if err != nil {
			return stringValue, jErrors.Annotate(err, "parse float32 param")
		}
		return protoreflect.ValueOfFloat32(float32(value)), nil
	case protoreflect.DoubleKind:
		value, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return stringValue, jErrors.Annotate(err, "parse float64 param")
		}
		return protoreflect.ValueOfFloat64(value), nil
	case protoreflect.EnumKind:
		return protoreflect.ValueOfEnum(field.Enum().Values().ByName(protoreflect.Name(fieldValue)).Number()), nil
	case protoreflect.StringKind:
		return stringValue, nil
	case protoreflect.BytesKind:
		return protoreflect.ValueOfBytes([]byte(fieldValue)), nil
	case protoreflect.MessageKind:
		protoMsg := dynamicpb.NewMessage(field.Message())
		err := protojson.Unmarshal([]byte(fieldValue), proto.Message(protoMsg))
		if err != nil {
			return stringValue, jErrors.Annotate(err, "parse message param")
		}
		return protoreflect.ValueOfMessage(protoMsg), nil
	case protoreflect.GroupKind:
		return stringValue, errUnsupportedFieldType
	}

	return stringValue, errUnsupportedFieldType
}

func validateFieldType(field protoreflect.FieldDescriptor) error {
	switch field.Kind() {
	case protoreflect.BoolKind,
		protoreflect.Int32Kind,
		protoreflect.Sint32Kind,
		protoreflect.Sfixed32Kind,
		protoreflect.Uint32Kind,
		protoreflect.Fixed32Kind,
		protoreflect.Int64Kind,
		protoreflect.Sint64Kind,
		protoreflect.Sfixed64Kind,
		protoreflect.Uint64Kind,
		protoreflect.Fixed64Kind,
		protoreflect.FloatKind,
		protoreflect.DoubleKind,
		protoreflect.EnumKind,
		protoreflect.StringKind,
		protoreflect.BytesKind,
		protoreflect.MessageKind:
		return nil
	case protoreflect.GroupKind:
		return errUnsupportedFieldType
	default:
		return errUnsupportedFieldType
	}
}
