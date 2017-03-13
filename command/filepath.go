package command

import (
	"path/filepath"
)

type FilePath string

func (f FilePath) IsSameFilePath(other FilePath) (bool, error) {
	abs1, err := filepath.Abs(string(f))
	if err != nil {
		return false, err
	}

	abs2, err := filepath.Abs(string(other))
	if err != nil {
		return false, err
	}

	return abs1 == abs2, nil
}
