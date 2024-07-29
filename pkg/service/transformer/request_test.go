// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package transformer_test

import (
	"testing"

	userpb "github.com/eset/grpc-rest-proxy/cmd/examples/grpcserver/gen/user/v1"
	"github.com/eset/grpc-rest-proxy/pkg/service/transformer"

	"github.com/stretchr/testify/require"

	"google.golang.org/protobuf/encoding/protojson"
)

func TestBasicRequestTransform(t *testing.T) {
	msgGetUser := &userpb.GetUserRequest{}
	msgGetUserDesc := msgGetUser.ProtoReflect().Descriptor()
	msgGetUsersResp := &userpb.GetUsersResponse{}
	msgGetUsersRespDesc := msgGetUsersResp.ProtoReflect().Descriptor()
	msgDeleteUser := &userpb.DeleteUserResponse{}
	msgDeleteUserDesc := msgDeleteUser.ProtoReflect().Descriptor()

	request, err := transformer.GetRPCRequest(nil, msgGetUserDesc, []transformer.Variable{
		{FieldPath: []string{"username"}, Value: "John"},
		{FieldPath: []string{"country"}, Value: "USA"},
	}, transformer.HTTPBodyRule{RuleType: transformer.NoBodyRule})
	require.NoError(t, err)
	_, err = protojson.Marshal(request)
	require.NoError(t, err)

	_, err = transformer.GetRPCRequest(nil, msgGetUserDesc, []transformer.Variable{
		{FieldPath: []string{"username2"}, Value: "1"},
	}, transformer.HTTPBodyRule{RuleType: transformer.NoBodyRule})
	require.Error(t, err, "field username2 not found in message user.GetUserRequest")

	_, err = transformer.GetRPCRequest(nil, msgDeleteUserDesc, []transformer.Variable{
		{FieldPath: []string{"id"}, Value: "not_number"},
	}, transformer.HTTPBodyRule{RuleType: transformer.NoBodyRule})
	require.Error(t, err, "field id is not a number")

	_, err = transformer.GetRPCRequest(nil, msgGetUsersRespDesc, []transformer.Variable{
		{FieldPath: []string{"users", "value"}, Value: "id"},
	}, transformer.HTTPBodyRule{RuleType: transformer.NoBodyRule})
	require.Error(t, err, "field users is repeatable")

	_, err = transformer.GetRPCRequest([]byte("{\"username\":\"John\"}"), msgGetUserDesc, []transformer.Variable{},
		transformer.HTTPBodyRule{RuleType: transformer.MapRootRule})
	require.NoError(t, err)

	_, err = transformer.GetRPCRequest([]byte("{\"test\":1}"), msgGetUserDesc, []transformer.Variable{},
		transformer.HTTPBodyRule{RuleType: transformer.MapRootRule})
	require.Error(t, err, "field test not exist")

	_, err = transformer.GetRPCRequest([]byte("John"), msgGetUserDesc, []transformer.Variable{},
		transformer.HTTPBodyRule{RuleType: transformer.FieldPathRule, FieldPath: []string{"username"}})
	require.NoError(t, err)

	_, err = transformer.GetRPCRequest([]byte("1"), msgGetUserDesc, []transformer.Variable{},
		transformer.HTTPBodyRule{RuleType: transformer.FieldPathRule, FieldPath: []string{"test"}})
	require.Error(t, err, "field test not exist")
}

func TestRepeatableTransform(t *testing.T) {
	msg := userpb.Summary{}
	msgDesc := msg.ProtoReflect().Descriptor()

	request, err := transformer.GetRPCRequest(nil, msgDesc, []transformer.Variable{
		{FieldPath: []string{"usernames"}, Value: "John"},
		{FieldPath: []string{"usernames"}, Value: "Diego"},
		{FieldPath: []string{"usernames"}, Value: "Alberto"},
	}, transformer.HTTPBodyRule{RuleType: transformer.NoBodyRule})
	require.NoError(t, err)
	_, err = protojson.Marshal(request)
	require.NoError(t, err)

	_, err = transformer.GetRPCRequest([]byte("usernames"), msgDesc, []transformer.Variable{},
		transformer.HTTPBodyRule{RuleType: transformer.FieldPathRule, FieldPath: []string{"usernames"}})
	require.NoError(t, err)
}
