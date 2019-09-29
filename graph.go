package pkg

import (
	"fmt"
	"io"
)

// dump direct and indirect dependencies packages of all packages.
func (depTree *DependencyTree) MarshalGraph(writer io.Writer) error {
	// dump direct and indirect dependencies of each package.
	pkgTraversalFlag := make(map[string]bool)
	err := depTree.TraversalDeep(func(tree *DependencyTree) error {
		if _, ok := (pkgTraversalFlag)[tree.Context.PackageName]; ok {
			return nil // skip
		}
		p, err := tree.ListDeps()
		if err != nil {
			return err
		}
		// write package name
		if _, err := writer.Write(([]byte)(tree.Context.PackageName + ": ")); err != nil {
			return err
		}
		// dump dependencies list to file.
		if _, err := fmt.Fprintln(writer, p); err != nil {
			return err
		}

		pkgTraversalFlag[tree.Context.PackageName] = true
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
