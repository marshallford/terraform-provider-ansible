package navigator

import (
	"fmt"
	"maps"
	"slices"

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
	filePermissions           = 0o600

	containerRunDir = "/tmp/run"

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
	Options         ansible.PlaybookOptions
	Settings        Settings
}

// Zero until Preflight has run.
type preflightResults struct {
	navigatorBinary string
	workingDir      string
}

type Run struct {
	fs   afero.Fs
	exec ansible.Executor

	config           RunConfig
	env              map[string]string
	dirs             runDirs
	resolved         preflightResults
	artifactContents []byte

	Command ansible.Command
	Output  string
	Status  ansible.Status
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

func NewRun(hostDir string, config RunConfig, opts ...RunOption) *Run {
	run := &Run{
		fs:     afero.NewOsFs(),
		exec:   ansible.OSExecutor(),
		config: config,
		dirs:   newRunDirs(hostDir, config.mode()),
	}
	for _, opt := range opts {
		opt(run)
	}

	return run
}

func (r *Run) Mode() Mode {
	return r.config.mode()
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

// SetEnv must be called before Setup, which bakes the names into the generated
// settings.
func (r *Run) SetEnv(name string, value string) {
	if r.env == nil {
		r.env = map[string]string{}
	}

	r.env[name] = value
}

func (r *Run) Cleanup() error {
	if err := r.fs.RemoveAll(r.HostDir()); err != nil {
		return fmt.Errorf("failed to remove run directory, %w", err)
	}

	return nil
}

func (r *Run) settings() Settings {
	settings := r.config.Settings
	execEnv := &settings.ExecutionEnvironment

	execEnv.EnvironmentVariables.Pass = slices.Clone(execEnv.EnvironmentVariables.Pass)
	execEnv.EnvironmentVariables.Set = maps.Clone(execEnv.EnvironmentVariables.Set)

	for _, name := range slices.Sorted(maps.Keys(r.env)) {
		execEnv.EnvironmentVariables.pass(name)
	}

	if r.config.mode().UsesEE() {
		execEnv.VolumeMounts = append(slices.Clone(execEnv.VolumeMounts), VolumeMount{
			Src:     r.HostDir(),
			Dest:    r.PlaybookDir(),
			Options: VolumeMountOptions{VolumeMountRelabelUnshared},
		})
	}

	return settings
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
