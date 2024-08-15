// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package pattern

import "github.com/eset/grpc-rest-proxy/pkg/service/transformer"

type Matcher struct {
	ops  []Operation
	verb string
}

type MatchResult struct {
	Matched bool
	Vars    []transformer.Variable
}

func (m *Matcher) Match(path string) MatchResult {
	res := MatchResult{
		Matched: false,
	}

	path, verb := splitByVerb(path)
	if verb != m.verb {
		return res
	}

	segmentItr := newSegmentItr(path)

	matched := match(m.ops, &segmentItr, &res)
	res.Matched = matched

	// check if all segments are matched
	if segmentItr.hasNext() {
		res.Matched = false
	}
	return res
}

func match(ops []Operation, segmentItr *segmentItr, res *MatchResult) bool {
	for _, op := range ops {
		switch op.OpCode {
		case NoneOpCode:
			return false
		case MatchOpCode:
			if !segmentItr.hasNext() {
				return false
			}

			if op.Values[0] != segmentItr.next() {
				return false
			}
		case AnyOnceCode:
			if !segmentItr.hasNext() {
				return false
			}

			segmentItr.next()
		case StartCaptureCode:
			segmentItr.startCapture()
		case AnyZeroOrMoreCode:
			segmentItr.skipToEnd()
			continue
		case EndCaptureCode:
			variableValue := segmentItr.capture()
			res.Vars = append(res.Vars, transformer.Variable{
				FieldPath: op.Values,
				Value:     variableValue,
			})
			continue
		}
	}

	return true
}

func (m *Matcher) GetAllVariablePaths() [][]string {
	var paths [][]string
	for _, op := range m.ops {
		if op.OpCode == EndCaptureCode {
			paths = append(paths, op.Values)
		}
	}

	return paths
}
