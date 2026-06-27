package application

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/tuannm99/judge-loop/internal/domain"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

var identifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type harnessInput struct {
	Args        []json.RawMessage `json:"args"`
	Constructor []json.RawMessage `json:"constructor"`
	Calls       []harnessCall     `json:"calls"`
}

type harnessCall struct {
	Method string            `json:"method"`
	Args   []json.RawMessage `json:"args"`
}

func renderExecutionHarness(
	sub *domain.Submission,
	spec domain.ExecutionSpec,
	tc domain.TestCase,
) (outport.RunRequest, error) {
	if !languageSupported(spec, sub.Language) {
		return outport.RunRequest{}, fmt.Errorf(
			"%s is not supported by this problem execution spec",
			sub.Language,
		)
	}

	input, err := decodeHarnessInput(tc)
	if err != nil {
		return outport.RunRequest{}, err
	}

	var code string
	switch sub.Language {
	case domain.LanguagePython:
		code, err = renderPythonHarness(sub.Code, spec, input)
	case domain.LanguageGo:
		code, err = renderGoHarness(sub.Code, spec, input)
	case domain.LanguageJavascript:
		code, err = renderJavaScriptHarness(sub.Code, spec, input, false)
	case domain.LanguageTypescript:
		code, err = renderJavaScriptHarness(sub.Code, spec, input, true)
	case domain.LanguageRust:
		code, err = renderRustHarness(sub.Code, spec, input)
	default:
		err = fmt.Errorf("unsupported submission language: %s", sub.Language)
	}
	if err != nil {
		return outport.RunRequest{}, err
	}
	return outport.RunRequest{
		Language: string(sub.Language),
		Code:     code,
		MemoryMB: spec.MemoryMB,
	}, nil
}

func languageSupported(spec domain.ExecutionSpec, language domain.Language) bool {
	if len(spec.SupportedLanguages) == 0 {
		return true
	}
	for _, supported := range spec.SupportedLanguages {
		if supported == language {
			return true
		}
	}
	return false
}

func decodeHarnessInput(tc domain.TestCase) (harnessInput, error) {
	raw := tc.InputJSON
	if len(raw) == 0 {
		raw = json.RawMessage(tc.Input)
	}
	if !json.Valid(raw) {
		return harnessInput{}, errors.New("structured execution requires valid json input_json")
	}
	var input harnessInput
	if err := json.Unmarshal(raw, &input); err == nil &&
		(input.Args != nil || input.Constructor != nil || input.Calls != nil) {
		return input, nil
	}
	var args []json.RawMessage
	if err := json.Unmarshal(raw, &args); err != nil {
		return harnessInput{}, errors.New(`structured input must contain "args" or class call data`)
	}
	input.Args = args
	return input, nil
}

func executionBinding(spec domain.ExecutionSpec, language domain.Language) domain.ExecutionLanguageBind {
	binding := spec.Bindings[string(language)]
	if binding.Entrypoint == "" {
		binding.Entrypoint = spec.Entrypoint
		if language == domain.LanguageRust {
			binding.Entrypoint = snakeCase(binding.Entrypoint)
		}
	}
	if binding.ClassName == "" {
		binding.ClassName = spec.ClassName
	}
	if binding.ClassName == "" && spec.Mode != domain.ExecutionModeClass {
		binding.ClassName = "Solution"
	}
	if binding.Constructor == "" {
		switch language {
		case domain.LanguageGo:
			binding.Constructor = "Constructor"
		case domain.LanguageRust:
			binding.Constructor = "new"
		}
	}
	return binding
}

func validateFunctionSpec(spec domain.ExecutionSpec, args []json.RawMessage) error {
	if spec.Mode != domain.ExecutionModeFunction && spec.Mode != domain.ExecutionModeInPlace {
		return fmt.Errorf("unsupported function execution mode: %s", spec.Mode)
	}
	if !identifierPattern.MatchString(spec.Entrypoint) {
		return fmt.Errorf("invalid entrypoint: %q", spec.Entrypoint)
	}
	if len(spec.Signature.Params) > 0 && len(args) != len(spec.Signature.Params) {
		return fmt.Errorf("expected %d arguments, got %d", len(spec.Signature.Params), len(args))
	}
	if spec.Mode == domain.ExecutionModeInPlace {
		index := spec.Output.ParamIndex
		if index < 0 || index >= len(args) {
			return fmt.Errorf("invalid in-place output param index: %d", index)
		}
	}
	return nil
}

func resultType(spec domain.ExecutionSpec) string {
	if spec.Mode == domain.ExecutionModeInPlace || spec.Output.Source == "param" {
		index := spec.Output.ParamIndex
		if index >= 0 && index < len(spec.Signature.Params) {
			return spec.Signature.Params[index].Type
		}
	}
	return spec.Signature.Returns
}

