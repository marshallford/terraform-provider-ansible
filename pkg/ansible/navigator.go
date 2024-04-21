package ansible

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const NavigatorProgram = "ansible-navigator"

var (
	ErrDirectory           = errors.New("directory is not valid")
	ErrContainerEnginePath = errors.New("container engine (podman or docker) must exist in PATH")
	ErrContainerEngine     = errors.New("container engine is not running or usable")
	ErrNavigatorAbsPath    = fmt.Errorf("absolute path of %s cannot be represented", NavigatorProgram)
	ErrNavigatorPath       = fmt.Errorf("%s does not exist in PATH", NavigatorProgram)
	ErrNavigator           = fmt.Errorf("%s is not functional", NavigatorProgram)
)

func ContainerEngineOptions() []string {
	return []string{"auto", "podman", "docker"}
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
	podman := programExistsOnPath("podman")
	docker := programExistsOnPath("docker")

	if containerEngine == "podman" && podman != nil {
		return podman
	}

	if containerEngine == "docker" && docker != nil {
		return docker
	}

	if containerEngine == "auto" {
		if podman != nil && docker != nil {
			return ErrContainerEnginePath
		}

		if podman == nil {
			containerEngine = "podman"
		} else {
			containerEngine = "docker"
		}
	}

	command := exec.Command(containerEngine, "info")
	if err := command.Run(); err != nil {
		return fmt.Errorf("%w, '%s info' command failed, %w", ErrContainerEngine, containerEngine, err)
	}

	return nil
}

func NavigatorPath(path string) (string, error) {
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
