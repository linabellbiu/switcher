package test

import (
	"golang.org/x/mod/modfile"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnv(t *testing.T) {
	srcDir := "githup/switcher/parse"
	moduleMode := os.Getenv("GO111MODULE")
	// trying to find the module
	if moduleMode != "off" {
		currentDir := srcDir
		for {
			dat, err := ioutil.ReadFile(filepath.Join(currentDir, "go.mod"))
			if os.IsNotExist(err) {
				if currentDir == filepath.Dir(currentDir) {
					// at the root
					break
				}
				currentDir = filepath.Dir(currentDir)
				continue
			} else if err != nil {
				t.Fatal(err)
			}
			modulePath := modfile.ModulePath(dat)
			t.Log(filepath.ToSlash(filepath.Join(modulePath, strings.TrimPrefix(srcDir, currentDir))))
			return
		}
	}
	// fall back to GOPATH mode
	goPaths := os.Getenv("GOPATH")
	if goPaths == "" {
		t.Fatal("GOPATH is not set")
	}
	goPathList := strings.Split(goPaths, string(os.PathListSeparator))
	for _, goPath := range goPathList {
		sourceRoot := filepath.Join(goPath, "src") + string(os.PathSeparator)
		if strings.HasPrefix(srcDir, sourceRoot) {
			t.Log(filepath.ToSlash(strings.TrimPrefix(srcDir, sourceRoot)))
		}
	}

	return
}
