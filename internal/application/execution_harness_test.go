package application

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
	"github.com/tuannm99/judge-loop/internal/infrastructure/sandbox"
)

func TestBundledRegistryHarnessesRender(t *testing.T) {
	workingDirectory, err := os.Getwd()
	require.NoError(t, err)
	path := filepath.Join(
		workingDirectory,
		"..",
		"..",
		"registry",
		"providers",
		"leetcode",
		"free",
		"problems.json",
	)
	body, err := os.ReadFile(path)
	require.NoError(t, err)
	var manifest struct {
		Problems []domain.ProblemManifest `json:"problems"`
	}
	require.NoError(t, json.Unmarshal(body, &manifest))

	rendered := 0
	for _, problem := range manifest.Problems {
		if !problem.JudgeReady {
			continue
		}
		for _, testCase := range problem.TestCases {
			tc := domain.TestCase{
				Input:        testCase.Input,
				Expected:     testCase.Expected,
				InputJSON:    testCase.InputJSON,
				ExpectedJSON: testCase.ExpectedJSON,
			}
			for _, language := range problem.ExecutionSpec.SupportedLanguages {
				code := problem.StarterCode[string(language)]
				_, err := renderExecutionHarness(
					&domain.Submission{Language: language, Code: code},
					problem.ExecutionSpec,
					tc,
				)
				require.NoError(t, err, "%s/%s/%s", problem.Slug, testCase.Name, language)
				rendered++
			}
		}
	}
	require.Greater(t, rendered, 0)
}

func TestFunctionHarnessesExecuteAcrossLanguages(t *testing.T) {
	spec := domain.ExecutionSpec{
		Mode:       domain.ExecutionModeFunction,
		Entrypoint: "twoSum",
		Signature: domain.ExecutionSignature{
			Params: []domain.ExecutionParam{
				{Name: "nums", Type: "int[]"},
				{Name: "target", Type: "int"},
			},
			Returns: "int[]",
		},
	}
	tc := domain.TestCase{InputJSON: []byte(`{"args":[[2,7,11,15],9]}`)}
	sources := map[domain.Language]string{
		domain.LanguagePython: `class Solution:
    def twoSum(self, nums, target):
        return [0, 1]`,
		domain.LanguageGo: `func twoSum(nums []int, target int) []int {
	return []int{0, 1}
}`,
		domain.LanguageJavascript: `var twoSum = function(nums, target) {
  return [0, 1]
};`,
		domain.LanguageTypescript: `function twoSum(nums: number[], target: number): number[] {
  return [0, 1]
}`,
		domain.LanguageRust: `impl Solution {
    pub fn two_sum(nums: Vec<i32>, target: i32) -> Vec<i32> {
        vec![0, 1]
    }
}`,
	}

	for language, source := range sources {
		t.Run(string(language), func(t *testing.T) {
			sub := &domain.Submission{Language: language, Code: source}
			req, err := buildRunRequest(sub, spec, tc)
			require.NoError(t, err)
			output := runGeneratedHarness(t, language, req.Code)
			require.JSONEq(t, `[0,1]`, output)
		})
	}
}

func TestClassAndInPlaceHarnessesRender(t *testing.T) {
	classSpec := domain.ExecutionSpec{
		Mode:      domain.ExecutionModeClass,
		ClassName: "Counter",
		Constructor: domain.ExecutionSignature{
			Params: []domain.ExecutionParam{{Name: "value", Type: "int"}},
		},
		Methods: map[string]domain.MethodSpec{
			"add": {
				Params:  []domain.ExecutionParam{{Name: "value", Type: "int"}},
				Returns: "int",
			},
			"reset": {Returns: "void"},
		},
	}
	classCase := domain.TestCase{
		InputJSON: []byte(`{"constructor":[1],"calls":[{"method":"add","args":[2]},{"method":"reset","args":[]}]}`),
	}
	for _, language := range []domain.Language{
		domain.LanguagePython,
		domain.LanguageGo,
		domain.LanguageJavascript,
		domain.LanguageTypescript,
		domain.LanguageRust,
	} {
		t.Run("class_"+string(language), func(t *testing.T) {
			req, err := renderExecutionHarness(
				&domain.Submission{Language: language, Code: classSource(language)},
				classSpec,
				classCase,
			)
			require.NoError(t, err, language)
			output := runGeneratedHarness(t, language, req.Code)
			require.JSONEq(t, `[3,null]`, output, language)
		})
	}

	inPlace := domain.ExecutionSpec{
		Mode:       domain.ExecutionModeInPlace,
		Entrypoint: "rotate",
		Signature: domain.ExecutionSignature{
			Params: []domain.ExecutionParam{
				{Name: "nums", Type: "int[]"},
				{Name: "k", Type: "int"},
			},
			Returns: "void",
		},
		Output: domain.ExecutionOutput{Source: "param", ParamIndex: 0},
	}
	req, err := renderExecutionHarness(
		&domain.Submission{
			Language: domain.LanguageRust,
			Code: `impl Solution {
    pub fn rotate(nums: &mut Vec<i32>, k: i32) {}
}`,
		},
		inPlace,
		domain.TestCase{InputJSON: []byte(`{"args":[[1,2,3],1]}`)},
	)
	require.NoError(t, err)
	require.JSONEq(t, `[1,2,3]`, runGeneratedHarness(t, domain.LanguageRust, req.Code))
}