func renderPythonHarness(userCode string, spec domain.ExecutionSpec, input harnessInput) (string, error) {
	binding := executionBinding(spec, domain.LanguagePython)
	if spec.Mode == domain.ExecutionModeClass {
		return renderPythonClassHarness(userCode, spec, binding, input)
	}
	if err := validateFunctionSpec(spec, input.Args); err != nil {
		return "", err
	}
	if !identifierPattern.MatchString(binding.Entrypoint) ||
		!identifierPattern.MatchString(binding.ClassName) {
		return "", errors.New("invalid Python execution binding")
	}
	args, _ := json.Marshal(input.Args)
	paramTypes, _ := json.Marshal(paramTypeNames(spec.Signature.Params))
	result := fmt.Sprintf("%s().%s(*__jl_args)", binding.ClassName, binding.Entrypoint)
	if spec.Mode == domain.ExecutionModeInPlace || spec.Output.Source == "param" {
		result = fmt.Sprintf("(__jl_obj.%s(*__jl_args), __jl_args[%d])[1]",
			binding.Entrypoint, spec.Output.ParamIndex)
		return fmt.Sprintf(`%s

%s

import json as __jl_json
__jl_raw_args = __jl_json.loads(%s)
__jl_types = __jl_json.loads(%s)
__jl_args = [__jl_decode(t, value) for t, value in zip(__jl_types, __jl_raw_args)]
__jl_obj = %s()
__jl_result = %s
print(__jl_json.dumps(__jl_encode(%s, __jl_result), separators=(",", ":")))
`, pythonCodec, userCode, strconv.Quote(string(args)), strconv.Quote(string(paramTypes)),
			binding.ClassName, result, strconv.Quote(resultType(spec))), nil
	}
	return fmt.Sprintf(`%s

%s

import json as __jl_json
__jl_raw_args = __jl_json.loads(%s)
__jl_types = __jl_json.loads(%s)
__jl_args = [__jl_decode(t, value) for t, value in zip(__jl_types, __jl_raw_args)]
__jl_result = %s
print(__jl_json.dumps(__jl_encode(%s, __jl_result), separators=(",", ":")))
`, pythonCodec, userCode, strconv.Quote(string(args)), strconv.Quote(string(paramTypes)),
		result, strconv.Quote(resultType(spec))), nil
}

func renderPythonFunctionHarness(
	userCode string,
	spec domain.ExecutionSpec,
	tc domain.TestCase,
) (string, error) {
	input, err := decodeHarnessInput(tc)
	if err != nil {
		return "", err
	}
	return renderPythonHarness(userCode, spec, input)
}

func renderPythonClassHarness(
	userCode string,
	spec domain.ExecutionSpec,
	binding domain.ExecutionLanguageBind,
	input harnessInput,
) (string, error) {
	if !identifierPattern.MatchString(binding.ClassName) {
		return "", errors.New("invalid Python class binding")
	}
	payload, _ := json.Marshal(input)
	methodParams := make(map[string][]string, len(spec.Methods))
	methodReturns := make(map[string]string, len(spec.Methods))
	methodNames := make(map[string]string, len(spec.Methods))
	for name, method := range spec.Methods {
		methodParams[name] = paramTypeNames(method.Params)
		methodReturns[name] = method.Returns
		methodNames[name] = methodBinding(name, method, domain.LanguagePython)
	}
	constructorTypes, _ := json.Marshal(paramTypeNames(spec.Constructor.Params))
	methodParamJSON, _ := json.Marshal(methodParams)
	methodReturnJSON, _ := json.Marshal(methodReturns)
	methodNameJSON, _ := json.Marshal(methodNames)
	return fmt.Sprintf(`%s

%s

import json as __jl_json
__jl_case = __jl_json.loads(%s)
__jl_constructor_types = __jl_json.loads(%s)
__jl_method_types = __jl_json.loads(%s)
__jl_method_returns = __jl_json.loads(%s)
__jl_method_names = __jl_json.loads(%s)
__jl_constructor = [
    __jl_decode(t, value)
    for t, value in zip(__jl_constructor_types, __jl_case.get("constructor", []))
]
__jl_obj = %s(*__jl_constructor)
__jl_result = []
for __jl_call in __jl_case.get("calls", []):
    __jl_name = __jl_call["method"]
    __jl_args = [
        __jl_decode(t, value)
        for t, value in zip(__jl_method_types[__jl_name], __jl_call.get("args", []))
    ]
    __jl_value = getattr(__jl_obj, __jl_method_names[__jl_name])(*__jl_args)
    __jl_result.append(__jl_encode(__jl_method_returns[__jl_name], __jl_value))
print(__jl_json.dumps(__jl_result, separators=(",", ":")))
`, pythonCodec, userCode, strconv.Quote(string(payload)),
		strconv.Quote(string(constructorTypes)), strconv.Quote(string(methodParamJSON)),
		strconv.Quote(string(methodReturnJSON)), strconv.Quote(string(methodNameJSON)),
		binding.ClassName), nil
}

const pythonCodec = `from typing import *
from collections import deque

class ListNode:
    def __init__(self, val=0, next=None):
        self.val = val
        self.next = next

class TreeNode:
    def __init__(self, val=0, left=None, right=None):
        self.val = val
        self.left = left
        self.right = right

class Node:
    def __init__(self, val=0, neighbors=None):
        self.val = val
        self.neighbors = neighbors if neighbors is not None else []

def __jl_decode(kind, value):
    normalized = kind.replace(" ", "").lower()
    if normalized in ("listnode", "listnode?", "optional[listnode]"):
        dummy = ListNode()
        tail = dummy
        for item in value or []:
            tail.next = ListNode(item)
            tail = tail.next
        return dummy.next
    if normalized in ("treenode", "treenode?", "optional[treenode]"):
        if not value:
            return None
        nodes = [None if item is None else TreeNode(item) for item in value]
        children = iter(nodes[1:])
        for node in nodes:
            if node is None:
                continue
            node.left = next(children, None)
            node.right = next(children, None)
        return nodes[0]
    if normalized in ("graphnode", "graphnode?", "optional[graphnode]"):
        if not value:
            return None
        nodes = [Node(index + 1) for index in range(len(value))]
        for index, neighbors in enumerate(value):
            nodes[index].neighbors = [nodes[item - 1] for item in neighbors]
        return nodes[0]
    return value

def __jl_encode(kind, value):
    normalized = kind.replace(" ", "").lower()
    if normalized in ("void", "none", "nonetype"):
        return None
    if normalized in ("listnode", "listnode?", "optional[listnode]"):
        result = []
        while value is not None:
            result.append(value.val)
            value = value.next
        return result
    if normalized in ("treenode", "treenode?", "optional[treenode]"):
        if value is None:
            return []
        result = []
        queue = deque([value])
        while queue:
            node = queue.popleft()
            if node is None:
                result.append(None)
                continue
            result.append(node.val)
            queue.append(node.left)
            queue.append(node.right)
        while result and result[-1] is None:
            result.pop()
        return result
    if normalized in ("graphnode", "graphnode?", "optional[graphnode]"):
        if value is None:
            return []
        nodes = {}
        queue = deque([value])
        while queue:
            node = queue.popleft()
            if node.val in nodes:
                continue
            nodes[node.val] = node
            queue.extend(node.neighbors)
        return [
            sorted(neighbor.val for neighbor in nodes[index].neighbors)
            for index in sorted(nodes)
        ]
    return value`

