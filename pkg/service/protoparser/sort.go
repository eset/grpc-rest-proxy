// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package protoparser

import (
	"slices"

	"google.golang.org/protobuf/types/descriptorpb"
)

// Sorts filedescriptors by their dependencies so that they are in correct order for further processing.
func SortByDependencies(fdSets []*descriptorpb.FileDescriptorSet) []*descriptorpb.FileDescriptorProto {
	var sortedDescs []*descriptorpb.FileDescriptorProto

	for _, fdSet := range fdSets {
		for _, fd := range fdSet.GetFile() {
			processDependencies(fdSets, fd, &sortedDescs)
		}
	}

	return sortedDescs
}

func processDependencies(
	fdSets []*descriptorpb.FileDescriptorSet,
	currentFileDesc *descriptorpb.FileDescriptorProto,
	result *[]*descriptorpb.FileDescriptorProto,
) {
	for _, dependency := range currentFileDesc.GetDependency() {
		dependencyFileDesc := findFileSet(fdSets, dependency)
		if dependencyFileDesc == nil {
			// This dependency does not exist in our sets. Skip it.
			continue
		}

		contains := slices.ContainsFunc(*result, func(resFd *descriptorpb.FileDescriptorProto) bool {
			return resFd.GetName() == dependency
		})

		// This dependency was not processed, yet.
		if !contains {
			processDependencies(fdSets, dependencyFileDesc, result)
		}
	}

	*result = append(*result, currentFileDesc)
}

func findFileSet(fdSets []*descriptorpb.FileDescriptorSet, path string) *descriptorpb.FileDescriptorProto {
	for _, fdSet := range fdSets {
		for _, fd := range fdSet.GetFile() {
			if path == fd.GetName() {
				return fd
			}
		}
	}

	return nil
}
