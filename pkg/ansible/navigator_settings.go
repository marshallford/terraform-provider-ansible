package ansible

import (
	"fmt"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const navigatorSettingsFilename = "ansible-navigator.yaml"

type NavigatorSettings struct {
	Timeout                  time.Duration
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

type navigatorSettingsFormatAnsibleRunner struct {
	Timeout uint32 `yaml:"timeout"`
}

type navigatorSettingsFormatColor struct {
	Enable bool `yaml:"enable"`
	OSC4   bool `yaml:"osc4"`
}

type navigatorSettingsFormatLogging struct {
	Level string `yaml:"level"`
}

type navigatorSettingsFormatPlaybookArtifact struct {
	Enable bool `yaml:"enable"`
}

type navigatorSettingsFormatEnvironmentVariables struct {
	Pass []string          `yaml:"pass"`
	Set  map[string]string `yaml:"set"`
}

type navigatorSettingsFormatPull struct {
	Arguments []string `yaml:"arguments"`
	Policy    string   `yaml:"policy"`
}

type navigatorSettingsFormatVolumeMounts struct {
	Src     string `yaml:"src"`
	Dest    string `yaml:"dest"`
	Options string `yaml:"options"`
}

type navigatorSettingsFormatExecutionEnvironment struct {
	ContainerEngine      string                                      `yaml:"container-engine"` //nolint:tagliatelle
	Enabled              bool                                        `yaml:"enabled"`
	EnvironmentVariables navigatorSettingsFormatEnvironmentVariables `yaml:"environment-variables"` //nolint:tagliatelle
	Image                string                                      `yaml:"image"`
	Pull                 navigatorSettingsFormatPull                 `yaml:"pull"`
	VolumeMounts         []navigatorSettingsFormatVolumeMounts       `yaml:"volume-mounts"`     //nolint:tagliatelle
	ContainerOptions     []string                                    `yaml:"container-options"` //nolint:tagliatelle
}

type navigatorSettingsFormatAnsibleNavigator struct {
	AnsibleRunner        navigatorSettingsFormatAnsibleRunner        `yaml:"ansible-runner"` //nolint:tagliatelle
	Color                navigatorSettingsFormatColor                `yaml:"color"`
	ExecutionEnvironment navigatorSettingsFormatExecutionEnvironment `yaml:"execution-environment"` //nolint:tagliatelle
	Logging              navigatorSettingsFormatLogging              `yaml:"logging"`
	Mode                 string                                      `yaml:"mode"`
	PlaybookArtifact     navigatorSettingsFormatPlaybookArtifact     `yaml:"playbook-artifact"` //nolint:tagliatelle
	Timezone             string                                      `yaml:"time-zone"`         //nolint:tagliatelle
}

type navigatorSettingsFormat struct {
	AnsibleNavigator navigatorSettingsFormatAnsibleNavigator `yaml:"ansible-navigator"` //nolint:tagliatelle
}

func GenerateNavigatorSettings(settings *NavigatorSettings) (string, error) {
	volumeMounts := make([]navigatorSettingsFormatVolumeMounts, 0, len(settings.VolumeMounts))
	for src, dest := range settings.VolumeMounts {
		volumeMounts = append(volumeMounts, navigatorSettingsFormatVolumeMounts{Src: src, Dest: dest, Options: "Z"})
	}

	settingsFormat := navigatorSettingsFormat{
		AnsibleNavigator: navigatorSettingsFormatAnsibleNavigator{
			AnsibleRunner: navigatorSettingsFormatAnsibleRunner{
				Timeout: uint32(settings.Timeout.Seconds()),
			},
			Color: navigatorSettingsFormatColor{
				Enable: false,
				OSC4:   false,
			},
			ExecutionEnvironment: navigatorSettingsFormatExecutionEnvironment{
				ContainerEngine: settings.ContainerEngine,
				Enabled:         true,
				EnvironmentVariables: navigatorSettingsFormatEnvironmentVariables{
					Pass: settings.EnvironmentVariablesPass,
					Set:  settings.EnvironmentVariablesSet,
				},
				Image: settings.Image,
				Pull: navigatorSettingsFormatPull{
					Arguments: settings.PullArguments,
					Policy:    settings.PullPolicy,
				},
				VolumeMounts:     volumeMounts,
				ContainerOptions: settings.ContainerOptions,
			},
			Logging: navigatorSettingsFormatLogging{
				Level: "debug",
			},
			Mode: "stdout",
			PlaybookArtifact: navigatorSettingsFormatPlaybookArtifact{
				Enable: true,
			},
			Timezone: settings.Timezone,
		},
	}

	data, err := yaml.Marshal(&settingsFormat)
	if err != nil {
		return "", fmt.Errorf("failed to build ansible-navigator settings file, %w", err)
	}

	return string(data), nil
}

func CreateNavigatorSettingsFile(dir string, settingsContents string) error {
	path := filepath.Join(dir, navigatorSettingsFilename)

	err := writeFile(path, settingsContents)
	if err != nil {
		return fmt.Errorf("failed to create %s settings file, %w", NavigatorProgram, err)
	}

	return nil
}
