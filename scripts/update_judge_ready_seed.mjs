#!/usr/bin/env node
import { createHash } from 'node:crypto'
import { mkdir, readFile, writeFile } from 'node:fs/promises'
import path from 'node:path'

const rootDir = path.resolve(import.meta.dirname, '..')
const manifestPath = path.join(rootDir, 'registry/providers/judge-ready/blind75.json')
const indexPath = path.join(rootDir, 'registry/index.json')

function twoSumStarter() {
  return {
    python: `import json
import sys

def two_sum(nums, target):
    return []

def main():
    case = json.loads(sys.stdin.read())
    result = two_sum(case["nums"], case["target"])
    print(json.dumps(result, separators=(",", ":")))

if __name__ == "__main__":
    main()
`,
    go: `package main

import (
	"encoding/json"
	"os"
)

type Case struct {
	Nums   []int \`json:"nums"\`
	Target int   \`json:"target"\`
}

func twoSum(nums []int, target int) []int {
	return []int{}
}

func main() {
	var c Case
	_ = json.NewDecoder(os.Stdin).Decode(&c)
	_ = json.NewEncoder(os.Stdout).Encode(twoSum(c.Nums, c.Target))
}
`,
    javascript: `const fs = require('node:fs')

function twoSum(nums, target) {
  return []
}

const input = fs.readFileSync(0, 'utf8')
const testCase = JSON.parse(input)
process.stdout.write(JSON.stringify(twoSum(testCase.nums, testCase.target)))
`,
    typescript: `function twoSum(nums: number[], target: number): number[] {
  return []
}

const input = await new Response(Deno.stdin.readable).text()
const testCase = JSON.parse(input) as { nums: number[]; target: number }
await Deno.stdout.write(
  new TextEncoder().encode(JSON.stringify(twoSum(testCase.nums, testCase.target)))
)
`,
    rust: `use std::io::{self, Read};

fn extract_ints(input: &str) -> Vec<i32> {
    let mut values = Vec::new();
    let mut current = String::new();
    for ch in input.chars() {
        if ch == '-' || ch.is_ascii_digit() {
            current.push(ch);
        } else if !current.is_empty() && current != "-" {
            if let Ok(value) = current.parse::<i32>() {
                values.push(value);
            }
            current.clear();
        } else {
            current.clear();
        }
    }
    if !current.is_empty() && current != "-" {
        if let Ok(value) = current.parse::<i32>() {
            values.push(value);
        }
    }
    values
}

fn two_sum(nums: &[i32], target: i32) -> Vec<usize> {
    Vec::new()
}

fn main() {
    let mut input = String::new();
    io::stdin().read_to_string(&mut input).unwrap();
    let values = extract_ints(&input);
    if values.is_empty() {
        print!("[]");
        return;
    }
    let target = *values.last().unwrap();
    let nums = &values[..values.len() - 1];
    let result = two_sum(nums, target);
    let body = result
        .iter()
        .map(|value| value.to_string())
        .collect::<Vec<_>>()
        .join(",");
    print!("[{}]", body);
}
`
  }
}

function containsDuplicateStarter() {
  return {
    python: `import json
import sys

def contains_duplicate(nums):
    return False

def main():
    case = json.loads(sys.stdin.read())
    print(json.dumps(contains_duplicate(case["nums"]), separators=(",", ":")))

if __name__ == "__main__":
    main()
`,
    go: `package main

import (
	"encoding/json"
	"os"
)

type Case struct {
	Nums []int \`json:"nums"\`
}

func containsDuplicate(nums []int) bool {
	return false
}

func main() {
	var c Case
	_ = json.NewDecoder(os.Stdin).Decode(&c)
	_ = json.NewEncoder(os.Stdout).Encode(containsDuplicate(c.Nums))
}
`,
    javascript: `const fs = require('node:fs')

function containsDuplicate(nums) {
  return false
}

const input = fs.readFileSync(0, 'utf8')
const testCase = JSON.parse(input)
process.stdout.write(JSON.stringify(containsDuplicate(testCase.nums)))
`,
    typescript: `function containsDuplicate(nums: number[]): boolean {
  return false
}

const input = await new Response(Deno.stdin.readable).text()
const testCase = JSON.parse(input) as { nums: number[] }
await Deno.stdout.write(
  new TextEncoder().encode(JSON.stringify(containsDuplicate(testCase.nums)))
)
`,
    rust: `use std::io::{self, Read};

fn extract_ints(input: &str) -> Vec<i32> {
    let mut values = Vec::new();
    let mut current = String::new();
    for ch in input.chars() {
        if ch == '-' || ch.is_ascii_digit() {
            current.push(ch);
        } else if !current.is_empty() && current != "-" {
            if let Ok(value) = current.parse::<i32>() {
                values.push(value);
            }
            current.clear();
        } else {
            current.clear();
        }
    }
    if !current.is_empty() && current != "-" {
        if let Ok(value) = current.parse::<i32>() {
            values.push(value);
        }
    }
    values
}

fn contains_duplicate(nums: &[i32]) -> bool {
    false
}

fn main() {
    let mut input = String::new();
    io::stdin().read_to_string(&mut input).unwrap();
    let nums = extract_ints(&input);
    print!("{}", if contains_duplicate(&nums) { "true" } else { "false" });
}
`
  }
}

