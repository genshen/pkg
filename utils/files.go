package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CheckDirectoryLists(dirs ...string) error {
	for _, dir := range dirs {
		if err := CheckDir(filepath.Join(dir)); err != nil {
			return err
		}
	}
	return nil
}

// check if the dir exits,if not create it.
func CheckDir(path string) error {
	if fileInfo, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) { // directory not exists.
			if err := os.MkdirAll(path, 0744); err != nil { // todo
				return err // create dir error.
			}
		} else {
			return err
		}
	} else if !fileInfo.IsDir() { // if exists,but is not dir.
		return fmt.Errorf("%s is not a directory", path)
	}
	return nil
}

//// change path to relative path, replace PKG_DIR with relative path.
func RelativePath(base, target string) string {
	//	// replace absolute path with relative path.
	relPath := strings.TrimPrefix(target, base) // relative pkg path
	relPath = strings.TrimPrefix(relPath, string(filepath.Separator))
	return relPath

}