func renderJavaScriptHarness(
	userCode string,
	spec domain.ExecutionSpec,
	input harnessInput,
	typescript bool,
) (string, error) {
	language := domain.LanguageJavascript
	if typescript {
		language = domain.LanguageTypescript
	}
	binding := executionBinding(spec, language)
	payload, _ := json.Marshal(input)
	prefix := ""
	if typescript {
		prefix = "// deno-lint-ignore-file no-explicit-any\n"
	}
	if spec.Mode == domain.ExecutionModeClass {
		if !identifierPattern.MatchString(binding.ClassName) {
			return "", errors.New("invalid JavaScript class binding")
		}
		methodParams := make(map[string][]string, len(spec.Methods))
		methodReturns := make(map[string]string, len(spec.Methods))
		methodNames := make(map[string]string, len(spec.Methods))
		for name, method := range spec.Methods {
			methodParams[name] = paramTypeNames(method.Params)
			methodReturns[name] = method.Returns
			methodNames[name] = methodBinding(name, method, language)
		}
		constructorTypes, _ := json.Marshal(paramTypeNames(spec.Constructor.Params))
		methodParamJSON, _ := json.Marshal(methodParams)
		methodReturnJSON, _ := json.Marshal(methodReturns)
		methodNameJSON, _ := json.Marshal(methodNames)
		return fmt.Sprintf(`%s%s

%s

const __jlCase = JSON.parse(%s)
const __jlConstructorTypes = JSON.parse(%s)
const __jlMethodTypes = JSON.parse(%s)
const __jlMethodReturns = JSON.parse(%s)
const __jlMethodNames = JSON.parse(%s)
const __jlConstructor = (__jlCase.constructor ?? []).map(
  (value, index) => __jlDecode(__jlConstructorTypes[index], value)
)
const __jlObject = new %s(...__jlConstructor)
const __jlResult = []
for (const __jlCall of (__jlCase.calls ?? [])) {
  const __jlArgs = (__jlCall.args ?? []).map(
    (value, index) => __jlDecode(__jlMethodTypes[__jlCall.method][index], value)
  )
  const __jlValue = __jlObject[__jlMethodNames[__jlCall.method]](...__jlArgs)
  __jlResult.push(__jlEncode(__jlMethodReturns[__jlCall.method], __jlValue))
}
console.log(JSON.stringify(__jlResult))
`, prefix, javaScriptCodec, userCode, strconv.Quote(string(payload)),
			strconv.Quote(string(constructorTypes)), strconv.Quote(string(methodParamJSON)),
			strconv.Quote(string(methodReturnJSON)), strconv.Quote(string(methodNameJSON)),
			binding.ClassName), nil
	}
	if err := validateFunctionSpec(spec, input.Args); err != nil {
		return "", err
	}
	if !identifierPattern.MatchString(binding.Entrypoint) {
		return "", errors.New("invalid JavaScript entrypoint")
	}
	args, _ := json.Marshal(input.Args)
	paramTypes, _ := json.Marshal(paramTypeNames(spec.Signature.Params))
	result := fmt.Sprintf("%s(...__jlArgs)", binding.Entrypoint)
	if spec.Mode == domain.ExecutionModeInPlace || spec.Output.Source == "param" {
		result = fmt.Sprintf("(%s(...__jlArgs), __jlArgs[%d])",
			binding.Entrypoint, spec.Output.ParamIndex)
	}
	return fmt.Sprintf(`%s%s

%s

const __jlRawArgs = JSON.parse(%s)
const __jlTypes = JSON.parse(%s)
const __jlArgs = __jlRawArgs.map((value, index) => __jlDecode(__jlTypes[index], value))
const __jlResult = %s
console.log(JSON.stringify(__jlEncode(%s, __jlResult)))
`, prefix, javaScriptCodec, userCode, strconv.Quote(string(args)),
		strconv.Quote(string(paramTypes)), result, strconv.Quote(resultType(spec))), nil
}

