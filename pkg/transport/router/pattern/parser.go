// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package pattern

import (
	"fmt"
	"strings"

	jErrors "github.com/juju/errors"
)

func Parse(pattern string) (*Matcher, error) {
	if pattern == "" {
		return nil, jErrors.Trace(fmt.Errorf("empty pattern"))
	}

	if pattern[0] != '/' {
		return nil, jErrors.Trace(fmt.Errorf("pattern must start with '/'"))
	}

	pattern, verb := splitByVerb(pattern)
	ops, err := parseSegmentsItr(pattern, false)
	if err != nil {
		return nil, jErrors.Trace(err)
	}

	err = validateOps(ops)
	if err != nil {
		return nil, jErrors.Trace(err)
	}

	return &Matcher{
		ops:  ops,
		verb: verb,
	}, nil
}

func getFieldPath(fieldPath string) ([]string, error) {
	parsedFieldPath := strings.Split(fieldPath, ".")

	for _, field := range parsedFieldPath {
		err := validateLiteral(field)
		if err != nil {
			return nil, jErrors.Trace(err)
		}
	}

	return parsedFieldPath, nil
}

func parseVariable(variable string) ([]string, []Operation, error) {
	var rawFieldPath = variable
	var varPath string

	if assignIdx := strings.IndexRune(variable, '='); assignIdx != -1 {
		rawFieldPath = variable[:assignIdx]
		varPath = variable[assignIdx+1:]
	}

	if rawFieldPath == "" {
		return nil, nil, fmt.Errorf("empty field path in variable %s", variable)
	}

	pathOps, err := parseVariablePattern(varPath)
	if err != nil {
		return nil, nil, jErrors.Trace(err)
	}

	parsedFieldPath, err := getFieldPath(rawFieldPath)
	if err != nil {
		return nil, nil, jErrors.Trace(err)
	}

	return parsedFieldPath, pathOps, nil
}

func parseVariablePattern(pattern string) ([]Operation, error) {
	if pattern == "" {
		// corresponds to {var=*} or {var}
		return []Operation{{OpCode: AnyOnceCode}}, nil
	}

	segments, err := parseSegmentsItr(pattern, true)
	if err != nil {
		return nil, jErrors.Trace(err)
	}

	return segments, nil
}

func parseSegmentsItr(pattern string, inVariable bool) ([]Operation, error) {
	var ops []Operation

	itr := newParserSegmentItr(pattern)
	for itr.hasNext() {
		segment := itr.next()
		switch segment {
		case "":
			if !itr.isFirst() {
				return nil, fmt.Errorf("empty segment")
			}
			continue
		case "*":
			ops = append(ops, Operation{OpCode: AnyOnceCode})
			continue
		case "**":
			ops = append(ops, Operation{OpCode: AnyZeroOrMoreCode})
			continue
		}

		// check if segment is a variable
		if segment[0] == '{' {
			if inVariable {
				return nil, fmt.Errorf("unexpected '{'")
			}

			if segment[len(segment)-1] != '}' {
				return nil, fmt.Errorf("variable must end with '}'")
			}

			variableSeg := segment[1 : len(segment)-1]
			fieldPath, varOps, err := parseVariable(variableSeg)
			if err != nil {
				return nil, jErrors.Trace(err)
			}
			ops = append(ops, Operation{OpCode: StartCaptureCode})
			ops = append(ops, varOps...)
			ops = append(ops, Operation{OpCode: EndCaptureCode, Values: fieldPath})
			continue
		}

		err := validateLiteral(segment)
		if err != nil {
			return nil, jErrors.Trace(err)
		}

		ops = append(ops, Operation{OpCode: MatchOpCode, Values: []string{segment}})
	}

	return ops, nil
}

func validateLiteral(literal string) error {
	if literal == "" {
		return fmt.Errorf("empty literal")
	}

	if strings.ContainsAny(literal, "{}*=") {
		return fmt.Errorf("literal %s cannot contain '{', '}' or '*'", literal)
	}

	return nil
}

func validateOps(ops []Operation) error {
	hasAnyZeroOrMore := false

	for _, op := range ops {
		if hasAnyZeroOrMore && op.OpCode != EndCaptureCode {
			return fmt.Errorf("segment '**' must be last segment")
		}

		if op.OpCode == MatchOpCode && len(op.Values) != 1 {
			return fmt.Errorf("match operation must have exactly one value")
		}

		if AnyZeroOrMoreCode == op.OpCode {
			hasAnyZeroOrMore = true
		}
	}

	return nil
}
