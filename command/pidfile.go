package command

import (
	"os"
)

type PidFile struct {
	FilePath
}

func NewPidFile(path string) (*PidFile, error) {

	_, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return &PidFile{FilePath(path)}, nil
}

func DeletePidFile(path string) error {

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info == nil {
		return nil
	}

	err = os.Remove(path)
	if err != nil {
		return err
	}

	return nil
}
