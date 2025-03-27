package ansible

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	extraVarsDir = "extra-vars"
	// TODO assumes EE is unix-like with a /tmp dir.
	eeExtraVarsDir = "/tmp/extra-vars"
)

type ExtraVarsFile struct {
	Name     string
	Contents string
}

func ExtraVarsPath(dir string, name string, eeEnabled bool) string {
	if eeEnabled {
		return strings.Join([]string{eeExtraVarsDir, name}, "/") // assume EE is unix-like
	}

	return filepath.Join(dir, extraVarsDir, name)
}

func CreateExtraVarsFiles(dir string, extraVarsFiles []ExtraVarsFile, settings *NavigatorSettings) error {
	for _, extraVarsFile := range extraVarsFiles {
		err := writeFile(filepath.Join(dir, extraVarsDir, extraVarsFile.Name), extraVarsFile.Contents)
		if err != nil {
			return fmt.Errorf("failed to create ansible extra-vars file for run, %w", err)
		}
	}

	if !settings.EEEnabled {
		return nil
	}

	// TODO better option?
	if settings.VolumeMounts == nil {
		settings.VolumeMounts = map[string]string{}
	}

	settings.VolumeMounts[filepath.Join(dir, extraVarsDir)] = eeExtraVarsDir

	return nil
}