const javaScriptCodec = `function __jlDecode(kind, value) {
  const normalized = String(kind ?? '').replaceAll(' ', '').toLowerCase()
  if (['listnode', 'listnode?', 'optional[listnode]'].includes(normalized)) {
    const dummy = { val: 0, next: null }
    let tail = dummy
    for (const item of (value ?? [])) {
      tail.next = { val: item, next: null }
      tail = tail.next
    }
    return dummy.next
  }
  if (['treenode', 'treenode?', 'optional[treenode]'].includes(normalized)) {
    if (!value?.length) return null
    const nodes = value.map((item) =>
      item === null ? null : { val: item, left: null, right: null }
    )
    let child = 1
    for (const node of nodes) {
      if (node === null) continue
      node.left = nodes[child++] ?? null
      node.right = nodes[child++] ?? null
    }
    return nodes[0]
  }
  if (['graphnode', 'graphnode?', 'optional[graphnode]'].includes(normalized)) {
    if (!value?.length) return null
    const nodes = value.map((_, index) => ({ val: index + 1, neighbors: [] }))
    value.forEach((neighbors, index) => {
      nodes[index].neighbors = neighbors.map((item) => nodes[item - 1])
    })
    return nodes[0]
  }
  return value
}

function __jlEncode(kind, value) {
  const normalized = String(kind ?? '').replaceAll(' ', '').toLowerCase()
  if (['void', 'none', 'nonetype'].includes(normalized)) return null
  if (['listnode', 'listnode?', 'optional[listnode]'].includes(normalized)) {
    const result = []
    while (value !== null && value !== undefined) {
      result.push(value.val)
      value = value.next
    }
    return result
  }
  if (['treenode', 'treenode?', 'optional[treenode]'].includes(normalized)) {
    if (value === null || value === undefined) return []
    const result = []
    const queue = [value]
    for (let index = 0; index < queue.length; index += 1) {
      const node = queue[index]
      if (node === null || node === undefined) {
        result.push(null)
        continue
      }
      result.push(node.val)
      queue.push(node.left, node.right)
    }
    while (result.at(-1) === null) result.pop()
    return result
  }
  if (['graphnode', 'graphnode?', 'optional[graphnode]'].includes(normalized)) {
    if (value === null || value === undefined) return []
    const nodes = new Map()
    const queue = [value]
    for (let index = 0; index < queue.length; index += 1) {
      const node = queue[index]
      if (nodes.has(node.val)) continue
      nodes.set(node.val, node)
      queue.push(...node.neighbors)
    }
    return [...nodes.keys()].sort((a, b) => a - b).map((key) =>
      nodes.get(key).neighbors.map((node) => node.val).sort((a, b) => a - b)
    )
  }
  return value
}`

func renderGoHarness(userCode string, spec domain.ExecutionSpec, input harnessInput) (string, error) {
	if spec.Mode == domain.ExecutionModeClass {
		return renderGoClassHarness(userCode, spec, input)
	}
	if err := validateFunctionSpec(spec, input.Args); err != nil {
		return "", err
	}
	binding := executionBinding(spec, domain.LanguageGo)
	if !identifierPattern.MatchString(binding.Entrypoint) {
		return "", errors.New("invalid Go entrypoint")
	}
	argNames, declarations, err := goArguments("__jlArg", spec.Signature.Params, input.Args)
	if err != nil {
		return "", err
	}
	call := fmt.Sprintf("%s(%s)", binding.Entrypoint, strings.Join(argNames, ", "))
	body := fmt.Sprintf("\t__jlResult := %s\n", goEncodeExpression(call, resultType(spec)))
	if normalizeType(spec.Signature.Returns) == "void" {
		body = fmt.Sprintf("\t%s\n\tvar __jlResult any = nil\n", call)
	}
	if spec.Mode == domain.ExecutionModeInPlace || spec.Output.Source == "param" {
		body = fmt.Sprintf("\t%s\n\t__jlResult := %s\n", call,
			goEncodeExpression(argNames[spec.Output.ParamIndex], resultType(spec)))
	}
	return goProgram(userCode, declarations+body), nil
}

func renderGoClassHarness(userCode string, spec domain.ExecutionSpec, input harnessInput) (string, error) {
	binding := executionBinding(spec, domain.LanguageGo)
	if !identifierPattern.MatchString(binding.Constructor) {
		return "", errors.New("invalid Go constructor binding")
	}
	constructorArgs, declarations, err := goArguments("__jlCtor", spec.Constructor.Params, input.Constructor)
	if err != nil {
		return "", err
	}
	var body strings.Builder
	body.WriteString(declarations)
	fmt.Fprintf(&body, "\t__jlObject := %s(%s)\n", binding.Constructor, strings.Join(constructorArgs, ", "))
	body.WriteString("\t__jlResults := make([]any, 0)\n")
	for i, call := range input.Calls {
		method, ok := spec.Methods[call.Method]
		if !ok || !identifierPattern.MatchString(call.Method) {
			return "", fmt.Errorf("unknown class method: %s", call.Method)
		}
		args, decl, err := goArguments(fmt.Sprintf("__jlCall%dArg", i), method.Params, call.Args)
		if err != nil {
			return "", err
		}
		body.WriteString(decl)
		invocation := fmt.Sprintf("__jlObject.%s(%s)",
			methodBinding(call.Method, method, domain.LanguageGo), strings.Join(args, ", "))
		if normalizeType(method.Returns) == "void" {
			fmt.Fprintf(&body, "\t%s\n\t__jlResults = append(__jlResults, nil)\n", invocation)
		} else {
			fmt.Fprintf(&body, "\t__jlResults = append(__jlResults, %s)\n",
				goEncodeExpression(invocation, method.Returns))
		}
	}
	body.WriteString("\t__jlResult := __jlResults\n")
	return goProgram(userCode, body.String()), nil
}

func goArguments(prefix string, params []domain.ExecutionParam, raw []json.RawMessage) ([]string, string, error) {
	if len(params) != len(raw) {
		return nil, "", fmt.Errorf("expected %d arguments, got %d", len(params), len(raw))
	}
	args := make([]string, len(raw))
	var declarations strings.Builder
	for i := range raw {
		t, err := goType(params[i].Type)
		if err != nil {
			return nil, "", err
		}
		args[i] = fmt.Sprintf("%s%d", prefix, i)
		normalized := structuredType(params[i].Type)
		switch normalized {
		case "listnode":
			rawName := args[i] + "Raw"
			fmt.Fprintf(&declarations,
				"\tvar %s []int\n\tif err := __jlJSON.Unmarshal([]byte(%s), &%s); err != nil { panic(err) }\n\t%s := __jlListFromSlice(%s)\n",
				rawName, strconv.Quote(string(raw[i])), rawName, args[i], rawName)
		case "treenode":
			rawName := args[i] + "Raw"
			fmt.Fprintf(&declarations,
				"\tvar %s []*int\n\tif err := __jlJSON.Unmarshal([]byte(%s), &%s); err != nil { panic(err) }\n\t%s := __jlTreeFromSlice(%s)\n",
				rawName, strconv.Quote(string(raw[i])), rawName, args[i], rawName)
		case "graphnode":
			rawName := args[i] + "Raw"
			fmt.Fprintf(&declarations,
				"\tvar %s [][]int\n\tif err := __jlJSON.Unmarshal([]byte(%s), &%s); err != nil { panic(err) }\n\t%s := __jlGraphFromAdjacency(%s)\n",
				rawName, strconv.Quote(string(raw[i])), rawName, args[i], rawName)
		default:
			fmt.Fprintf(&declarations,
				"\tvar %s %s\n\tif err := __jlJSON.Unmarshal([]byte(%s), &%s); err != nil { panic(err) }\n",
				args[i], t, strconv.Quote(string(raw[i])), args[i])
		}
	}
	return args, declarations.String(), nil
}

