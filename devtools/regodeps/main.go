package main

import (
	"os"
	"path/filepath"

	"github.com/sargassum-world/pslive/internal/app/pslive"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	dir := filepath.Join(cwd, ".regodeps")
	const permissions = 0o775
	for _, module := range pslive.RegoDeps() {
		_, err := module.WriteFile(dir, permissions)
		if err != nil {
			panic(err)
		}
	}
}
