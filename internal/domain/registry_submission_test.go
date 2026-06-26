package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProblemManifestUnmarshalNormalizesTags(t *testing.T) {
	var manifest ProblemManifest
	err := json.Unmarshal([]byte(`{
		"provider": "leetcode",
		"external_id": "1",
		"slug": "two-sum",
		"title": "Two Sum",
		"difficulty": "easy",
		"tags": ["array", " hash-map ", "array", ""],
		"pattern_tags": ["hash-map", "two-pointer"],
		"starter_code": {"go": "package main"},
		"test_cases": [{"input": "1 2", "expected": "3", "is_hidden": true}],
		"version": 2
	}`), &manifest)
	require.NoError(t, err)
	require.Equal(t, []string{"array", "hash-map", "two-pointer"}, manifest.Tags)
	require.Equal(t, "two-sum", manifest.Slug)
	require.Equal(t, "package main", manifest.StarterCode["go"])
	require.Len(t, manifest.TestCases, 1)
}

func TestProblemManifestUnmarshalReturnsJSONError(t *testing.T) {
	var manifest ProblemManifest
	err := json.Unmarshal([]byte(`{`), &manifest)
	require.Error(t, err)
}

func TestNormalizeSubmissionLanguage(t *testing.T) {
	require.Equal(t, LanguageGo, NormalizeSubmissionLanguage(" Go "))
}

func TestIsSupportedSubmissionLanguage(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "uppercase supported", raw: "PYTHON", want: true},
		{name: "lowercase supported", raw: "typescript", want: true},
		{name: "unsupported", raw: "ruby", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, IsSupportedSubmissionLanguage(tc.raw))
		})
	}
}
