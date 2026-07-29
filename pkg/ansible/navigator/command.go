package navigator

import (
	"fmt"

	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

func (r *Run) navigatorCommand() ansible.Command {
	command := ansible.Command{
		Name: r.navigatorBinary,
		Args: []string{
			"run",
			r.navigatorJoin(playbookFilename),
			"--playbook-artifact-save-as",
			r.navigatorJoin(playbookArtifactFilename),
			"--log-file",
			r.navigatorJoin(navigatorLogFilename),
		},
		Dir: r.config.WorkingDir,
		Env: r.exec.Environ(),
	}

	command = command.AppendArgs(r.navigatorArgs()...)
	command = command.AppendEnv("ANSIBLE_NAVIGATOR_CONFIG", r.navigatorJoin(navigatorSettingsFilename))

	for name, value := range r.config.Env {
		command = command.AppendEnv(name, value)
	}

	if r.config.HostKeyChecking != ansible.RunnerDefaultHostKeyChecking {
		command = command.AppendEnv("ANSIBLE_HOST_KEY_CHECKING", fmt.Sprintf("%t", r.config.HostKeyChecking))
	}

	return command
}

func (r *Run) navigatorArgs() []string {
	var args []string

	for _, inventory := range r.config.Inventories {
		if inventory.Exclude {
			continue
		}

		args = append(args, "--inventory", r.navigatorJoin(inventoriesDir, inventory.Name))
	}

	for _, f := range r.config.ExtraVars {
		args = append(args, "--extra-vars", fmt.Sprintf("@%s", r.playbookJoin(extraVarsDir, f.Name)))
	}

	args = append(args, r.config.Options.Args()...)

	for _, key := range r.config.PrivateKeys {
		args = append(args, "--private-key", r.playbookJoin(privateKeysDir, key.Name))
	}

	if r.config.UseKnownHosts {
		args = append(args, "--extra-vars", fmt.Sprintf("%s=%s", ansible.SSHKnownHostsFileVar, r.playbookJoin(knownHostsDir, knownHostsFile)))
	}

	return args
}
