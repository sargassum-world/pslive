package opa

import (
	"embed"
	"io/fs"
	"strings"

	"github.com/open-policy-agent/opa/rego"
	"github.com/pkg/errors"
)

const moduleFileExt = ".rego"

func filterModule(path string) bool {
	return strings.HasSuffix(path, moduleFileExt)
}

type Module struct {
	Filename string
	Contents string
}

func Modules(modules ...[]Module) func(r *rego.Rego) {
	return func(r *rego.Rego) {
		for _, m := range modules {
			for _, module := range m {
				rego.Module(module.Filename, module.Contents)(r)
			}
		}
	}
}

func readFiles(f fs.FS, filter func(path string) bool) ([]Module, error) {
	modules := []Module{}
	err := fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filter == nil || filter(path) {
			b, err := fs.ReadFile(f, path)
			if err != nil {
				return errors.Wrapf(nil, "couldn't read file %s", path)
			}
			modules = append(modules, Module{
				Filename: path,
				Contents: string(b),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return modules, nil
}

func FSModules(fsys fs.FS, filePrefix string) ([]Module, error) {
	modules, err := readFiles(fsys, filterModule)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't list rego modules")
	}
	qualifiedModules := make([]Module, len(modules))
	for i, module := range modules {
		qualifiedModules[i] = Module{
			Filename: filePrefix + module.Filename,
			Contents: module.Contents,
		}
	}
	return qualifiedModules, nil
}

var (
	//go:embed routing/*
	routingFS         embed.FS
	routingModules, _ = FSModules(routingFS, "github.com/sargassum-world/pslive/pkg/godest/opa/")
	//go:embed errors/*
	errorsFS         embed.FS
	errorsModules, _ = FSModules(errorsFS, "github.com/sargassum-world/pslive/pkg/godest/opa/")
)

func RegoModules() []Module {
	return append(
		routingModules,
		errorsModules...,
	)
}
