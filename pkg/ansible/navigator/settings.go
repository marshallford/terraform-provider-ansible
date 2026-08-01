package navigator

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type VolumeMountOption string

const (
	VolumeMountOverlay         VolumeMountOption = "O"
	VolumeMountReadOnly        VolumeMountOption = "ro"
	VolumeMountReadWrite       VolumeMountOption = "rw"
	VolumeMountRelabelShared   VolumeMountOption = "z"
	VolumeMountRelabelUnshared VolumeMountOption = "Z"
)

type VolumeMountOptions []VolumeMountOption

func (o VolumeMountOptions) String() string {
	options := make([]string, 0, len(o))
	for _, option := range o {
		options = append(options, string(option))
	}

	return strings.Join(options, ",")
}

func validVolumeMountOptions() VolumeMountOptions {
	return VolumeMountOptions{
		VolumeMountOverlay,
		VolumeMountReadOnly,
		VolumeMountReadWrite,
		VolumeMountRelabelShared,
		VolumeMountRelabelUnshared,
	}
}

type VolumeMount struct {
	Src     string
	Dest    string
	Options VolumeMountOptions
}

type Pull struct {
	Arguments []string
	Policy    string
}

type EnvironmentVariables struct {
	Pass []string
	Set  map[string]string
}

// The value comes from the navigator process, so a Set entry for the same name
// would compete with it.
func (e *EnvironmentVariables) pass(name string) {
	if !slices.Contains(e.Pass, name) {
		e.Pass = append(e.Pass, name)
	}

	delete(e.Set, name)
}

type ExecutionEnvironment struct {
	Enabled              bool
	ContainerEngine      string
	Image                string
	Pull                 Pull
	EnvironmentVariables EnvironmentVariables
	VolumeMounts         []VolumeMount
	ContainerOptions     []string
}

type Settings struct {
	Timeout              time.Duration
	Timezone             string
	ExecutionEnvironment ExecutionEnvironment
}

type settingsFormatAnsibleRunner struct {
	Timeout uint32 `yaml:"timeout"`
}

type settingsFormatColor struct {
	Enable bool `yaml:"enable"`
	OSC4   bool `yaml:"osc4"`
}

type settingsFormatLogging struct {
	Level string `yaml:"level"`
}

type settingsFormatPlaybookArtifact struct {
	Enable bool `yaml:"enable"`
}

type settingsFormatEnvironmentVariables struct {
	Pass []string          `yaml:"pass"` // #nosec G117
	Set  map[string]string `yaml:"set"`
}

type settingsFormatPull struct {
	Arguments []string `yaml:"arguments"`
	Policy    string   `yaml:"policy"`
}

type settingsFormatVolumeMounts struct {
	Src     string `yaml:"src"`
	Dest    string `yaml:"dest"`
	Options string `yaml:"options"`
}

type settingsFormatExecutionEnvironment struct {
	ContainerEngine      string                             `yaml:"container-engine"` //nolint:tagliatelle
	Enabled              bool                               `yaml:"enabled"`
	EnvironmentVariables settingsFormatEnvironmentVariables `yaml:"environment-variables"` //nolint:tagliatelle
	Image                string                             `yaml:"image"`
	Pull                 settingsFormatPull                 `yaml:"pull"`
	VolumeMounts         []settingsFormatVolumeMounts       `yaml:"volume-mounts"`     //nolint:tagliatelle
	ContainerOptions     []string                           `yaml:"container-options"` //nolint:tagliatelle
}

type settingsFormatAnsibleNavigator struct {
	AnsibleRunner        settingsFormatAnsibleRunner        `yaml:"ansible-runner"` //nolint:tagliatelle
	Color                settingsFormatColor                `yaml:"color"`
	ExecutionEnvironment settingsFormatExecutionEnvironment `yaml:"execution-environment"` //nolint:tagliatelle
	Logging              settingsFormatLogging              `yaml:"logging"`
	Mode                 string                             `yaml:"mode"`
	PlaybookArtifact     settingsFormatPlaybookArtifact     `yaml:"playbook-artifact"` //nolint:tagliatelle
	Timezone             string                             `yaml:"time-zone"`         //nolint:tagliatelle
}

type settingsFormat struct {
	AnsibleNavigator settingsFormatAnsibleNavigator `yaml:"ansible-navigator"` //nolint:tagliatelle
}

func (s Settings) generate() (string, error) {
	execEnv := s.ExecutionEnvironment

	volumeMounts := make([]settingsFormatVolumeMounts, 0, len(execEnv.VolumeMounts))
	for _, mount := range execEnv.VolumeMounts {
		volumeMounts = append(volumeMounts, settingsFormatVolumeMounts{
			Src:     mount.Src,
			Dest:    mount.Dest,
			Options: mount.Options.String(),
		})
	}

	format := settingsFormat{
		AnsibleNavigator: settingsFormatAnsibleNavigator{
			AnsibleRunner: settingsFormatAnsibleRunner{
				Timeout: uint32(s.Timeout.Seconds()),
			},
			Color: settingsFormatColor{
				Enable: false,
				OSC4:   false,
			},
			ExecutionEnvironment: settingsFormatExecutionEnvironment{
				ContainerEngine: execEnv.ContainerEngine,
				Enabled:         execEnv.Enabled,
				EnvironmentVariables: settingsFormatEnvironmentVariables{
					Pass: execEnv.EnvironmentVariables.Pass,
					Set:  execEnv.EnvironmentVariables.Set,
				},
				Image: execEnv.Image,
				Pull: settingsFormatPull{
					Arguments: execEnv.Pull.Arguments,
					Policy:    execEnv.Pull.Policy,
				},
				VolumeMounts:     volumeMounts,
				ContainerOptions: execEnv.ContainerOptions,
			},
			Logging: settingsFormatLogging{
				Level: "debug",
			},
			Mode: "stdout",
			PlaybookArtifact: settingsFormatPlaybookArtifact{
				Enable: true,
			},
			Timezone: s.Timezone,
		},
	}

	data, err := yaml.Marshal(&format)
	if err != nil {
		return "", fmt.Errorf("failed to build ansible-navigator settings file, %w", err)
	}

	return string(data), nil
}

func ContainerEngines() []string {
	return []string{"podman", "docker"}
}

func ContainerEngineOptions() []string {
	return append(ContainerEngines(), ContainerEngineAuto)
}

func PullPolicyOptions() []string {
	return []string{"always", "missing", "never", "tag"}
}
