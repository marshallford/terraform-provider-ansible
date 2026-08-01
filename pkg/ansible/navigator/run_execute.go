package navigator

import (
	"context"
	"fmt"

	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
	"github.com/spf13/afero"
)

func (r *Run) Execute(ctx context.Context) error {
	r.Command = r.navigatorCommand()

	commandOutput, err := r.exec.Run(ctx, r.Command)
	if err != nil {
		if artifact, readErr := r.playbookArtifact(); readErr == nil {
			r.Output = artifact.Stdout.String()
			r.Status = artifact.Status
		}

		if r.Output == "" {
			r.Output = string(commandOutput)
		}

		return fmt.Errorf("%s run command failed, %w", Program, err)
	}

	r.Output = string(commandOutput)
	r.Status = ansible.StatusSuccessful

	return nil
}

func (r *Run) readPlaybookArtifact() ([]byte, error) {
	if r.artifactContents != nil {
		return r.artifactContents, nil
	}

	contents, err := afero.ReadFile(r.fs, r.hostJoin(playbookArtifactFilename))
	if err != nil {
		return nil, fmt.Errorf("failed to read playbook artifact, %w", err)
	}

	r.artifactContents = contents

	return contents, nil
}

func (r *Run) playbookArtifact() (*ansible.PlaybookArtifact, error) {
	contents, err := r.readPlaybookArtifact()
	if err != nil {
		return nil, err
	}

	return ansible.ParsePlaybookArtifact(contents)
}
