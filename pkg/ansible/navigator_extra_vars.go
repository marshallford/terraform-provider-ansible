package ansible

import (
	"fmt"
)

const (
	extraVarsDir = "extra-vars"
)

type ExtraVarsFile struct {
	Name     string
	Contents string
}

func CreateExtraVarsFiles(runDir *RunDir, extraVarsFiles []ExtraVarsFile) error {
	for _, extraVarsFile := range extraVarsFiles {
		err := writeFile(runDir.HostJoin(extraVarsDir, extraVarsFile.Name), extraVarsFile.Contents)
		if err != nil {
			return fmt.Errorf("failed to create ansible extra-vars file for run, %w", err)
		}
	}

	return nil
}
