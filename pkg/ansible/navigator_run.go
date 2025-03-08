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
	ForceHandlers   bool
	SkipTags        []string
	StartAtTask     string
	Limit           []string
	Tags            []string
	PrivateKeys     []string
	KnownHosts      bool
	HostKeyChecking bool
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

	for _, inventory := range options.Inventories {
		command.Args = append(command.Args, "--inventory", InventoryPath(runDir, inventory, eeEnabled, false))
	}

	if options.ForceHandlers {
		command.Args = append(command.Args, "--force-handlers")
	}

	if len(options.SkipTags) > 0 {
		command.Args = append(command.Args, "--skip-tags", strings.Join(options.SkipTags, ","))
	}

	if options.StartAtTask != "" {
		command.Args = append(command.Args, "--start-at-task", options.StartAtTask)
	}

	if len(options.Limit) > 0 {
		command.Args = append(command.Args, "--limit", strings.Join(options.Limit, ","))
	}

	if len(options.Tags) > 0 {
		command.Args = append(command.Args, "--tags", strings.Join(options.Tags, ","))
	}

	for _, key := range options.PrivateKeys {
		command.Args = append(command.Args, "--private-key", PrivateKeyPath(runDir, key, eeEnabled))
	}

	if options.KnownHosts {
		command.Args = append(command.Args, "--extra-vars", fmt.Sprintf("%s=%s", SSHKnownHostsFileVar, KnownHostsPath(runDir, eeEnabled)))
	}

	if options.HostKeyChecking != RunnerDefaultHostKeyChecking { //nolint:gosimple
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
	if err := os.Mkdir(dir, 0o700); err != nil { //nolint:gomnd,mnd
		return fmt.Errorf("failed to create directory for run, %w", err)
	}

	if err := os.Mkdir(filepath.Join(dir, inventoriesDir), 0o700); err != nil { //nolint:gomnd,mnd
		return fmt.Errorf("failed to create inventories directory for run, %w", err)
	}

	if err := os.Mkdir(filepath.Join(dir, privateKeysDir), 0o700); err != nil { //nolint:gomnd,mnd
		return fmt.Errorf("failed to create private keys directory for run, %w", err)
	}

	if err := os.Mkdir(filepath.Join(dir, knownHostsDir), 0o700); err != nil { //nolint:gomnd,mnd
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
