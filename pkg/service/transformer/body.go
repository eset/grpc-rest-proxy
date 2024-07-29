// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package transformer

import "strings"

type BodyRuleType int

const (
	NoBodyRule BodyRuleType = iota
	MapRootRule
	FieldPathRule
)

type HTTPBodyRule struct {
	RuleType  BodyRuleType
	FieldPath []string
}

func GetHTTPBodyRule(rule string) HTTPBodyRule {
	switch rule {
	case "":
		return HTTPBodyRule{RuleType: NoBodyRule}
	case "*":
		return HTTPBodyRule{RuleType: MapRootRule}
	default:
		fieldPath := strings.Split(rule, ".")
		return HTTPBodyRule{RuleType: FieldPathRule, FieldPath: fieldPath}
	}
}
