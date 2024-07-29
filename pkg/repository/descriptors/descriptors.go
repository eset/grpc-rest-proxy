// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package descriptors

import (
	"context"

	grpcClient "github.com/eset/grpc-rest-proxy/pkg/gateway/grpc"
	"github.com/eset/grpc-rest-proxy/pkg/repository/descriptors/local"
	"github.com/eset/grpc-rest-proxy/pkg/repository/descriptors/remote"

	jErrors "github.com/juju/errors"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	localType  = "local"
	remoteType = "remote"
)

type Config struct {
	Local  *local.Config  `mapstructure:"local"`
	Remote *remote.Config `mapstructure:"remote"`
	Kind   string         `mapstructure:"kind" validate:"required,oneof=local remote"`
}

type Descriptors interface {
	GetProtoFileDescriptorSet(ctx context.Context) ([]*descriptorpb.FileDescriptorSet, error)
}

func New(cfg *Config, client grpcClient.ClientInterface) (Descriptors, error) {
	if cfg.Local != nil && cfg.Kind == localType {
		return local.New(cfg.Local)
	} else if cfg.Remote != nil && cfg.Kind == remoteType {
		return remote.New(cfg.Remote, client)
	}
	return nil, jErrors.Errorf("Undefined type of descriptors repository in config.")
}