func goProgram(userCode, body string) string {
	trimmed := strings.TrimSpace(userCode)
	if strings.HasPrefix(trimmed, "package ") {
		lines := strings.SplitN(trimmed, "\n", 2)
		rest := ""
		if len(lines) == 2 {
			rest = lines[1]
		}
		userCode = lines[0] + "\nimport __jlJSON \"encoding/json\"\n" + rest
	} else {
		userCode = "package main\n\nimport __jlJSON \"encoding/json\"\n\n" + userCode
	}
	return fmt.Sprintf(`%s

type ListNode struct {
	Val int
	Next *ListNode
}

type TreeNode struct {
	Val int
	Left *TreeNode
	Right *TreeNode
}

type Node struct {
	Val int
	Neighbors []*Node
}

func __jlListFromSlice(values []int) *ListNode {
	dummy := &ListNode{}
	tail := dummy
	for _, value := range values {
		tail.Next = &ListNode{Val: value}
		tail = tail.Next
	}
	return dummy.Next
}

func __jlListToSlice(node *ListNode) []int {
	values := make([]int, 0)
	for node != nil {
		values = append(values, node.Val)
		node = node.Next
	}
	return values
}

func __jlTreeFromSlice(values []*int) *TreeNode {
	if len(values) == 0 || values[0] == nil { return nil }
	root := &TreeNode{Val: *values[0]}
	queue := []*TreeNode{root}
	index := 1
	for len(queue) > 0 && index < len(values) {
		node := queue[0]
		queue = queue[1:]
		if index < len(values) && values[index] != nil {
			node.Left = &TreeNode{Val: *values[index]}
			queue = append(queue, node.Left)
		}
		index++
		if index < len(values) && values[index] != nil {
			node.Right = &TreeNode{Val: *values[index]}
			queue = append(queue, node.Right)
		}
		index++
	}
	return root
}

func __jlTreeToSlice(root *TreeNode) []*int {
	if root == nil { return []*int{} }
	values := make([]*int, 0)
	queue := []*TreeNode{root}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		if node == nil {
			values = append(values, nil)
			continue
		}
		value := node.Val
		values = append(values, &value)
		queue = append(queue, node.Left, node.Right)
	}
	for len(values) > 0 && values[len(values)-1] == nil {
		values = values[:len(values)-1]
	}
	return values
}

func __jlGraphFromAdjacency(adjacency [][]int) *Node {
	if len(adjacency) == 0 { return nil }
	nodes := make([]*Node, len(adjacency))
	for index := range nodes {
		nodes[index] = &Node{Val: index + 1}
	}
	for index, neighbors := range adjacency {
		for _, neighbor := range neighbors {
			nodes[index].Neighbors = append(nodes[index].Neighbors, nodes[neighbor-1])
		}
	}
	return nodes[0]
}

func __jlGraphToAdjacency(root *Node) [][]int {
	if root == nil { return [][]int{} }
	nodes := map[int]*Node{}
	queue := []*Node{root}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		if _, exists := nodes[node.Val]; exists { continue }
		nodes[node.Val] = node
		queue = append(queue, node.Neighbors...)
	}
	result := make([][]int, len(nodes))
	for value, node := range nodes {
		for _, neighbor := range node.Neighbors {
			result[value-1] = append(result[value-1], neighbor.Val)
		}
	}
	return result
}

func main() {
%s	__jlEncoded, __jlErr := __jlJSON.Marshal(__jlResult)
	if __jlErr != nil { panic(__jlErr) }
	println(string(__jlEncoded))
}
`, userCode, body)
}

func goType(raw string) (string, error) {
	t := normalizeType(raw)
	if strings.HasSuffix(t, "[]") {
		child, err := goType(strings.TrimSuffix(t, "[]"))
		if err != nil {
			return "", err
		}
		return "[]" + child, nil
	}
	switch strings.ToLower(t) {
	case "int":
		return "int", nil
	case "int64", "long":
		return "int64", nil
	case "float", "double":
		return "float64", nil
	case "bool":
		return "bool", nil
	case "string":
		return "string", nil
	case "void":
		return "", nil
	case "listnode", "listnode?", "optional[listnode]":
		return "*ListNode", nil
	case "treenode", "treenode?", "optional[treenode]":
		return "*TreeNode", nil
	case "graphnode", "graphnode?", "optional[graphnode]":
		return "*Node", nil
	default:
		return "", fmt.Errorf("Go harness does not support type %q yet", raw)
	}
}

func goEncodeExpression(expression, rawType string) string {
	switch structuredType(rawType) {
	case "listnode":
		return "__jlListToSlice(" + expression + ")"
	case "treenode":
		return "__jlTreeToSlice(" + expression + ")"
	case "graphnode":
		return "__jlGraphToAdjacency(" + expression + ")"
	default:
		return expression
	}
}

