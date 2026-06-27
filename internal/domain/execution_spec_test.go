package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateJudgeReadyManifest(t *testing.T) {
	stdinManifest := func() ProblemManifest {
		return ProblemManifest{
			Slug:       "stdin",
			JudgeReady: true,
			StarterCode: map[string]string{
				string(LanguageGo): "package main",
			},
			TestCases: []TestCaseManifest{{Input: "1", Expected: "2"}},
		}
	}
	functionManifest := func() ProblemManifest {
		return ProblemManifest{
			Slug:       "function",
			JudgeReady: true,
			StarterCode: map[string]string{
				string(LanguageGo): "func solve(value int) int",
			},
			ExecutionSpec: ExecutionSpec{
				Mode:       ExecutionModeFunction,
				Entrypoint: "solve",
				Signature: ExecutionSignature{
					Params:  []ExecutionParam{{Name: "value", Type: "int"}},
					Returns: "int",
				},
				SupportedLanguages: []Language{LanguageGo},
			},
			TestCases: []TestCaseManifest{{
				InputJSON:    json.RawMessage(`{"args":[1]}`),
				ExpectedJSON: json.RawMessage(`2`),
			}},
		}
	}

	t.Run("non judge-ready manifest bypasses validation", func(t *testing.T) {
		require.NoError(t, ValidateJudgeReadyManifest(ProblemManifest{}))
	})
	t.Run("valid stdin with inferred starter language", func(t *testing.T) {
		require.NoError(t, ValidateJudgeReadyManifest(stdinManifest()))
	})
	t.Run("valid function", func(t *testing.T) {
		require.NoError(t, ValidateJudgeReadyManifest(functionManifest()))
	})
	t.Run("valid in-place", func(t *testing.T) {
		manifest := functionManifest()
		manifest.ExecutionSpec.Mode = ExecutionModeInPlace
		manifest.ExecutionSpec.Output.ParamIndex = 0
		require.NoError(t, ValidateJudgeReadyManifest(manifest))
	})
	t.Run("valid class", func(t *testing.T) {
		manifest := functionManifest()
		manifest.ExecutionSpec = ExecutionSpec{
			Mode:      ExecutionModeClass,
			ClassName: "Counter",
			Constructor: ExecutionSignature{
				Params: []ExecutionParam{{Name: "start", Type: "int64"}},
			},
			Methods: map[string]MethodSpec{
				"add": {
					Params:  []ExecutionParam{{Name: "value", Type: "int[]"}},
					Returns: "void",
				},
			},
			SupportedLanguages: []Language{LanguageGo},
		}
		require.NoError(t, ValidateJudgeReadyManifest(manifest))
	})

	tests := []struct {
		name   string
		mutate func(*ProblemManifest)
		error  string
	}{
		{
			name: "requires cases",
			mutate: func(m *ProblemManifest) {
				m.TestCases = nil
			},
			error: "requires test cases",
		},
		{
			name: "rejects comparator",
			mutate: func(m *ProblemManifest) {
				m.ExecutionSpec.Comparator.Kind = "unknown"
			},
			error: "unsupported comparator",
		},
		{
			name: "requires function entrypoint",
			mutate: func(m *ProblemManifest) {
				m.ExecutionSpec.Entrypoint = " "
			},
			error: "entrypoint is required",
		},
		{
			name: "rejects parameter type",
			mutate: func(m *ProblemManifest) {
				m.ExecutionSpec.Signature.Params[0].Type = "map"
			},
			error: `unsupported parameter type "map"`,
		},
		{
			name: "rejects return type",
			mutate: func(m *ProblemManifest) {
				m.ExecutionSpec.Signature.Returns = "map"
			},
			error: `unsupported return type "map"`,
		},
		{
			name: "rejects in-place output index",
			mutate: func(m *ProblemManifest) {
				m.ExecutionSpec.Mode = ExecutionModeInPlace
				m.ExecutionSpec.Output.ParamIndex = 2
			},
			error: "invalid in-place output param index",
		},
		{
			name: "requires class name",
			mutate: func(m *ProblemManifest) {
				m.ExecutionSpec = ExecutionSpec{Mode: ExecutionModeClass}
			},
			error: "class_name is required",
		},
		{
			name: "rejects constructor type",
			mutate: func(m *ProblemManifest) {
				m.ExecutionSpec = validClassSpec()
				m.ExecutionSpec.Constructor.Params[0].Type = "map"
			},
			error: `unsupported constructor type "map"`,
		},
		{
			name: "requires class methods",
			mutate: func(m *ProblemManifest) {
				m.ExecutionSpec = validClassSpec()
				m.ExecutionSpec.Methods = nil
			},
			error: "class methods are required",
		},
		{
			name: "requires class method name",
			mutate: func(m *ProblemManifest) {
				m.ExecutionSpec = validClassSpec()
				m.ExecutionSpec.Methods = map[string]MethodSpec{" ": {Returns: "int"}}
			},
			error: "class method name is required",
		},
		{
			name: "rejects class method parameter",
			mutate: func(m *ProblemManifest) {
				m.ExecutionSpec = validClassSpec()
				m.ExecutionSpec.Methods["add"] = MethodSpec{
					Params:  []ExecutionParam{{Type: "map"}},
					Returns: "int",
				}
			},
			error: `method add has unsupported type "map"`,
		},
		{
			name: "rejects class method return",
			mutate: func(m *ProblemManifest) {
				m.ExecutionSpec = validClassSpec()
				m.ExecutionSpec.Methods["add"] = MethodSpec{Returns: "map"}
			},
			error: `method add has unsupported return type "map"`,
		},
		{
			name: "rejects interactive mode",
			mutate: func(m *ProblemManifest) {
				m.ExecutionSpec.Mode = ExecutionModeInteractive
			},
			error: "interactive mode is not locally judge-ready",
		},
		{
			name: "rejects custom mode",
			mutate: func(m *ProblemManifest) {
				m.ExecutionSpec.Mode = ExecutionModeCustom
			},
			error: "custom mode is not locally judge-ready",
		},
		{
			name: "rejects unknown mode",
			mutate: func(m *ProblemManifest) {
				m.ExecutionSpec.Mode = ExecutionMode("unknown")
			},
			error: `unsupported execution mode "unknown"`,
		},
		{
			name: "requires valid structured input",
			mutate: func(m *ProblemManifest) {
				m.TestCases[0].InputJSON = json.RawMessage(`{`)
			},
			error: "valid input_json is required",
		},
		{
			name: "requires valid structured expected output",
			mutate: func(m *ProblemManifest) {
				m.TestCases[0].ExpectedJSON = nil
			},
			error: "valid expected_json is required",
		},
		{
			name: "requires starter language",
			mutate: func(m *ProblemManifest) {
				m.ExecutionSpec.SupportedLanguages = nil
				m.StarterCode = nil
			},
			error: "at least one starter language is required",
		},
		{
			name: "rejects unsupported language",
			mutate: func(m *ProblemManifest) {
				m.ExecutionSpec.SupportedLanguages = []Language{"ruby"}
				m.StarterCode["ruby"] = "def solve"
			},
			error: `unsupported language "ruby"`,
		},
		{
			name: "requires starter code for declared language",
			mutate: func(m *ProblemManifest) {
				m.StarterCode[string(LanguageGo)] = " "
			},
			error: "missing go starter code",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := functionManifest()
			test.mutate(&manifest)
			require.ErrorContains(t, ValidateJudgeReadyManifest(manifest), test.error)
		})
	}

	t.Run("stdin requires input", func(t *testing.T) {
		manifest := stdinManifest()
		manifest.TestCases[0].Input = ""
		require.ErrorContains(t, ValidateJudgeReadyManifest(manifest), "input is required")
	})
	t.Run("stdin requires expected output", func(t *testing.T) {
		manifest := stdinManifest()
		manifest.TestCases[0].Expected = ""
		require.ErrorContains(t, ValidateJudgeReadyManifest(manifest), "expected is required")
	})
}

func TestSupportedExecutionType(t *testing.T) {
	for _, value := range []string{
		"int", "long", "double", "bool", "string", "int[][]",
		"ListNode?", "Optional[TreeNode]", "graphnode[]",
	} {
		require.True(t, supportedExecutionType(value, false), value)
	}
	require.True(t, supportedExecutionType("void", true))
	require.False(t, supportedExecutionType("void", false))
	require.False(t, supportedExecutionType("map[string]int", true))
}

func validClassSpec() ExecutionSpec {
	return ExecutionSpec{
		Mode:      ExecutionModeClass,
		ClassName: "Counter",
		Constructor: ExecutionSignature{
			Params: []ExecutionParam{{Name: "start", Type: "int"}},
		},
		Methods: map[string]MethodSpec{
			"add": {
				Params:  []ExecutionParam{{Name: "value", Type: "int"}},
				Returns: "int",
			},
		},
		SupportedLanguages: []Language{LanguageGo},
	}
}
