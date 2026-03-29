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

	command := r.generateCommand(ctx)
	r.Command = command.String()

	commandOutput, err := command.Run()
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

	contents, err := generateSettings(r.config.Settings)
	if err != nil {
		return fmt.Errorf("failed to generate navigator settings, %w", err)
	}

	return writeFile(r.fs, r.hostJoin(navigatorSettingsFilename), contents)
}

func (r *Run) generateCommand(ctx context.Context) ansible.Cmd { //nolint:ireturn
	command := r.exec.CommandContext(ctx, r.binary,
		"run",
		r.hostJoin(playbookFilename),
		"--playbook-artifact-save-as",
		r.hostJoin(playbookArtifactFilename),
		"--log-file",
		r.hostJoin(navigatorLogFilename),
	)
	command.SetDir(r.config.WorkingDir)

	command.SetEnv(r.exec.Environ())
	command.AppendEnv("ANSIBLE_NAVIGATOR_CONFIG", r.hostJoin(navigatorSettingsFilename))

	for name, value := range r.config.Env {
		command.AppendEnv(name, value)
	}

	command.AppendArgs(r.args()...)

	if r.config.HostKeyChecking != ansible.RunnerDefaultHostKeyChecking {
		command.AppendEnv("ANSIBLE_HOST_KEY_CHECKING", fmt.Sprintf("%t", r.config.HostKeyChecking))
	}

	return r.launcher.PrepareCommand(command, r.config)
}

func (r *Run) args() []string {
	var args []string

	for _, inventory := range r.config.Inventories {
		if inventory.Exclude {
			continue
		}
		args = append(args, "--inventory", r.hostJoin(inventoriesDir, inventory.Name))
	}

	for _, f := range r.config.ExtraVars {
		args = append(args, "--extra-vars", fmt.Sprintf("@%s", r.resolvedJoin(extraVarsDir, f.Name)))
	}

	args = append(args, r.config.Options.Args()...)

	for _, key := range r.config.PrivateKeys {
		args = append(args, "--private-key", r.resolvedJoin(privateKeysDir, key.Name))
	}

	if r.config.UseKnownHosts {
		args = append(args, "--extra-vars", fmt.Sprintf("%s=%s", ansible.SSHKnownHostsFileVar, r.resolvedJoin(knownHostsDir, knownHostsFile)))
	}

	return args
}

func (r *Run) readPlaybookArtifact() (*ansible.PlaybookArtifact, error) {
	contents, err := afero.ReadFile(r.fs, r.hostJoin(playbookArtifactFilename))
	if err != nil {
		return nil, fmt.Errorf("failed to read playbook artifact, %w", err)
	}

	return ansible.ParsePlaybookArtifact(contents)
}
