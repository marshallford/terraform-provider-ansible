package ansible

import (
	"context"
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
	inventoryFilename        = "inventory"
	playbookArtifactFilename = "playbook-artifact.json"
	navigatorLogFilename     = "ansible-navigator.log"
	navigatorRunLogFilename  = "output.log"
)

type RunOptions struct {
	ForceHandlers bool
	Limit         []string
	Tags          []string
}

type ArtifactQuery struct {
	JSONPath   string
	JSONOutput bool
	Result     string
}

func GenerateNavigatorRunCommand(ctx context.Context, workingDirectory string, ansibleNavigatorBinary string, tempRunDir string, opts *RunOptions) *exec.Cmd {
	command := exec.CommandContext(ctx, ansibleNavigatorBinary, []string{ // #nosec G204
		"run",
		filepath.Join(tempRunDir, playbookFilename),
		"--inventory",
		filepath.Join(tempRunDir, inventoryFilename),
		"--playbook-artifact-save-as",
		filepath.Join(tempRunDir, playbookArtifactFilename),
		"--log-file",
		filepath.Join(tempRunDir, navigatorLogFilename),
	}...)
	command.Dir = workingDirectory

	command.Env = append(os.Environ(), fmt.Sprintf("ANSIBLE_NAVIGATOR_CONFIG=%s", filepath.Join(tempRunDir, navigatorSettingsFilename)))
	command.WaitDelay = commandWaitDelay

	if opts.ForceHandlers {
		command.Args = append(command.Args, "--force-handlers")
	}

	if len(opts.Limit) > 0 {
		command.Args = append(command.Args, "--limit", strings.Join(opts.Limit, ","))
	}

	if len(opts.Tags) > 0 {
		command.Args = append(command.Args, "--tags", strings.Join(opts.Tags, ","))
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

func CreateTempRunDir(baseRunDir string, pattern string) (string, error) {
	dir, err := os.MkdirTemp(baseRunDir, fmt.Sprintf("%s-", pattern))
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory for run, %w", err)
	}

	return dir, nil
}

func RemoveTempRunDir(tempRunDir string) error {
	err := os.RemoveAll(tempRunDir)
	if err != nil {
		return fmt.Errorf("failed to remove temporary directory for run, %w", err)
	}

	return nil
}

func CreatePlaybookFile(tempRunDir string, playbookContents string) error {
	path := filepath.Join(tempRunDir, playbookFilename)

	err := writeFile(path, playbookContents)
	if err != nil {
		return fmt.Errorf("failed to create ansible playbook file for run, %w", err)
	}

	return nil
}

func CreateInventoryFile(tempRunDir string, inventoryContents string) error {
	path := filepath.Join(tempRunDir, inventoryFilename)

	err := writeFile(path, inventoryContents)
	if err != nil {
		return fmt.Errorf("failed to create ansible inventory file, %w", err)
	}

	return nil
}

func CreateNavigatorRunLogFile(tempRunDir string, outputContents string) error {
	path := filepath.Join(tempRunDir, navigatorRunLogFilename)

	err := writeFile(path, outputContents)
	if err != nil {
		return fmt.Errorf("failed to create %s file for run, %w", NavigatorProgram, err)
	}

	return nil
}

func QueryPlaybookArtifact(tempRunDir string, queries map[string]ArtifactQuery) error {
	path := filepath.Join(tempRunDir, playbookArtifactFilename)

	contents, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read playbook artifact, %w", err)
	}

	for name, query := range queries {
		result, err := jsonPath(contents, query.JSONPath)
		if err != nil {
			return fmt.Errorf("failed to query playbook artifact with JSONPath, %w", err)
		}

		query.Result = result
		queries[name] = query
	}

	return nil
}
