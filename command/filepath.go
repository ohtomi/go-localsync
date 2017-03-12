package command

import (
	"os"
)

type FilePath string

func (f FilePath) IsSameFilePath(other FilePath) (bool, error) {
	info1, err := os.Stat(string(f))
	if err != nil {
		return false, err
	}

	info2, err := os.Stat(string(other))
	if err != nil {
		return false, err
	}

	return info1.Name() == info2.Name(), nil
}
