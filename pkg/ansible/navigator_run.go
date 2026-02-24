package ansible

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	commandWaitDelay         = 10 * time.Second
	playbookArtifactFilename = "playbook-artifact.json"
	navigatorLogFilename     = "ansible-navigator.log"
	SSHKnownHostsFileVar     = "ansible_ssh_known_hosts_file"
)

type Options struct {
	Inventories     []string
	ExtraVarsFiles  []string
	ForceHandlers   bool
	SkipTags        []string
	StartAtTask     string
	Limit           []string
	Tags            []string
	PrivateKeys     []string // #nosec G117
	KnownHosts      bool
	HostKeyChecking bool
}

func navigatorRunCommandArgs(runDir *RunDir, options *Options) []string {
	var args []string

	for _, inventory := range options.Inventories {
		args = append(args, "--inventory", runDir.HostJoin(inventoriesDir, inventory))
	}

	for _, extraVarsFile := range options.ExtraVarsFiles {
		args = append(args, "--extra-vars", fmt.Sprintf("@%s", runDir.ResolvedJoin(extraVarsDir, extraVarsFile)))
	}

	if options.ForceHandlers {
		args = append(args, "--force-handlers")
	}

	if len(options.SkipTags) > 0 {
		args = append(args, "--skip-tags", strings.Join(options.SkipTags, ","))
	}

	if options.StartAtTask != "" {
		args = append(args, "--start-at-task", options.StartAtTask)
	}

	if len(options.Limit) > 0 {
		args = append(args, "--limit", strings.Join(options.Limit, ","))
	}

	if len(options.Tags) > 0 {
		args = append(args, "--tags", strings.Join(options.Tags, ","))
	}

	for _, key := range options.PrivateKeys {
		args = append(args, "--private-key", runDir.ResolvedJoin(privateKeysDir, key))
	}

	if options.KnownHosts {
		args = append(args, "--extra-vars", fmt.Sprintf("%s=%s", SSHKnownHostsFileVar, runDir.ResolvedJoin(knownHostsDir, knownHostsFile)))
	}

	return args
}

func GenerateNavigatorRunCommand(ctx context.Context, runDir *RunDir, workingDir string, ansibleNavigatorBinary string, options *Options) *exec.Cmd {
	command := exec.CommandContext(ctx, ansibleNavigatorBinary, []string{ // #nosec G204
		"run",
		runDir.HostJoin(playbookFilename),
		"--playbook-artifact-save-as",
		runDir.HostJoin(playbookArtifactFilename),
		"--log-file",
		runDir.HostJoin(navigatorLogFilename),
	}...)
	command.Dir = workingDir

	// TODO allow setting env vars directly for when EE is disabled
	command.Env = append(
		os.Environ(),
		fmt.Sprintf("ANSIBLE_NAVIGATOR_CONFIG=%s", runDir.HostJoin(navigatorSettingsFilename)),
	)
	command.WaitDelay = commandWaitDelay

	command.Args = append(command.Args, navigatorRunCommandArgs(runDir, options)...)

	if options.HostKeyChecking != RunnerDefaultHostKeyChecking { //nolint:staticcheck
		command.Env = append(command.Env, fmt.Sprintf("ANSIBLE_HOST_KEY_CHECKING=%t", options.HostKeyChecking))
	}

	return command
}

func ExecNavigatorRunCommand(command *exec.Cmd) (string, error) {
	stdoutStderr, err := command.CombinedOutput()
	if err != nil {
		return string(stdoutStderr), fmt.Errorf("%s run command failed, %w", NavigatorProgram, err)
	}

	return string(stdoutStderr), nil
}
