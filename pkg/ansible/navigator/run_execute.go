package navigator

import (
	"context"
	"fmt"
	"slices"

	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
	"github.com/spf13/afero"
)

func (r *Run) Execute(ctx context.Context) error {
	if err := r.writeGeneratedSettings(); err != nil {
		return err
	}

	command := r.navigatorCommand()
	r.Command = command.String()

	commandOutput, err := r.exec.Run(ctx, command)
	if err != nil {
		if artifact, readErr := r.readPlaybookArtifact(); readErr == nil {
			r.Output = artifact.Stdout.String()
			r.Status = artifact.Status
		}

		if r.Output == "" {
			r.Output = string(commandOutput)
		}

		return fmt.Errorf("%s run command failed, %w", Program, err)
	}

	r.Output = string(commandOutput)
	r.Status = "successful"

	return nil
}

func (r *Run) writeGeneratedSettings() error {
	for name := range r.config.Env {
		if !slices.Contains(r.config.Settings.EnvironmentVariablesPass, name) {
			r.config.Settings.EnvironmentVariablesPass = append(r.config.Settings.EnvironmentVariablesPass, name)
		}

		delete(r.config.Settings.EnvironmentVariablesSet, name)
	}

	contents, err := r.config.Settings.generate()
	if err != nil {
		return fmt.Errorf("failed to generate navigator settings, %w", err)
	}

	return writeFile(r.fs, r.hostJoin(navigatorSettingsFilename), contents)
}

func (r *Run) readPlaybookArtifact() (*ansible.PlaybookArtifact, error) {
	contents, err := afero.ReadFile(r.fs, r.hostJoin(playbookArtifactFilename))
	if err != nil {
		return nil, fmt.Errorf("failed to read playbook artifact, %w", err)
	}

	return ansible.ParsePlaybookArtifact(contents)
}