func TestListNodeCodecAcrossLanguages(t *testing.T) {
	spec := domain.ExecutionSpec{
		Mode:       domain.ExecutionModeFunction,
		Entrypoint: "reverseList",
		Signature: domain.ExecutionSignature{
			Params:  []domain.ExecutionParam{{Name: "head", Type: "ListNode"}},
			Returns: "ListNode",
		},
	}
	tc := domain.TestCase{InputJSON: []byte(`{"args":[[1,2,3]]}`)}
	sources := map[domain.Language]string{
		domain.LanguagePython: `class Solution:
    def reverseList(self, head):
        previous = None
        while head:
            following = head.next
            head.next = previous
            previous = head
            head = following
        return previous`,
		domain.LanguageGo: `func reverseList(head *ListNode) *ListNode {
	var previous *ListNode
	for head != nil {
		following := head.Next
		head.Next = previous
		previous = head
		head = following
	}
	return previous
}`,
		domain.LanguageJavascript: `var reverseList = function(head) {
  let previous = null
  while (head) {
    const following = head.next
    head.next = previous
    previous = head
    head = following
  }
  return previous
};`,
		domain.LanguageRust: `impl Solution {
    pub fn reverse_list(mut head: Option<Box<ListNode>>) -> Option<Box<ListNode>> {
        let mut previous = None;
        while let Some(mut node) = head {
            head = node.next.take();
            node.next = previous;
            previous = Some(node);
        }
        previous
    }
}`,
	}
	for language, source := range sources {
		t.Run(string(language), func(t *testing.T) {
			req, err := renderExecutionHarness(
				&domain.Submission{Language: language, Code: source},
				spec,
				tc,
			)
			require.NoError(t, err)
			require.JSONEq(t, `[3,2,1]`, runGeneratedHarness(t, language, req.Code))
		})
	}
}

func TestGraphNodeCodecAcrossLanguages(t *testing.T) {
	spec := domain.ExecutionSpec{
		Mode:       domain.ExecutionModeFunction,
		Entrypoint: "cloneGraph",
		Signature: domain.ExecutionSignature{
			Params:  []domain.ExecutionParam{{Name: "node", Type: "GraphNode"}},
			Returns: "GraphNode",
		},
	}
	tc := domain.TestCase{InputJSON: []byte(`{"args":[[[2,4],[1,3],[2,4],[1,3]]]}`)}
	sources := map[domain.Language]string{
		domain.LanguagePython: `class Solution:
    def cloneGraph(self, node):
        return node`,
		domain.LanguageGo:         `func cloneGraph(node *Node) *Node { return node }`,
		domain.LanguageJavascript: `var cloneGraph = function(node) { return node };`,
		domain.LanguageRust: `impl Solution {
    pub fn clone_graph(node: Option<std::rc::Rc<std::cell::RefCell<Node>>>) -> Option<std::rc::Rc<std::cell::RefCell<Node>>> {
        node
    }
}`,
	}
	for language, source := range sources {
		t.Run(string(language), func(t *testing.T) {
			req, err := renderExecutionHarness(
				&domain.Submission{Language: language, Code: source},
				spec,
				tc,
			)
			require.NoError(t, err)
			require.JSONEq(t, `[[2,4],[1,3],[2,4],[1,3]]`,
				runGeneratedHarness(t, language, req.Code))
		})
	}
}

