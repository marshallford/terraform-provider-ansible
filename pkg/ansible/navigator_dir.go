package ansible

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	containerRunDir        = "/tmp/run" // TODO assumes container is unix-like with a /tmp dir.
	containerPathSeparator = "/"        // TODO assumes container is unix-like.
)

type RunDir struct {
	Host                  string
	Resolved              string
	resolvedPathSeparator string
}

func (d RunDir) HostJoin(paths ...string) string {
	paths = append([]string{d.Host}, paths...)

	return filepath.Join(paths...)
}

func (d RunDir) ResolvedJoin(paths ...string) string {
	paths = append([]string{d.Resolved}, paths...)

	return filepath.Clean(strings.Join(paths, d.resolvedPathSeparator))
}

func (d RunDir) Remove() error {
	if err := os.RemoveAll(d.Host); err != nil {
		return fmt.Errorf("failed to remove run directory, %w", err)
	}

	return nil
}

func CreateRunDir(hostDir string, settings *NavigatorSettings) (*RunDir, error) {
	runDir := RunDir{
		Host:                  filepath.Clean(hostDir),
		Resolved:              filepath.Clean(hostDir),
		resolvedPathSeparator: string(os.PathSeparator),
	}

	if err := os.Mkdir(runDir.Host, 0o700); err != nil { //nolint:mnd
		return nil, fmt.Errorf("failed to create directory for run, %w", err)
	}

	if err := os.Mkdir(runDir.HostJoin(inventoriesDir), 0o700); err != nil { //nolint:mnd
		return nil, fmt.Errorf("failed to create inventories directory for run, %w", err)
	}

	if err := os.Mkdir(runDir.HostJoin(extraVarsDir), 0o700); err != nil { //nolint:mnd
		return nil, fmt.Errorf("failed to create extra vars directory for run, %w", err)
	}

	if err := os.Mkdir(runDir.HostJoin(privateKeysDir), 0o700); err != nil { //nolint:mnd
		return nil, fmt.Errorf("failed to create private keys directory for run, %w", err)
	}

	if err := os.Mkdir(runDir.HostJoin(knownHostsDir), 0o700); err != nil { //nolint:mnd
		return nil, fmt.Errorf("failed to create known hosts directory for run, %w", err)
	}

	if !settings.EEEnabled {
		return &runDir, nil
	}

	runDir.Resolved = containerRunDir
	runDir.resolvedPathSeparator = containerPathSeparator

	if settings.VolumeMounts == nil {
		settings.VolumeMounts = map[string]string{}
	}

	settings.VolumeMounts[runDir.Host] = runDir.Resolved

	return &runDir, nil
}
