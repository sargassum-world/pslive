package database

import (
	"io/fs"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
)

// Filesystem Utils

func listFiles(f fs.FS, filter func(path string) bool) ([]string, error) {
	files := []string{}
	err := fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filter == nil || filter(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func readFile(filename string, f fs.FS) ([]byte, error) {
	file, err := f.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return data, err
}

const queryFileExt = ".sql"

func filterQuery(path string) bool {
	return strings.HasSuffix(path, queryFileExt)
}

func readQueries(f fs.FS, filter func(path string) bool) ([]string, error) {
	files, err := listFiles(f, filter)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't list query files")
	}
	queries := make([]string, len(files))
	for i, file := range files {
		migration, err := readFile(file, f)
		if err != nil {
			return nil, errors.Wrapf(err, "couldn't open query file %s", file)
		}
		queries[i] = string(migration)
	}
	return queries, nil
}
