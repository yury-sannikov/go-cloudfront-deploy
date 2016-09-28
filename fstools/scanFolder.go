package fstools

import (
	"os"
	"path/filepath"
	"strings"

	ld "github.com/ahl5esoft/golang-underscore"
)

// ReadFiles reads specified folder recursively and return array of file names
func ReadFiles(fsPath string) ([]string, error) {

	result := make([]string, 0, 128)

	walker := func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}

		if strings.HasPrefix(f.Name(), ".") {
			return nil
		}

		pathList := filepath.SplitList(path)
		idx := ld.FindIndex(pathList, func(n string, _ int) bool {
			return strings.HasPrefix(n, ".")
		})
		if idx != -1 {
			return nil
		}
		result = append(result, path)
		return nil
	}
	err := filepath.Walk(fsPath, walker)
	return result, err
}