func TestTypeScriptHarnessInSandbox(t *testing.T) {
	if os.Getenv("JUDGE_LOOP_DOCKER_TESTS") == "" {
		t.Skip("set JUDGE_LOOP_DOCKER_TESTS=1 to run container smoke tests")
	}
	spec := domain.ExecutionSpec{
		Mode:       domain.ExecutionModeFunction,
		Entrypoint: "twoSum",
		Signature: domain.ExecutionSignature{
			Params: []domain.ExecutionParam{
				{Name: "nums", Type: "int[]"},
				{Name: "target", Type: "int"},
			},
			Returns: "int[]",
		},
		MemoryMB: 128,
	}
	req, err := renderExecutionHarness(
		&domain.Submission{
			Language: domain.LanguageTypescript,
			Code: `function twoSum(nums: number[], target: number): number[] {
  return [0, 1]
}`,
		},
		spec,
		domain.TestCase{InputJSON: []byte(`{"args":[[2,7],9]}`)},
	)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result, err := sandbox.Run(ctx, sandbox.RunRequest{
		Language: req.Language,
		Code:     req.Code,
		MemoryMB: req.MemoryMB,
	})
	require.NoError(t, err)
	require.False(t, result.TimedOut, result.Stderr)
	require.Zero(t, result.ExitCode, result.Stderr)
	require.JSONEq(t, `[0,1]`, result.Output)
}

func runGeneratedHarness(t *testing.T, language domain.Language, code string) string {
	t.Helper()
	command := map[domain.Language]string{
		domain.LanguagePython:     "python3",
		domain.LanguageGo:         "go",
		domain.LanguageJavascript: "node",
		domain.LanguageTypescript: "deno",
		domain.LanguageRust:       "rustc",
	}[language]
	if _, err := exec.LookPath(command); err != nil {
		t.Skipf("%s is not installed", command)
	}

	dir := t.TempDir()
	var cmd *exec.Cmd
	switch language {
	case domain.LanguagePython:
		path := filepath.Join(dir, "solution.py")
		require.NoError(t, os.WriteFile(path, []byte(code), 0o600))
		cmd = exec.Command("python3", path)
	case domain.LanguageGo:
		path := filepath.Join(dir, "main.go")
		require.NoError(t, os.WriteFile(path, []byte(code), 0o600))
		cmd = exec.Command("go", "run", path)
	case domain.LanguageJavascript:
		path := filepath.Join(dir, "solution.js")
		require.NoError(t, os.WriteFile(path, []byte(code), 0o600))
		cmd = exec.Command("node", path)
	case domain.LanguageTypescript:
		path := filepath.Join(dir, "solution.ts")
		require.NoError(t, os.WriteFile(path, []byte(code), 0o600))
		cmd = exec.Command("deno", "run", "--quiet", "--no-check", path)
	case domain.LanguageRust:
		path := filepath.Join(dir, "main.rs")
		binary := filepath.Join(dir, "solution")
		require.NoError(t, os.WriteFile(path, []byte(code), 0o600))
		compile := exec.Command("rustc", path, "-o", binary)
		output, err := compile.CombinedOutput()
		require.NoError(t, err, string(output))
		cmd = exec.Command(binary)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	return strings.TrimSpace(string(output))
}

func classSource(language domain.Language) string {
	switch language {
	case domain.LanguagePython:
		return `class Counter:
    def __init__(self, value):
        self.value = value
    def add(self, value):
        self.value += value
        return self.value
    def reset(self):
        self.value = 0`
	case domain.LanguageGo:
		return `type Counter struct { value int }
func Constructor(value int) Counter { return Counter{value: value} }
func (c *Counter) Add(value int) int { c.value += value; return c.value }
func (c *Counter) Reset() { c.value = 0 }`
	case domain.LanguageJavascript, domain.LanguageTypescript:
		return `class Counter {
  value
  constructor(value) { this.value = value }
  add(value) { this.value += value; return this.value }
  reset() { this.value = 0 }
}`
	case domain.LanguageRust:
		return `struct Counter { value: i32 }
impl Counter {
    fn new(value: i32) -> Self { Self { value } }
    fn add(&mut self, value: i32) -> i32 { self.value += value; self.value }
    fn reset(&mut self) { self.value = 0; }
}`
	default:
		return ""
	}
}
