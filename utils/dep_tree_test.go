package utils

import (
	"testing"
)

func TestDependencyTree_Traversal(t *testing.T) {
	var root, l11, l12, l21, l22, l23, l24 DependencyTree
	l21.Context.PackageName = "c"
	l22.Context.PackageName = "d"
	l23.Context.PackageName = "e"
	l24.Context.PackageName = "f"
	l11.Context.PackageName = "a"
	l12.Context.PackageName = "b"
	root.Context.PackageName = "r"

	l11.Dependencies = append(l11.Dependencies, &l21, &l22)
	l12.Dependencies = append(l12.Dependencies, &l23, &l24)
	root.Dependencies = append(root.Dependencies, &l11, &l12)
	//root -> {l11 -> {l21, l22}, l12 -> {l23, l24}}
	var tstr = ""
	root.Traversal(func(tree *DependencyTree) bool {
		tstr += tree.Context.PackageName
		return true
	})

	if tstr != "racdbef" {
		t.Error("error traversal of dependency tree")
	}
}
