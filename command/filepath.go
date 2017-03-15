package command

import (
	"path/filepath"
)

type FilePath string

func (f FilePath) IsSameFilePath(other FilePath) bool {
	abs1, err := filepath.Abs(string(f))
	if err != nil {
		return false
	}

	abs2, err := filepath.Abs(string(other))
	if err != nil {
		return false
	}

	return abs1 == abs2
}
