package navigator

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

type Settings struct {
	Timeout                  time.Duration
	EEEnabled                bool
	ContainerEngine          string
	EnvironmentVariablesPass []string
	EnvironmentVariablesSet  map[string]string
	Image                    string
	PullArguments            []string
	PullPolicy               string
	VolumeMounts             map[string]string
	ContainerOptions         []string
	Timezone                 string
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

func generateSettings(settings *Settings) (string, error) {
	volumeMounts := make([]settingsFormatVolumeMounts, 0, len(settings.VolumeMounts))
	for src, dest := range settings.VolumeMounts {
		volumeMounts = append(volumeMounts, settingsFormatVolumeMounts{Src: src, Dest: dest, Options: "Z"})
	}

	format := settingsFormat{
		AnsibleNavigator: settingsFormatAnsibleNavigator{
			AnsibleRunner: settingsFormatAnsibleRunner{
				Timeout: uint32(settings.Timeout.Seconds()),
			},
			Color: settingsFormatColor{
				Enable: false,
				OSC4:   false,
			},
			ExecutionEnvironment: settingsFormatExecutionEnvironment{
				ContainerEngine: settings.ContainerEngine,
				Enabled:         settings.EEEnabled,
				EnvironmentVariables: settingsFormatEnvironmentVariables{
					Pass: settings.EnvironmentVariablesPass,
					Set:  settings.EnvironmentVariablesSet,
				},
				Image: settings.Image,
				Pull: settingsFormatPull{
					Arguments: settings.PullArguments,
					Policy:    settings.PullPolicy,
				},
				VolumeMounts:     volumeMounts,
				ContainerOptions: settings.ContainerOptions,
			},
			Logging: settingsFormatLogging{
				Level: "debug",
			},
			Mode: "stdout",
			PlaybookArtifact: settingsFormatPlaybookArtifact{
				Enable: true,
			},
			Timezone: settings.Timezone,
		},
	}

	data, err := yaml.Marshal(&format)
	if err != nil {
		return "", fmt.Errorf("failed to build ansible-navigator settings file, %w", err)
	}

	return string(data), nil
}

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