func renderRustHarness(userCode string, spec domain.ExecutionSpec, input harnessInput) (string, error) {
	if spec.Mode == domain.ExecutionModeClass {
		return renderRustClassHarness(userCode, spec, input)
	}
	if err := validateFunctionSpec(spec, input.Args); err != nil {
		return "", err
	}
	binding := executionBinding(spec, domain.LanguageRust)
	if !identifierPattern.MatchString(binding.Entrypoint) {
		return "", errors.New("invalid Rust entrypoint")
	}
	args := make([]string, len(input.Args))
	for i, raw := range input.Args {
		literal, err := rustLiteral(raw, spec.Signature.Params[i].Type)
		if err != nil {
			return "", err
		}
		args[i] = literal
	}
	call := fmt.Sprintf("Solution::%s(%s)", binding.Entrypoint, strings.Join(args, ", "))
	mainBody := fmt.Sprintf("let __jl_result = %s;", call)
	outputType := spec.Signature.Returns
	if spec.Mode == domain.ExecutionModeInPlace || spec.Output.Source == "param" {
		index := spec.Output.ParamIndex
		var declarations strings.Builder
		callArgs := make([]string, len(args))
		for i, arg := range args {
			name := fmt.Sprintf("__jl_arg%d", i)
			if i == index {
				fmt.Fprintf(&declarations, "let mut %s = %s;\n", name, arg)
				callArgs[i] = "&mut " + name
			} else {
				fmt.Fprintf(&declarations, "let %s = %s;\n", name, arg)
				callArgs[i] = name
			}
		}
		mainBody = declarations.String() +
			fmt.Sprintf("Solution::%s(%s);\nlet __jl_result = __jl_arg%d;",
				binding.Entrypoint, strings.Join(callArgs, ", "), index)
		outputType = spec.Signature.Params[index].Type
	}
	if _, err := rustType(outputType); err != nil {
		return "", err
	}
	return rustProgram(userCode, mainBody), nil
}

func renderRustClassHarness(userCode string, spec domain.ExecutionSpec, input harnessInput) (string, error) {
	binding := executionBinding(spec, domain.LanguageRust)
	if !identifierPattern.MatchString(binding.ClassName) ||
		!identifierPattern.MatchString(binding.Constructor) {
		return "", errors.New("invalid Rust class binding")
	}
	constructorArgs, err := rustArguments(spec.Constructor.Params, input.Constructor)
	if err != nil {
		return "", err
	}
	var body strings.Builder
	fmt.Fprintf(&body, "let mut __jl_object = %s::%s(%s);\n",
		binding.ClassName, binding.Constructor, strings.Join(constructorArgs, ", "))
	body.WriteString("let mut __jl_results: Vec<String> = Vec::new();\n")
	for _, call := range input.Calls {
		method, ok := spec.Methods[call.Method]
		if !ok || !identifierPattern.MatchString(call.Method) {
			return "", fmt.Errorf("unknown class method: %s", call.Method)
		}
		args, err := rustArguments(method.Params, call.Args)
		if err != nil {
			return "", err
		}
		methodName := methodBinding(call.Method, method, domain.LanguageRust)
		invocation := fmt.Sprintf("__jl_object.%s(%s)", methodName, strings.Join(args, ", "))
		if normalizeType(method.Returns) == "void" {
			fmt.Fprintf(&body, "%s;\n__jl_results.push(String::from(\"null\"));\n", invocation)
		} else {
			if _, err := rustType(method.Returns); err != nil {
				return "", err
			}
			fmt.Fprintf(&body, "__jl_results.push(__jl_json(&%s));\n", invocation)
		}
	}
	body.WriteString("let __jl_result = JLRawArray(__jl_results);")
	return rustProgram(userCode, body.String()), nil
}

func rustArguments(params []domain.ExecutionParam, raw []json.RawMessage) ([]string, error) {
	if len(params) != len(raw) {
		return nil, fmt.Errorf("expected %d arguments, got %d", len(params), len(raw))
	}
	args := make([]string, len(raw))
	for i := range raw {
		literal, err := rustLiteral(raw[i], params[i].Type)
		if err != nil {
			return nil, err
		}
		args[i] = literal
	}
	return args, nil
}

