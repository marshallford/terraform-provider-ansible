package ansible

import (
	"errors"
	"fmt"

	"github.com/spf13/afero"
)

var (
	ErrDirectory = errors.New("directory is not valid")
)

func CheckDirectory(fs afero.Fs, path string) error {
	info, err := fs.Stat(path)
	if err != nil {
		return fmt.Errorf("%w, %w", ErrDirectory, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%w, %s is not a directory", ErrDirectory, path)
	}

	return nil
}
