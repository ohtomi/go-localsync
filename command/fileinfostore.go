package command

import (
	"fmt"
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

func (f *FileInfoStore) Add(rel string) error {
	file, err := os.Open(rel)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	f.storage = append(f.storage, &FileInfo{rel, info})
	fmt.Println("add: " + rel)
	return nil
}

func (f *FileInfoStore) Remove(rel string) error {

	for i, item := range f.storage {
		if item.path == rel {
			switch i {
			case 0:
				f.storage = f.storage[i:]
			case len(f.storage):
				f.storage = f.storage[:i]
			default:
				f.storage = append(f.storage[:i], f.storage[i:]...)
			}
			fmt.Println("remove: " + rel)
			return nil
		}
	}

	return nil // TODO return Error
}