func rustProgram(userCode, body string) string {
	return fmt.Sprintf(`struct Solution;

#[derive(PartialEq, Eq, Clone, Debug)]
pub struct ListNode {
    pub val: i32,
    pub next: Option<Box<ListNode>>,
}

#[derive(PartialEq, Eq, Clone, Debug)]
pub struct TreeNode {
    pub val: i32,
    pub left: Option<std::rc::Rc<std::cell::RefCell<TreeNode>>>,
    pub right: Option<std::rc::Rc<std::cell::RefCell<TreeNode>>>,
}

#[derive(Clone, Debug)]
pub struct Node {
    pub val: i32,
    pub neighbors: Vec<std::rc::Rc<std::cell::RefCell<Node>>>,
}

fn __jl_list(values: Vec<i32>) -> Option<Box<ListNode>> {
    let mut head = None;
    for value in values.into_iter().rev() {
        head = Some(Box::new(ListNode { val: value, next: head }));
    }
    head
}

fn __jl_tree(values: Vec<Option<i32>>) -> Option<std::rc::Rc<std::cell::RefCell<TreeNode>>> {
    let Some(Some(root_value)) = values.first().cloned() else { return None };
    let root = std::rc::Rc::new(std::cell::RefCell::new(TreeNode {
        val: root_value, left: None, right: None,
    }));
    let mut queue = std::collections::VecDeque::from([std::rc::Rc::clone(&root)]);
    let mut index = 1;
    while let Some(node) = queue.pop_front() {
        if index < values.len() {
            if let Some(value) = values[index] {
                let child = std::rc::Rc::new(std::cell::RefCell::new(TreeNode {
                    val: value, left: None, right: None,
                }));
                node.borrow_mut().left = Some(std::rc::Rc::clone(&child));
                queue.push_back(child);
            }
            index += 1;
        }
        if index < values.len() {
            if let Some(value) = values[index] {
                let child = std::rc::Rc::new(std::cell::RefCell::new(TreeNode {
                    val: value, left: None, right: None,
                }));
                node.borrow_mut().right = Some(std::rc::Rc::clone(&child));
                queue.push_back(child);
            }
            index += 1;
        }
    }
    Some(root)
}

fn __jl_graph(adjacency: Vec<Vec<i32>>) -> Option<std::rc::Rc<std::cell::RefCell<Node>>> {
    if adjacency.is_empty() { return None; }
    let nodes: Vec<_> = (0..adjacency.len())
        .map(|index| std::rc::Rc::new(std::cell::RefCell::new(Node {
            val: index as i32 + 1,
            neighbors: Vec::new(),
        })))
        .collect();
    for (index, neighbors) in adjacency.iter().enumerate() {
        nodes[index].borrow_mut().neighbors = neighbors.iter()
            .map(|value| std::rc::Rc::clone(&nodes[*value as usize - 1]))
            .collect();
    }
    Some(std::rc::Rc::clone(&nodes[0]))
}

trait JLJson { fn jl_json(&self) -> String; }
fn __jl_json<T: JLJson>(value: &T) -> String { value.jl_json() }
impl JLJson for i32 { fn jl_json(&self) -> String { self.to_string() } }
impl JLJson for i64 { fn jl_json(&self) -> String { self.to_string() } }
impl JLJson for f64 { fn jl_json(&self) -> String { self.to_string() } }
impl JLJson for bool { fn jl_json(&self) -> String { self.to_string() } }
impl JLJson for () { fn jl_json(&self) -> String { String::from("null") } }
impl JLJson for String {
    fn jl_json(&self) -> String {
        let mut out = String::from("\"");
        for ch in self.chars() {
            match ch {
                '"' => out.push_str("\\\""),
                '\\' => out.push_str("\\\\"),
                '\n' => out.push_str("\\n"),
                '\r' => out.push_str("\\r"),
                '\t' => out.push_str("\\t"),
                _ => out.push(ch),
            }
        }
        out.push('"');
        out
    }
}
impl<T: JLJson> JLJson for Vec<T> {
    fn jl_json(&self) -> String {
        let mut out = String::from("[");
        for (index, value) in self.iter().enumerate() {
            if index > 0 { out.push(','); }
            out.push_str(&value.jl_json());
        }
        out.push(']');
        out
    }
}
impl<T: JLJson> JLJson for Option<T> {
    fn jl_json(&self) -> String {
        match self { Some(value) => value.jl_json(), None => String::from("null") }
    }
}
impl JLJson for Box<ListNode> {
    fn jl_json(&self) -> String {
        let mut values = Vec::new();
        let mut current = Some(self.as_ref());
        while let Some(node) = current {
            values.push(node.val);
            current = node.next.as_deref();
        }
        values.jl_json()
    }
}
impl JLJson for std::rc::Rc<std::cell::RefCell<TreeNode>> {
    fn jl_json(&self) -> String {
        let mut values: Vec<Option<i32>> = Vec::new();
        let mut queue = std::collections::VecDeque::from([Some(std::rc::Rc::clone(self))]);
        while let Some(node) = queue.pop_front() {
            match node {
                None => values.push(None),
                Some(node) => {
                    let borrowed = node.borrow();
                    values.push(Some(borrowed.val));
                    queue.push_back(borrowed.left.clone());
                    queue.push_back(borrowed.right.clone());
                }
            }
        }
        while values.last() == Some(&None) { values.pop(); }
        values.jl_json()
    }
}
impl JLJson for std::rc::Rc<std::cell::RefCell<Node>> {
    fn jl_json(&self) -> String {
        let mut nodes = std::collections::BTreeMap::new();
        let mut queue = std::collections::VecDeque::from([std::rc::Rc::clone(self)]);
        while let Some(node) = queue.pop_front() {
            let borrowed = node.borrow();
            if nodes.contains_key(&borrowed.val) { continue; }
            let neighbors: Vec<i32> = borrowed.neighbors.iter()
                .map(|neighbor| neighbor.borrow().val)
                .collect();
            queue.extend(borrowed.neighbors.iter().cloned());
            nodes.insert(borrowed.val, neighbors);
        }
        nodes.values().cloned().collect::<Vec<Vec<i32>>>().jl_json()
    }
}
struct JLRawArray(Vec<String>);
impl JLJson for JLRawArray {
    fn jl_json(&self) -> String { format!("[{}]", self.0.join(",")) }
}

%s

fn main() {
%s
    println!("{}", __jl_json(&__jl_result));
}
`, userCode, indent(body, "    "))
}

