package tagging

import (
	"sort"
	"strings"
)

var keywordMap = map[string][]string{
	"brute force":               {"brute force"},
	"constructive algorithms":   {"constructive algorithms", "constructive"},
	"divide and conquer":        {"divide and conquer"},
	"dp":                        {"dp", "dynamic programming"},
	"greedy":                    {"greedy"},
	"implementation":            {"implementation", "simulate", "simulation"},
	"interactive":               {"interactive"},
	"chinese remainder theorem": {"chinese remainder theorem", "chinese remainder", "crt"},
	"combinatorics":             {"combinatorics", "combinatorial"},
	"games":                     {"game theory", "nim", "grundy", "sprague-grundy", "minimax", "winning strategy"},
	"geometry":                  {"geometry", "coordinate", "convex hull", "polygon", "point", "distance"},
	"math":                      {"math", "gcd", "g.c.d", "lcm", "l.c.m", "modulo", "modular", "prime", "factorial", "divisible", "divisibility"},
	"matrices":                  {"matrix", "matrices", "matrix exponentiation"},
	"number theory":             {"number theory", "prime number", "divisor", "sieve", "euler totient", "totient"},
	"probabilities":             {"probability", "probabilities", "expected value", "random"},
	"binary search":             {"binary search"},
	"meet-in-the-middle":        {"meet in the middle", "meet-in-the-middle"},
	"sorting":                   {"sorting", "sort", "merge sort", "quick sort", "bubble sort", "heap sort"},
	"ternary search":            {"ternary search"},
	"two pointers":              {"two pointers", "two pointer", "2 pointers", "2 pointer", "two-pointers"},
	"data structures":           {"data structures", "segment tree", "fenwick tree", "binary indexed tree", "treap", "splay tree", "heap", "priority queue"},
	"dsu":                       {"dsu", "disjoint set union", "union find", "disjoint set"},
	"expression parsing":        {"expression parsing", "infix", "postfix", "polish notation", "shunting yard"},
	"hashing":                   {"hashing", "hash", "rolling hash", "hash table", "hash map", "hashset"},
	"string suffix structures":  {"suffix array", "suffix automaton", "suffix tree"},
	"strings":                   {"string", "string matching", "kmp", "z-algorithm", "z function", "trie", "palindrome", "anagram", "substring", "subsequence"},
	"2-sat":                     {"2-sat", "2sat", "2 satisfiability"},
	"dfs and similar":           {"dfs", "depth first search", "depth-first", "bfs", "breadth first search", "breadth-first"},
	"flows":                     {"flow", "max flow", "min cost max flow", "max-flow", "min-cost max-flow", "dinic", "edmonds-karp", "ford-fulkerson"},
	"graph matchings":           {"matching", "bipartite matching", "hopcroft-karp", "könig"},
	"graphs":                    {"graph", "graph theory", "directed", "undirected", "adjacency", "edge", "vertex", "vertices", "nodes"},
	"shortest paths":            {"shortest path", "dijkstra", "bellman-ford", "floyd-warshall", "floyd warshall"},
	"trees":                     {"tree", "binary tree", "bst", "tree diameter", "lowest common ancestor", "lca"},
	"bitmasks":                  {"bitmask", "bitmasks", "bit manipulation", "bitwise", "xor", "bitwise or", "bitwise and"},
	"fft":                       {"fft", "fast fourier transform", "number theoretic transform", "ntt", "convolution"},
	"schedules":                 {"schedule", "scheduling"},
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func containsWord(text, word string) bool {
	idx := strings.Index(text, word)
	for idx != -1 {
		beforeOK := idx == 0 || !isLetter(text[idx-1])
		afterOK := idx+len(word) == len(text) || !isLetter(text[idx+len(word)])
		if beforeOK && afterOK {
			return true
		}
		next := strings.Index(text[idx+1:], word)
		if next == -1 {
			break
		}
		idx += 1 + next
	}
	return false
}

// GetAllTags returns the sorted list of all canonical tag names.
func GetAllTags() []string {
	tags := make([]string, 0, len(keywordMap))
	for tag := range keywordMap {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

// DetectTags scans a problem name + description and returns matched Codeforces-style tags.
func DetectTags(name, description string) []string {
	text := strings.ToLower(name + " " + description)
	seen := make(map[string]struct{})
	for tag, keywords := range keywordMap {
		for _, kw := range keywords {
			found := false
			if len(kw) <= 4 {
				found = containsWord(text, kw)
			} else {
				found = strings.Contains(text, kw)
			}
			if found {
				seen[tag] = struct{}{}
				break
			}
		}
	}
	tags := make([]string, 0, len(seen))
	for tag := range seen {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}
