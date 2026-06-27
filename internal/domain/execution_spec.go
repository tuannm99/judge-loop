package domain

import (
	"encoding/json"
	"fmt"
	"strings"
)

var supportedComparatorKinds = map[string]struct{}{
	"":                        {},
	"exact":                   {},
	"unordered_array":         {},
	"unordered_nested_array":  {},
	"float_epsilon":           {},
	"any_of":                  {},
	"json_subset":             {},
	"valid_topological_order": {},
}

func ValidateJudgeReadyManifest(manifest ProblemManifest) error {
	if !manifest.JudgeReady {
		return nil
	}
	if len(manifest.TestCases) == 0 {
		return fmt.Errorf("%s: judge-ready problem requires test cases", manifest.Slug)
	}
	if _, ok := supportedComparatorKinds[manifest.ExecutionSpec.Comparator.Kind]; !ok {
		return fmt.Errorf(
			"%s: unsupported comparator %q",
			manifest.Slug,
			manifest.ExecutionSpec.Comparator.Kind,
		)
	}
	if err := validateExecutionSpec(manifest.ExecutionSpec); err != nil {
		return fmt.Errorf("%s: %w", manifest.Slug, err)
	}
	for index, testCase := range manifest.TestCases {
		if err := validateManifestTestCase(manifest.ExecutionSpec, testCase); err != nil {
			return fmt.Errorf("%s: test_cases[%d]: %w", manifest.Slug, index, err)
		}
	}
	languages := manifest.ExecutionSpec.SupportedLanguages
	if len(languages) == 0 {
		for _, language := range supportedSubmissionLanguages {
			if strings.TrimSpace(manifest.StarterCode[string(language)]) != "" {
				languages = append(languages, language)
			}
		}
	}
	if len(languages) == 0 {
		return fmt.Errorf("%s: at least one starter language is required", manifest.Slug)
	}
	for _, language := range languages {
		if !IsSupportedSubmissionLanguage(string(language)) {
			return fmt.Errorf("%s: unsupported language %q", manifest.Slug, language)
		}
		if strings.TrimSpace(manifest.StarterCode[string(language)]) == "" {
			return fmt.Errorf("%s: missing %s starter code", manifest.Slug, language)
		}
	}
	return nil
}

func validateExecutionSpec(spec ExecutionSpec) error {
	switch spec.Mode {
	case "", ExecutionModeStdin:
		return nil
	case ExecutionModeFunction, ExecutionModeInPlace:
		if strings.TrimSpace(spec.Entrypoint) == "" {
			return fmt.Errorf("entrypoint is required")
		}
		for _, param := range spec.Signature.Params {
			if !supportedExecutionType(param.Type, false) {
				return fmt.Errorf("unsupported parameter type %q", param.Type)
			}
		}
		if !supportedExecutionType(spec.Signature.Returns, true) {
			return fmt.Errorf("unsupported return type %q", spec.Signature.Returns)
		}
		if spec.Mode == ExecutionModeInPlace {
			if spec.Output.ParamIndex < 0 ||
				spec.Output.ParamIndex >= len(spec.Signature.Params) {
				return fmt.Errorf("invalid in-place output param index")
			}
		}
		return nil
	case ExecutionModeClass:
		if strings.TrimSpace(spec.ClassName) == "" {
			return fmt.Errorf("class_name is required")
		}
		for _, param := range spec.Constructor.Params {
			if !supportedExecutionType(param.Type, false) {
				return fmt.Errorf("unsupported constructor type %q", param.Type)
			}
		}
		if len(spec.Methods) == 0 {
			return fmt.Errorf("class methods are required")
		}
		for name, method := range spec.Methods {
			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("class method name is required")
			}
			for _, param := range method.Params {
				if !supportedExecutionType(param.Type, false) {
					return fmt.Errorf("method %s has unsupported type %q", name, param.Type)
				}
			}
			if !supportedExecutionType(method.Returns, true) {
				return fmt.Errorf("method %s has unsupported return type %q", name, method.Returns)
			}
		}
		return nil
	case ExecutionModeInteractive, ExecutionModeCustom:
		return fmt.Errorf("%s mode is not locally judge-ready", spec.Mode)
	default:
		return fmt.Errorf("unsupported execution mode %q", spec.Mode)
	}
}

func validateManifestTestCase(spec ExecutionSpec, testCase TestCaseManifest) error {
	if spec.Mode == "" || spec.Mode == ExecutionModeStdin {
		if testCase.Input == "" {
			return fmt.Errorf("input is required")
		}
		if testCase.Expected == "" {
			return fmt.Errorf("expected is required")
		}
		return nil
	}
	if len(testCase.InputJSON) == 0 || !json.Valid(testCase.InputJSON) {
		return fmt.Errorf("valid input_json is required")
	}
	if len(testCase.ExpectedJSON) == 0 || !json.Valid(testCase.ExpectedJSON) {
		return fmt.Errorf("valid expected_json is required")
	}
	return nil
}

func supportedExecutionType(raw string, allowVoid bool) bool {
	value := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(raw), " ", ""))
	for strings.HasSuffix(value, "[]") {
		value = strings.TrimSuffix(value, "[]")
	}
	switch value {
	case "int", "int64", "long", "float", "double", "bool", "string",
		"listnode", "listnode?", "optional[listnode]",
		"treenode", "treenode?", "optional[treenode]",
		"graphnode", "graphnode?", "optional[graphnode]":
		return true
	case "void", "none", "nonetype":
		return allowVoid
	default:
		return false
	}
}
