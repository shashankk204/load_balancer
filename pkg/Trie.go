package core

import (
	"strings"
)


type TrieNode struct {
	children map[string]*TrieNode
	isEnd    bool
}

type Trie struct {
	root *TrieNode
}


func NewTrie() *Trie {
	return &Trie{
		root: &TrieNode{children: make(map[string]*TrieNode)},
	}
}


func (t *Trie) Insert(path string) {
	segments := splitPath(path)
	node := t.root

	for _, seg := range segments {
		if _, ok := node.children[seg]; !ok {
			node.children[seg] = &TrieNode{children: make(map[string]*TrieNode)}
		}
		node = node.children[seg]
	}
	node.isEnd = true
}


func (t *Trie) MatchPrefix(url string) (matched bool, matchedPath string) {
	segments := splitPath(url)
	node := t.root
	var pathParts []string

	for _, seg := range segments {
		next, ok := node.children[seg]
		if !ok {
			break
		}
		node = next
		pathParts = append(pathParts, seg)
		if node.isEnd {
			matched = true
			matchedPath = "/" + strings.Join(pathParts, "/")
			break;
		}
	}

	return matched, matchedPath
}


func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