const manifest = {
  provider: 'judge-loop',
  kind: 'judge-ready',
  track: 'blind75',
  updated_at: new Date().toISOString(),
  problems: [
    {
      provider: 'leetcode',
      external_id: '1',
      slug: 'two-sum',
      title: 'Two Sum',
      difficulty: 'easy',
      tags: ['array', 'hash-table'],
      source_url: 'https://leetcode.com/problems/two-sum/',
      estimated_time: 15,
      description_markdown:
        'Given an array of integers and a target value, return the indices of two distinct elements whose values sum to the target. Return the pair in ascending index order for deterministic local judging.',
      starter_code: twoSumStarter(),
      test_cases: [
        {
          input: '{"nums":[2,7,11,15],"target":9}',
          expected: '[0,1]'
        },
        {
          input: '{"nums":[3,2,4],"target":6}',
          expected: '[1,2]'
        },
        {
          input: '{"nums":[3,3],"target":6}',
          expected: '[0,1]',
          is_hidden: true
        },
        {
          input: '{"nums":[-1,-2,-3,-4,-5],"target":-8}',
          expected: '[2,4]',
          is_hidden: true
        }
      ],
      version: 1
    },
    {
      provider: 'leetcode',
      external_id: '217',
      slug: 'contains-duplicate',
      title: 'Contains Duplicate',
      difficulty: 'easy',
      tags: ['array', 'hash-table', 'sorting'],
      source_url: 'https://leetcode.com/problems/contains-duplicate/',
      estimated_time: 15,
      description_markdown:
        'Given an integer array, return true when any value appears at least twice. Return false when every value appears exactly once.',
      starter_code: containsDuplicateStarter(),
      test_cases: [
        {
          input: '{"nums":[1,2,3,1]}',
          expected: 'true'
        },
        {
          input: '{"nums":[1,2,3,4]}',
          expected: 'false'
        },
        {
          input: '{"nums":[1,1,1,3,3,4,3,2,4,2]}',
          expected: 'true',
          is_hidden: true
        },
        {
          input: '{"nums":[]}',
          expected: 'false',
          is_hidden: true
        }
      ],
      version: 1
    }
  ]
}

await mkdir(path.dirname(manifestPath), { recursive: true })
const body = `${JSON.stringify(manifest, null, 2)}\n`
await writeFile(manifestPath, body)

const checksum = createHash('sha256').update(body).digest('hex')
const index = JSON.parse(await readFile(indexPath, 'utf8'))
const ref = {
  name: 'judge-ready-blind75',
  path: 'providers/judge-ready/blind75.json',
  checksum: `sha256:${checksum}`
}
const existing = index.manifests.findIndex((item) => item.name === ref.name)
if (existing === -1) {
  index.manifests.push(ref)
} else {
  index.manifests[existing] = ref
}
index.updated_at = new Date().toISOString()
await writeFile(indexPath, `${JSON.stringify(index, null, 2)}\n`)

console.log(`Updated ${manifestPath}`)
console.log(`Judge-ready problems: ${manifest.problems.length}`)
console.log(`Checksum: sha256:${checksum}`)
