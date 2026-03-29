package navigator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
	"github.com/spf13/afero"
)

const (
	Program             = "ansible-navigator"
	ContainerEngineAuto = "auto"

	playbookArtifactFilename  = "playbook-artifact.json"
	navigatorLogFilename      = "ansible-navigator.log"
	navigatorSettingsFilename = "ansible-navigator.yaml"
	dirPermissions            = 0o700

	containerRunDir        = "/tmp/run" // TODO assumes container is unix-like with a /tmp dir.
	containerPathSeparator = "/"        // TODO assumes container is unix-like.

	inventoriesDir   = "inventories"
	extraVarsDir     = "extra-vars"
	privateKeysDir   = "private-keys"
	knownHostsDir    = "known-hosts"
	knownHostsFile   = "known_hosts"
	playbookFilename = "playbook.yaml"
)

type RunConfig struct {
	WorkingDir      string
	Binary          string
	Playbook        string
	Inventories     []ansible.Inventory
	ExtraVars       []ansible.ExtraVarsFile
	PrivateKeys     []ansible.PrivateKey
	KnownHosts      []ansible.KnownHost
	UseKnownHosts   bool
	HostKeyChecking bool
	Options         *ansible.PlaybookOptions
	Settings        *Settings
	Env             map[string]string
}

type Run struct {
	fs       afero.Fs
	exec     ansible.Executor
	launcher Launcher

	config                *RunConfig
	hostDir               string
	resolvedDir           string
	resolvedPathSeparator string
	binary                string

	Command string
	Output  string
	Status  string
}

type RunOption func(*Run)

func WithFs(fs afero.Fs) RunOption {
	return func(r *Run) {
		r.fs = fs
	}
}

func WithExecutor(exec ansible.Executor) RunOption {
	return func(r *Run) {
		r.exec = exec
	}
}

func WithLauncher(launcher Launcher) RunOption {
	return func(r *Run) {
		r.launcher = launcher
	}
}

func NewRun(hostDir string, config *RunConfig, opts ...RunOption) *Run {
	run := &Run{
		fs:                    afero.NewOsFs(),
		exec:                  ansible.OSExecutor(),
		launcher:              NativeLauncher(),
		config:                config,
		hostDir:               filepath.Clean(hostDir),
		resolvedDir:           filepath.Clean(hostDir),
		resolvedPathSeparator: string(os.PathSeparator),
	}
	for _, opt := range opts {
		opt(run)
	}

	return run
}

func (r *Run) HostDir() string {
	return r.hostDir
}

func (r *Run) ResolvedDir() string {
	return r.resolvedDir
}

func (r *Run) InventoryPath(name string) string {
	return r.resolvedJoin(inventoriesDir, name)
}

func (r *Run) Cleanup() error {
	if err := r.fs.RemoveAll(r.hostDir); err != nil {
		return fmt.Errorf("failed to remove run directory, %w", err)
	}

	return nil
}

func (r *Run) hostJoin(paths ...string) string {
	paths = append([]string{r.hostDir}, paths...)

	return filepath.Join(paths...)
}

func (r *Run) resolvedJoin(paths ...string) string {
	paths = append([]string{r.resolvedDir}, paths...)

	return filepath.Clean(strings.Join(paths, r.resolvedPathSeparator))
}
