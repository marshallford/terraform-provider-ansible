package navigator

import (
	"fmt"

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

	containerRunDir = "/tmp/run"

	selinuxRelabelOption = "Z"

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
	fs   afero.Fs
	exec ansible.Executor

	config          *RunConfig
	mode            Mode
	dirs            runDirs
	navigatorBinary string

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

func NewRun(hostDir string, config *RunConfig, opts ...RunOption) *Run {
	run := &Run{
		fs:     afero.NewOsFs(),
		exec:   ansible.OSExecutor(),
		config: config,
		mode:   config.mode(),
		dirs:   newRunDirs(config.mode(), hostDir),
	}
	for _, opt := range opts {
		opt(run)
	}

	return run
}

func (r *Run) Mode() Mode {
	return r.mode
}

func (r *Run) HostDir() string {
	return r.dirs.host.root
}

func (r *Run) NavigatorDir() string {
	return r.dirs.navigator.root
}

func (r *Run) PlaybookDir() string {
	return r.dirs.playbook.root
}

func (r *Run) InventoryPath(name string) string {
	return r.playbookJoin(inventoriesDir, name)
}

func (r *Run) Cleanup() error {
	if err := r.fs.RemoveAll(r.HostDir()); err != nil {
		return fmt.Errorf("failed to remove run directory, %w", err)
	}

	return nil
}

func (r *Run) hostJoin(parts ...string) string {
	return r.dirs.host.join(parts...)
}

func (r *Run) navigatorJoin(parts ...string) string {
	return r.dirs.navigator.join(parts...)
}

func (r *Run) playbookJoin(parts ...string) string {
	return r.dirs.playbook.join(parts...)
}
