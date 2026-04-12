-- Seed data: sample problems for development and testing
-- Run after the full Postgres migration set (through 005)

BEGIN;

-- ============================================================
-- PROBLEMS
-- ============================================================

WITH seed_problems (
  slug,
  title,
  difficulty,
  tags,
  legacy_tags,
  provider,
  external_id,
  source_url,
  estimated_time
) AS (
VALUES

-- Arrays & Hashing
('two-sum',
 'Two Sum',
 'easy',
 ARRAY['array', 'hash-table'],
 ARRAY['lookup', 'hash-map'],
 'leetcode', '1',
 'https://leetcode.com/problems/two-sum/',
 15),

('best-time-to-buy-and-sell-stock',
 'Best Time to Buy and Sell Stock',
 'easy',
 ARRAY['array', 'dynamic-programming'],
 ARRAY['sliding-window', 'greedy'],
 'leetcode', '121',
 'https://leetcode.com/problems/best-time-to-buy-and-sell-stock/',
 15),

('contains-duplicate',
 'Contains Duplicate',
 'easy',
 ARRAY['array', 'hash-table', 'sorting'],
 ARRAY['hash-map'],
 'leetcode', '217',
 'https://leetcode.com/problems/contains-duplicate/',
 10),

('product-of-array-except-self',
 'Product of Array Except Self',
 'medium',
 ARRAY['array', 'prefix-sum'],
 ARRAY['prefix-sum'],
 'leetcode', '238',
 'https://leetcode.com/problems/product-of-array-except-self/',
 25),

('maximum-subarray',
 'Maximum Subarray',
 'medium',
 ARRAY['array', 'dynamic-programming', 'divide-and-conquer'],
 ARRAY['kadane', 'dynamic-programming'],
 'leetcode', '53',
 'https://leetcode.com/problems/maximum-subarray/',
 20),

-- Two Pointers
('valid-palindrome',
 'Valid Palindrome',
 'easy',
 ARRAY['two-pointers', 'string'],
 ARRAY['two-pointer'],
 'leetcode', '125',
 'https://leetcode.com/problems/valid-palindrome/',
 15),

('three-sum',
 '3Sum',
 'medium',
 ARRAY['array', 'two-pointers', 'sorting'],
 ARRAY['two-pointer', 'sorting'],
 'leetcode', '15',
 'https://leetcode.com/problems/3sum/',
 30),

('container-with-most-water',
 'Container With Most Water',
 'medium',
 ARRAY['array', 'two-pointers', 'greedy'],
 ARRAY['two-pointer'],
 'leetcode', '11',
 'https://leetcode.com/problems/container-with-most-water/',
 25),

-- Sliding Window
('longest-substring-without-repeating-characters',
 'Longest Substring Without Repeating Characters',
 'medium',
 ARRAY['hash-table', 'string', 'sliding-window'],
 ARRAY['sliding-window'],
 'leetcode', '3',
 'https://leetcode.com/problems/longest-substring-without-repeating-characters/',
 25),

('minimum-window-substring',
 'Minimum Window Substring',
 'hard',
 ARRAY['hash-table', 'string', 'sliding-window'],
 ARRAY['sliding-window'],
 'leetcode', '76',
 'https://leetcode.com/problems/minimum-window-substring/',
 45),

-- Stack
('valid-parentheses',
 'Valid Parentheses',
 'easy',
 ARRAY['string', 'stack'],
 ARRAY['stack'],
 'leetcode', '20',
 'https://leetcode.com/problems/valid-parentheses/',
 15),

('min-stack',
 'Min Stack',
 'medium',
 ARRAY['stack', 'design'],
 ARRAY['stack'],
 'leetcode', '155',
 'https://leetcode.com/problems/min-stack/',
 20),

-- Binary Search
('binary-search',
 'Binary Search',
 'easy',
 ARRAY['array', 'binary-search'],
 ARRAY['binary-search'],
 'leetcode', '704',
 'https://leetcode.com/problems/binary-search/',
 10),

('find-minimum-in-rotated-sorted-array',
 'Find Minimum in Rotated Sorted Array',
 'medium',
 ARRAY['array', 'binary-search'],
 ARRAY['binary-search'],
 'leetcode', '153',
 'https://leetcode.com/problems/find-minimum-in-rotated-sorted-array/',
 25),

-- Linked List
('reverse-linked-list',
 'Reverse Linked List',
 'easy',
 ARRAY['linked-list', 'recursion'],
 ARRAY['linked-list', 'iteration'],
 'leetcode', '206',
 'https://leetcode.com/problems/reverse-linked-list/',
 15),

('merge-two-sorted-lists',
 'Merge Two Sorted Lists',
 'easy',
 ARRAY['linked-list', 'recursion'],
 ARRAY['linked-list', 'merge'],
 'leetcode', '21',
 'https://leetcode.com/problems/merge-two-sorted-lists/',
 20),

('linked-list-cycle',
 'Linked List Cycle',
 'easy',
 ARRAY['hash-table', 'linked-list', 'two-pointers'],
 ARRAY['fast-slow-pointer'],
 'leetcode', '141',
 'https://leetcode.com/problems/linked-list-cycle/',
 15),

-- Trees
('maximum-depth-of-binary-tree',
 'Maximum Depth of Binary Tree',
 'easy',
 ARRAY['tree', 'dfs', 'bfs', 'binary-tree'],
 ARRAY['dfs', 'recursion'],
 'leetcode', '104',
 'https://leetcode.com/problems/maximum-depth-of-binary-tree/',
 15),

('invert-binary-tree',
 'Invert Binary Tree',
 'easy',
 ARRAY['tree', 'dfs', 'bfs', 'binary-tree'],
 ARRAY['dfs', 'recursion'],
 'leetcode', '226',
 'https://leetcode.com/problems/invert-binary-tree/',
 15),

('lowest-common-ancestor-of-a-binary-search-tree',
 'Lowest Common Ancestor of a Binary Search Tree',
 'medium',
 ARRAY['tree', 'dfs', 'binary-search-tree', 'binary-tree'],
 ARRAY['bst', 'recursion'],
 'leetcode', '235',
 'https://leetcode.com/problems/lowest-common-ancestor-of-a-binary-search-tree/',
 20),

-- Dynamic Programming
('climbing-stairs',
 'Climbing Stairs',
 'easy',
 ARRAY['math', 'dynamic-programming', 'memoization'],
 ARRAY['dynamic-programming', 'fibonacci'],
 'leetcode', '70',
 'https://leetcode.com/problems/climbing-stairs/',
 15),

('coin-change',
 'Coin Change',
 'medium',
 ARRAY['array', 'dynamic-programming', 'breadth-first-search'],
 ARRAY['dynamic-programming', 'unbounded-knapsack'],
 'leetcode', '322',
 'https://leetcode.com/problems/coin-change/',
 30),

('longest-common-subsequence',
 'Longest Common Subsequence',
 'medium',
 ARRAY['string', 'dynamic-programming'],
 ARRAY['dynamic-programming', '2d-dp'],
 'leetcode', '1143',
 'https://leetcode.com/problems/longest-common-subsequence/',
 35),

-- Graphs
('number-of-islands',
 'Number of Islands',
 'medium',
 ARRAY['array', 'dfs', 'bfs', 'union-find', 'matrix'],
 ARRAY['dfs', 'flood-fill'],
 'leetcode', '200',
 'https://leetcode.com/problems/number-of-islands/',
 25),

('clone-graph',
 'Clone Graph',
 'medium',
 ARRAY['hash-table', 'dfs', 'bfs', 'graph'],
 ARRAY['dfs', 'bfs', 'graph'],
 'leetcode', '133',
 'https://leetcode.com/problems/clone-graph/',
 30)
),
inserted_problems AS (
  INSERT INTO problems (
    slug,
    title,
    difficulty,
    provider,
    external_id,
    source_url,
    estimated_time
  )
  SELECT
    slug,
    title,
    difficulty::difficulty,
    provider::provider,
    external_id,
    source_url,
    estimated_time
  FROM seed_problems
  RETURNING id, slug
),
expanded_labels AS (
  SELECT DISTINCT unnest(tags || legacy_tags) AS slug
  FROM seed_problems
),
inserted_labels AS (
  INSERT INTO problem_labels (slug, name)
  SELECT
    slug,
    initcap(replace(slug, '-', ' '))
  FROM expanded_labels
  ON CONFLICT (slug) DO NOTHING
  RETURNING id
)
INSERT INTO problem_label_links (problem_id, problem_label_id)
SELECT DISTINCT
  problem.id,
  label.id
