package pkg

import (
	"io"
	"os"
	"path/filepath"

	"github.com/mholt/archiver/v4"
)

// FROM https://golangcode.com/unzip-files-in-go/
// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(f archiver.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// Store filename/path for returning and using later on
	fpath := filepath.Join(dest, f.NameInArchive)
	// filenames = append(filenames, fpath)

	if f.IsDir() {
		// Make Folder
		os.MkdirAll(fpath, os.ModePerm)
	} else {
		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()

		if err != nil {
			return err
		}
	}
	// }
	return nil
}
