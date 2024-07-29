// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package local

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	jErrors "github.com/juju/errors"
)

const (
	ProtoDescriptorExtension = ".desc"
	FileDescriptorsNotFound  = jErrors.ConstError("file descriptors in folder not found")
	PathIsNotDir             = jErrors.ConstError("path is not dir")
)

type Config struct {
	Dir string `mapstructure:"dir" validate:"required,dir"`
}

type Local struct {
	dir string
}

func New(cfg *Config) (*Local, error) {
	err := checkDir(cfg.Dir)
	if err != nil {
		return nil, jErrors.Trace(err)
	}

	return &Local{
		dir: cfg.Dir,
	}, nil
}

func checkDir(path string) error {
	pathInfo, err := os.Stat(path)
	if err != nil {
		return jErrors.Trace(err)
	}

	if !pathInfo.IsDir() {
		return jErrors.Trace(PathIsNotDir)
	}
	return nil
}

func readProtoDescriptorFile(file string) (*descriptorpb.FileDescriptorSet, error) {
	protoFile, err := os.ReadFile(file)
	if err != nil {
		return nil, jErrors.Trace(err)
	}

	fdSet := new(descriptorpb.FileDescriptorSet)
	if err = proto.Unmarshal(protoFile, fdSet); err != nil {
		return nil, jErrors.Trace(err)
	}

	return fdSet, nil
}

func (l *Local) GetProtoFileDescriptorSet(_ context.Context) ([]*descriptorpb.FileDescriptorSet, error) {
	var fdSets []*descriptorpb.FileDescriptorSet

	err := checkDir(l.dir)
	if err != nil {
		return nil, jErrors.Trace(err)
	}

	err = filepath.Walk(l.dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return jErrors.Trace(err)
		}
		if info.IsDir() || !strings.HasSuffix(path, ProtoDescriptorExtension) {
			return nil
		}

		fdSet, err := readProtoDescriptorFile(path)
		if err != nil {
			return jErrors.Trace(err)
		}
		fdSets = append(fdSets, fdSet)

		return nil
	})
	if err != nil {
		return nil, jErrors.Trace(err)
	}

	if len(fdSets) == 0 {
		return nil, jErrors.Trace(FileDescriptorsNotFound)
	}

	return fdSets, nil
}
