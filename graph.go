package pkg

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// dump direct and indirect dependencies packages of all packages.
func (depTree *DependencyTree) MarshalGraph(writer io.Writer) error {
	// dump direct and indirect dependencies of each package.
	pkgTraversalFlag := make(map[string]bool)
	err := depTree.TraversalDeep(func(tree *DependencyTree) error {
		if _, ok := (pkgTraversalFlag)[tree.Context.PackageName]; ok {
			return nil // skip
		}
		p, err := tree.ListDepsName()
		if err != nil {
			return err
		}
		// write package name
		if _, err := writer.Write(([]byte)(tree.Context.PackageName + ": ")); err != nil {
			return err
		}
		// dump dependencies list to file.
		for k, v := range p {
			if _, err := writer.Write([]byte(v)); err != nil {
				return err
			}
			if k != len(p)-1 { // not the last one
				if _, err := writer.Write([]byte(", ")); err != nil {
					return err
				}
			}
		}
		// new line
		if _, err := writer.Write([]byte("\n")); err != nil {
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

func LoadListFromGraph(graphPath, packageName string) ([]string, error) {
	if file, err := os.Open(graphPath); err != nil {
		return nil, err
	} else {
		defer file.Close()
		reader := bufio.NewReader(file)
		for {
			if line, _, err := reader.ReadLine(); err != nil {
				if err == io.EOF {
					return nil, fmt.Errorf("package `%s` is not fount in grapg file", packageName)
				} else {
					return nil, err
				}
			} else {
				sp := strings.SplitN(string(line), ":", 2)
				if len(sp) != 2 {
					return nil, fmt.Errorf("error format of graph file: %s", graphPath)
				}
				if strings.TrimSpace(sp[0]) == packageName {
					l := strings.Split(sp[1], ",")
					lists := make([]string, 0, len(l))
					for i := 0; i < len(l); i++ {
						if li := strings.TrimSpace(l[i]); li != "" {
							lists = append(lists, li)
						}
					}
					return lists, nil
				}
			}
		}
	}
}
