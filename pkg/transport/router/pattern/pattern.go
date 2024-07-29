// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package pattern

import "strings"

type OpType int

const (
	NoneOpCode OpType = iota
	MatchOpCode
	AnyOnceCode
	AnyZeroOrMoreCode
	StartCaptureCode
	EndCaptureCode
)

type Operation struct {
	OpCode OpType
	Values []string
}

type CaptureVariable struct {
	FieldPath []string
	Ops       []Operation
}

func splitByVerb(path string) (pattern string, verb string) {
	verbIdx := strings.IndexRune(path, ':')
	if verbIdx == -1 {
		return path, ""
	}

	return path[:verbIdx], path[verbIdx+1:]
}
