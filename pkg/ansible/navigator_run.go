package ansible

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	commandWaitDelay         = 10 * time.Second
	playbookFilename         = "playbook.yaml"
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
	PrivateKeys     []string
	KnownHosts      bool
	HostKeyChecking bool
}

func navigatorRunCommandArgs(runDir string, eeEnabled bool, options *Options) []string {
	var args []string //nolint:prealloc

	for _, inventory := range options.Inventories {
		args = append(args, "--inventory", InventoryPath(runDir, inventory, eeEnabled, false))
	}

	for _, extraVarsFile := range options.ExtraVarsFiles {
		args = append(args, "--extra-vars", fmt.Sprintf("@%s", ExtraVarsPath(runDir, extraVarsFile, eeEnabled)))
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
		args = append(args, "--private-key", PrivateKeyPath(runDir, key, eeEnabled))
	}

	if options.KnownHosts {
		args = append(args, "--extra-vars", fmt.Sprintf("%s=%s", SSHKnownHostsFileVar, KnownHostsPath(runDir, eeEnabled)))
	}

	return args
}

func GenerateNavigatorRunCommand(runDir string, workingDir string, ansibleNavigatorBinary string, eeEnabled bool, options *Options) *exec.Cmd {
	command := exec.Command(ansibleNavigatorBinary, []string{ // #nosec G204
		"run",
		filepath.Join(runDir, playbookFilename),
		"--playbook-artifact-save-as",
		filepath.Join(runDir, playbookArtifactFilename),
		"--log-file",
		filepath.Join(runDir, navigatorLogFilename),
	}...)
	command.Dir = workingDir

	// TODO allow setting env vars directly for when EE is disabled
	command.Env = append(
		os.Environ(),
		fmt.Sprintf("ANSIBLE_NAVIGATOR_CONFIG=%s", filepath.Join(runDir, navigatorSettingsFilename)),
	)
	command.WaitDelay = commandWaitDelay

	command.Args = append(command.Args, navigatorRunCommandArgs(runDir, eeEnabled, options)...)

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

func CreateRunDir(dir string) error {
	if err := os.Mkdir(dir, 0o700); err != nil { //nolint:mnd
		return fmt.Errorf("failed to create directory for run, %w", err)
	}

	if err := os.Mkdir(filepath.Join(dir, inventoriesDir), 0o700); err != nil { //nolint:mnd
		return fmt.Errorf("failed to create inventories directory for run, %w", err)
	}

	if err := os.Mkdir(filepath.Join(dir, privateKeysDir), 0o700); err != nil { //nolint:mnd
		return fmt.Errorf("failed to create private keys directory for run, %w", err)
	}

	if err := os.Mkdir(filepath.Join(dir, knownHostsDir), 0o700); err != nil { //nolint:mnd
		return fmt.Errorf("failed to create known hosts directory for run, %w", err)
	}

	return nil
}

func RemoveRunDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove directory for run, %w", err)
	}

	return nil
}

func CreatePlaybook(dir string, playbookContents string) error {
	path := filepath.Join(dir, playbookFilename)

	err := writeFile(path, playbookContents)
	if err != nil {
		return fmt.Errorf("failed to create ansible playbook file for run, %w", err)
	}

	return nil
}
