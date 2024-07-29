// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package pattern_test

import (
	"slices"
	"testing"

	"github.com/eset/grpc-rest-proxy/pkg/service/transformer"
	"github.com/eset/grpc-rest-proxy/pkg/transport/router/pattern"

	"github.com/stretchr/testify/require"
)

func TestPatternsMatching(t *testing.T) {
	type patternTest struct {
		path          string
		matched       bool
		matchedParams []transformer.Variable
	}

	patterns := []struct {
		pattern string
		valid   bool
		tests   []patternTest
	}{
		{
			pattern: "/v1/package/{id/other",
			valid:   false,
		},
		{
			pattern: "/api/v1/users/{user.id=/*/}/posts/{postid=**}",
			valid:   true,
			tests: []patternTest{
				{
					path:    "/api/v1/users/1/posts/abcde/other",
					matched: true,
					matchedParams: []transformer.Variable{
						{FieldPath: []string{"user", "id"}, Value: "1"},
						{FieldPath: []string{"postid"}, Value: "abcde/other"},
					},
				},
				{path: "/api/v1/users/1/posts/", matched: true},
			},
		},
		{
			pattern: "/api/v1/users/{user.id=*/id/*}",
			valid:   true,
			tests: []patternTest{
				{path: "/api/v1/users/1/id/2", matched: true},
				{path: "/api/v1/users/1/id/", matched: false},
				{path: "/api/v1/users//id/2", matched: true},
			},
		},
		{
			pattern: "/api/v1/users",
			valid:   true,
			tests: []patternTest{
				{path: "/api/v1/users", matched: true},
				{path: "/api/v1/users:foo", matched: false},
				{path: "/api/v2/users", matched: false},
			},
		},
		{pattern: "/",
			valid: true,
			tests: []patternTest{
				{path: "/", matched: true},
				{path: "", matched: true},
				{path: "/api", matched: false},
			},
		},
		{pattern: "/{inner.id=/*/{id=*}}/c/d", valid: false},               // inner variable not allowed
		{pattern: "/{id=/*/}}/c/d", valid: false},                          // unmatched brackets
		{pattern: "/{=*}/c/d", valid: false},                               // empty field path
		{pattern: "/api/p*rt/s{}", valid: false},                           // invalid characters
		{pattern: "api/v1", valid: false},                                  // missing leading slash
		{pattern: "/api/users/**/{user.id=*}/posts", valid: false},         // ** is not at the end
		{pattern: "/api/users/{user.id=**}/posts:deleteAll", valid: false}, // ** is not at the end
	}

	for _, p := range patterns {
		t.Logf("pattern: %s", p.pattern)
		matcher, err := pattern.Parse(p.pattern)
		if p.valid {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			continue
		}

		for _, test := range p.tests {
			t.Logf("\ttest path: %s", test.path)
			match := matcher.Match(test.path)
			require.Equal(t, test.matched, match.Matched)

			for _, param := range test.matchedParams {
				var found bool
				for _, v := range match.Vars {
					if slices.Equal(v.FieldPath, param.FieldPath) {
						require.Equal(t, param.Value, v.Value)
						found = true
						break
					}
				}
				require.True(t, found)
			}
		}
	}
}