FROM inserted_problems AS problem
JOIN seed_problems AS seed ON seed.slug = problem.slug
JOIN LATERAL unnest(seed.tags || seed.legacy_tags) AS tag(slug) ON true
JOIN problem_labels AS label ON label.slug = tag.slug
ON CONFLICT DO NOTHING;

-- ============================================================
-- TEST CASES (visible samples for seeded problems)
-- ============================================================

INSERT INTO test_cases (problem_id, input, expected, is_hidden, order_idx)
SELECT p.id, '{"nums": [2, 7, 11, 15], "target": 9}', '[0, 1]', FALSE, 1
FROM problems p WHERE p.slug = 'two-sum';

INSERT INTO test_cases (problem_id, input, expected, is_hidden, order_idx)
SELECT p.id, '{"nums": [3, 2, 4], "target": 6}', '[1, 2]', FALSE, 2
FROM problems p WHERE p.slug = 'two-sum';

INSERT INTO test_cases (problem_id, input, expected, is_hidden, order_idx)
SELECT p.id, '{"nums": [3, 3], "target": 6}', '[0, 1]', TRUE, 3
FROM problems p WHERE p.slug = 'two-sum';

INSERT INTO test_cases (problem_id, input, expected, is_hidden, order_idx)
SELECT p.id, '{"prices": [7, 1, 5, 3, 6, 4]}', '5', FALSE, 1
FROM problems p WHERE p.slug = 'best-time-to-buy-and-sell-stock';

INSERT INTO test_cases (problem_id, input, expected, is_hidden, order_idx)
SELECT p.id, '{"prices": [7, 6, 4, 3, 1]}', '0', FALSE, 2
FROM problems p WHERE p.slug = 'best-time-to-buy-and-sell-stock';

INSERT INTO test_cases (problem_id, input, expected, is_hidden, order_idx)
SELECT p.id, '{"s": "()[]{}"}', 'true', FALSE, 1
FROM problems p WHERE p.slug = 'valid-parentheses';

INSERT INTO test_cases (problem_id, input, expected, is_hidden, order_idx)
SELECT p.id, '{"s": "([)]"}', 'false', FALSE, 2
FROM problems p WHERE p.slug = 'valid-parentheses';

COMMIT;
