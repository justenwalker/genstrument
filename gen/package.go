package main

import (
	"fmt"
	"golang.org/x/mod/modfile"
	"os"
	"path/filepath"
	"strings"
)

type fullPackagePath struct {
	moduleName  string
	packagePath string
}

func getFullPackagePath(srcDir string) (fullPackagePath, error) {
	if mm := os.Getenv("GO111MODULE"); mm != "off" {
		currentDir := srcDir
		for {
			modData, err := os.ReadFile(filepath.Join(currentDir, "go.mod"))
			if os.IsNotExist(err) {
				if srcDir == filepath.Dir(srcDir) { // reached root
					break
				}
				currentDir = filepath.Dir(currentDir) // go up one level
				continue
			}
			if err != nil {
				return fullPackagePath{}, err
			}
			moduleName := modfile.ModulePath(modData)
			return fullPackagePath{
				moduleName:  moduleName,
				packagePath: filepath.ToSlash(filepath.Join(moduleName, strings.TrimPrefix(srcDir, currentDir))),
			}, nil

		}
	}
	goPaths := os.Getenv("GOPATH")
	if goPaths == "" {
		return fullPackagePath{}, fmt.Errorf("GOPATH not set")
	}
	goPathList := strings.Split(goPaths, string(os.PathListSeparator))
	for _, goPath := range goPathList {
		goSrc := filepath.Join(goPath, "src") + string(os.PathSeparator)
		if strings.HasPrefix(goSrc, srcDir) {
			return fullPackagePath{
				packagePath: strings.TrimPrefix(srcDir, goSrc),
			}, nil
		}
	}
	return fullPackagePath{}, fmt.Errorf("module not found: '%s' is outside GOPATH", srcDir)
}
