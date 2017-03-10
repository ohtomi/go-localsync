package command

import (
	"os"
	"path/filepath"
)

type FileInfo struct {
	path string
	info os.FileInfo
}

type FileInfoStore struct {
	root    string
	storage []*FileInfo
}

func (f *FileInfoStore) Load(recursive bool) error {
	return filepath.Walk(f.root, f.WalkFunc(recursive))
}

func (f *FileInfoStore) WalkFunc(recursive bool) filepath.WalkFunc {

	return func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if !recursive && info.IsDir() && info.Name() != f.root {
			return filepath.SkipDir
		}

		rel, err := filepath.Rel(f.root, path)
		if err != nil {
			return err
		}
		f.storage = append(f.storage, &FileInfo{rel, info})

		return nil
	}
}