func rustLiteral(raw json.RawMessage, rawType string) (string, error) {
	t := normalizeType(rawType)
	if strings.HasSuffix(t, "[]") {
		childType := strings.TrimSuffix(t, "[]")
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			return "", fmt.Errorf("expected array for %s", rawType)
		}
		values := make([]string, len(items))
		for i := range items {
			value, err := rustLiteral(items[i], childType)
			if err != nil {
				return "", err
			}
			values[i] = value
		}
		return "vec![" + strings.Join(values, ", ") + "]", nil
	}
	switch strings.ToLower(t) {
	case "int":
		var value int32
		if err := json.Unmarshal(raw, &value); err != nil {
			return "", err
		}
		return strconv.FormatInt(int64(value), 10), nil
	case "int64", "long":
		var value int64
		if err := json.Unmarshal(raw, &value); err != nil {
			return "", err
		}
		return strconv.FormatInt(value, 10) + "i64", nil
	case "float", "double":
		var value float64
		if err := json.Unmarshal(raw, &value); err != nil {
			return "", err
		}
		return strconv.FormatFloat(value, 'g', -1, 64) + "f64", nil
	case "bool":
		var value bool
		if err := json.Unmarshal(raw, &value); err != nil {
			return "", err
		}
		return strconv.FormatBool(value), nil
	case "string":
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return "", err
		}
		return "String::from(" + rustString(value) + ")", nil
	case "listnode", "listnode?", "optional[listnode]":
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			return "", fmt.Errorf("expected array for ListNode")
		}
		values := make([]string, len(items))
		for i := range items {
			value, err := rustLiteral(items[i], "int")
			if err != nil {
				return "", err
			}
			values[i] = value
		}
		return "__jl_list(vec![" + strings.Join(values, ", ") + "])", nil
	case "treenode", "treenode?", "optional[treenode]":
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			return "", fmt.Errorf("expected array for TreeNode")
		}
		values := make([]string, len(items))
		for i := range items {
			if string(items[i]) == "null" {
				values[i] = "None"
				continue
			}
			value, err := rustLiteral(items[i], "int")
			if err != nil {
				return "", err
			}
			values[i] = "Some(" + value + ")"
		}
		return "__jl_tree(vec![" + strings.Join(values, ", ") + "])", nil
	case "graphnode", "graphnode?", "optional[graphnode]":
		var rows []json.RawMessage
		if err := json.Unmarshal(raw, &rows); err != nil {
			return "", fmt.Errorf("expected adjacency array for GraphNode")
		}
		values := make([]string, len(rows))
		for i := range rows {
			value, err := rustLiteral(rows[i], "int[]")
			if err != nil {
				return "", err
			}
			values[i] = value
		}
		return "__jl_graph(vec![" + strings.Join(values, ", ") + "])", nil
	default:
		return "", fmt.Errorf("Rust harness does not support type %q yet", rawType)
	}
}

func rustType(raw string) (string, error) {
	t := normalizeType(raw)
	if strings.HasSuffix(t, "[]") {
		child, err := rustType(strings.TrimSuffix(t, "[]"))
		if err != nil {
			return "", err
		}
		return "Vec<" + child + ">", nil
	}
	switch strings.ToLower(t) {
	case "int":
		return "i32", nil
	case "int64", "long":
		return "i64", nil
	case "float", "double":
		return "f64", nil
	case "bool":
		return "bool", nil
	case "string":
		return "String", nil
	case "void":
		return "()", nil
	case "listnode", "listnode?", "optional[listnode]":
		return "Option<Box<ListNode>>", nil
	case "treenode", "treenode?", "optional[treenode]":
		return "Option<Rc<RefCell<TreeNode>>>", nil
	case "graphnode", "graphnode?", "optional[graphnode]":
		return "Option<Rc<RefCell<Node>>>", nil
	default:
		return "", fmt.Errorf("Rust harness does not support type %q yet", raw)
	}
}

func rustString(value string) string {
	var out strings.Builder
	out.WriteByte('"')
	for _, r := range value {
		switch r {
		case '\\':
			out.WriteString(`\\`)
		case '"':
			out.WriteString(`\"`)
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		default:
			if unicode.IsControl(r) {
				fmt.Fprintf(&out, `\u{%x}`, r)
			} else {
				out.WriteRune(r)
			}
		}
	}
	out.WriteByte('"')
	return out.String()
}

func normalizeType(raw string) string {
	value := strings.ReplaceAll(strings.TrimSpace(raw), " ", "")
	replacements := map[string]string{
		"List[int]":       "int[]",
		"list[int]":       "int[]",
		"List[str]":       "string[]",
		"list[str]":       "string[]",
		"List[bool]":      "bool[]",
		"list[bool]":      "bool[]",
		"List[List[int]]": "int[][]",
		"list[list[int]]": "int[][]",
		"List[List[str]]": "string[][]",
		"list[list[str]]": "string[][]",
		"None":            "void",
		"NoneType":        "void",
		"str":             "string",
		"boolean":         "bool",
	}
	if replacement, ok := replacements[value]; ok {
		return replacement
	}
	return value
}

func structuredType(raw string) string {
	switch strings.ToLower(normalizeType(raw)) {
	case "listnode", "listnode?", "optional[listnode]":
		return "listnode"
	case "treenode", "treenode?", "optional[treenode]":
		return "treenode"
	case "graphnode", "graphnode?", "optional[graphnode]":
		return "graphnode"
	default:
		return ""
	}
}

func paramTypeNames(params []domain.ExecutionParam) []string {
	types := make([]string, len(params))
	for i, param := range params {
		types[i] = param.Type
	}
	return types
}

func methodBinding(name string, method domain.MethodSpec, language domain.Language) string {
	if binding := method.Bindings[string(language)]; binding != "" {
		return binding
	}
	switch language {
	case domain.LanguageGo:
		if name == "" {
			return ""
		}
		return strings.ToUpper(name[:1]) + name[1:]
	case domain.LanguageRust:
		return snakeCase(name)
	default:
		return name
	}
}

func snakeCase(value string) string {
	var out strings.Builder
	for i, r := range value {
		if unicode.IsUpper(r) {
			if i > 0 {
				out.WriteByte('_')
			}
			out.WriteRune(unicode.ToLower(r))
		} else {
			out.WriteRune(r)
		}
	}
	return out.String()
}

func indent(value, prefix string) string {
	var out bytes.Buffer
	for _, line := range strings.Split(value, "\n") {
		if line != "" {
			out.WriteString(prefix)
		}
		out.WriteString(line)
		out.WriteByte('\n')
	}
	return out.String()
}
