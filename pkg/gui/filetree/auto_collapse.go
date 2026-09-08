package filetree

import (
	"path"
	"strings"

	"github.com/jesseduffield/generics/set"
)

// autoCollapseDirs collapses the directories matching the configured patterns.
// Each directory is collapsed only the first time it appears (tracked in
// seen), so a directory the user has expanded stays expanded across refreshes.
func autoCollapseDirs[T any](root *Node[T], patterns []string, collapsedPaths *CollapsedPaths, seen *set.Set[string]) {
	if len(patterns) == 0 || root == nil {
		return
	}

	var walk func(node *Node[T])
	walk = func(node *Node[T]) {
		for _, child := range node.Children {
			if child.IsFile() {
				continue
			}
			if matchesAutoCollapsePattern(patterns, child.GetPath()) && !seen.Includes(child.GetInternalPath()) {
				seen.Add(child.GetInternalPath())
				collapsedPaths.Collapse(child.GetInternalPath())
			}
			walk(child)
		}
	}
	walk(root)
}

// A pattern containing a slash is matched as a glob against the directory's
// whole path; otherwise it is matched against the directory's name, like
// gitignore.
func matchesAutoCollapsePattern(patterns []string, dirPath string) bool {
	for _, pattern := range patterns {
		candidate := path.Base(dirPath)
		if strings.ContainsRune(pattern, '/') {
			candidate = dirPath
		}
		if matched, err := path.Match(pattern, candidate); err == nil && matched {
			return true
		}
	}
	return false
}
