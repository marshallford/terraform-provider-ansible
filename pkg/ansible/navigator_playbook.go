package ansible

import (
	"fmt"
)

const (
	playbookFilename = "playbook.yaml"
)

func CreatePlaybook(runDir *RunDir, playbookContents string) error {
	path := runDir.HostJoin(playbookFilename)

	err := writeFile(path, playbookContents)
	if err != nil {
		return fmt.Errorf("failed to create ansible playbook file for run, %w", err)
	}

	return nil
}
