package ansible

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

const (
	NavigatorProgram    = "ansible-navigator"
	PlaybookProgram     = "ansible-playbook"
	ContainerEngineAuto = "auto"
)

var (
	ErrDirectory               = errors.New("directory is not valid")
	ErrContainerEngineValidate = errors.New("container engine is not valid")
	ErrContainerEnginePath     = errors.New("container engine must exist in PATH")
	ErrContainerEngineRunning  = errors.New("container engine is not running or usable")
	ErrNavigatorAbsPath        = fmt.Errorf("absolute path of %s cannot be represented", NavigatorProgram)
	ErrNavigatorPath           = fmt.Errorf("%s does not exist in PATH", NavigatorProgram)
	ErrNavigator               = fmt.Errorf("%s is not functional", NavigatorProgram)
	ErrPlaybookPath            = fmt.Errorf("%s does not exist in PATH", PlaybookProgram)
	ErrPlaybook                = fmt.Errorf("%s is not functional", PlaybookProgram)
)

func ContainerEngineOptions(auto bool) []string {
	containerEngines := []string{"podman", "docker"}

	if auto {
		containerEngines = append(containerEngines, ContainerEngineAuto)
	}

	return containerEngines
}

func PullPolicyOptions() []string {
	return []string{"always", "missing", "never", "tag"}
}

func DirectoryPreflight(dir string) error {
	stat, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("%w, %w", ErrDirectory, err)
	}

	if !stat.IsDir() {
		return fmt.Errorf("%w, %s is not a directory", ErrDirectory, dir)
	}

	return nil
}

func ContainerEnginePreflight(containerEngine string) error {
	if !slices.Contains(ContainerEngineOptions(true), containerEngine) {
		return fmt.Errorf("%w, %s is not an option", ErrContainerEngineValidate, containerEngine)
	}

	if containerEngine != ContainerEngineAuto && programExistsOnPath(containerEngine) != nil {
		return fmt.Errorf("%w, %s does not", ErrContainerEnginePath, containerEngine)
	}

	if containerEngine == ContainerEngineAuto {
		for _, option := range ContainerEngineOptions(false) {
			if programExistsOnPath(option) == nil {
				containerEngine = option

				break
			}
		}
	}

	if containerEngine == ContainerEngineAuto {
		return ErrContainerEnginePath
	}

	command := exec.Command(containerEngine, "info")
	if err := command.Run(); err != nil {
		return fmt.Errorf("%w, '%s info' command failed, %w", ErrContainerEngineRunning, containerEngine, err)
	}

	return nil
}

func NavigatorPathPreflight(path string) (string, error) {
	if path == "" {
		path, err := exec.LookPath(NavigatorProgram)
		if err != nil {
			return "", ErrNavigatorPath
		}

		return path, nil
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("%w, %w", ErrNavigatorAbsPath, err)
	}

	return path, nil
}

// TODO include output in error
// TODO require a min version
func NavigatorPreflight(binary string) error {
	command := exec.Command(binary, "--version")
	stdoutStderr, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w, '%s --version' command failed, %w", ErrNavigator, binary, err)
	}

	if !strings.HasPrefix(string(stdoutStderr), NavigatorProgram) {
		return fmt.Errorf("%w, '%s --version' command output not expected", ErrNavigator, binary)
	}

	return nil
}

// TODO include output in error
// TODO require a min version
func PlaybookPreflight() error {
	if err := programExistsOnPath(PlaybookProgram); err != nil {
		return fmt.Errorf("%w, ansible is required when running without an execution environment", ErrPlaybookPath)
	}

	command := exec.Command(PlaybookProgram, "--version")
	stdoutStderr, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w, '%s --version' command failed, %w", ErrPlaybook, PlaybookProgram, err)
	}

	if !strings.HasPrefix(string(stdoutStderr), PlaybookProgram) {
		return fmt.Errorf("%w, '%s --version' command output not expected", ErrPlaybook, PlaybookProgram)
	}

	return nil
}
