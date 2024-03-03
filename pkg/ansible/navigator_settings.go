package ansible

import (
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const navigatorSettingsFilename = "ansible-navigator.yaml"

type NavigatorSettings struct {
	ContainerEngine          string
	EnvironmentVariablesPass []string
	EnvironmentVariablesSet  map[string]string
	Image                    string
	PullArguments            []string
	PullPolicy               string
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

type navigatorSettingsFormatExecutionEnvironment struct {
	ContainerEngine      string                                      `yaml:"container-engine"`      //nolint:tagliatelle
	EnvironmentVariables navigatorSettingsFormatEnvironmentVariables `yaml:"environment-variables"` //nolint:tagliatelle
	Image                string                                      `yaml:"image"`
	Pull                 navigatorSettingsFormatPull                 `yaml:"pull"`
}

type navigatorSettingsFormatAnsibleNavigator struct { //nolint:maligned
	Color                navigatorSettingsFormatColor                `yaml:"color"`
	ExecutionEnvironment navigatorSettingsFormatExecutionEnvironment `yaml:"execution-environment"` //nolint:tagliatelle
	Logging              navigatorSettingsFormatLogging              `yaml:"logging"`
	Mode                 string                                      `yaml:"mode"`
	PlaybookArtifact     navigatorSettingsFormatPlaybookArtifact     `yaml:"playbook-artifact"` //nolint:tagliatelle
}

type navigatorSettingsFormat struct {
	AnsibleNavigator navigatorSettingsFormatAnsibleNavigator `yaml:"ansible-navigator"` //nolint:tagliatelle
}

func GenerateNavigatorSettings(settings *NavigatorSettings) (string, error) {
	settingsFormat := navigatorSettingsFormat{
		AnsibleNavigator: navigatorSettingsFormatAnsibleNavigator{
			Color: navigatorSettingsFormatColor{
				Enable: false,
				OSC4:   false,
			},
			ExecutionEnvironment: navigatorSettingsFormatExecutionEnvironment{
				ContainerEngine: settings.ContainerEngine,
				EnvironmentVariables: navigatorSettingsFormatEnvironmentVariables{
					Pass: settings.EnvironmentVariablesPass,
					Set:  settings.EnvironmentVariablesSet,
				},
				Image: settings.Image,
				Pull: navigatorSettingsFormatPull{
					Arguments: settings.PullArguments,
					Policy:    settings.PullPolicy,
				},
			},
			Logging: navigatorSettingsFormatLogging{
				Level: "warning",
			},
			Mode: "stdout",
			PlaybookArtifact: navigatorSettingsFormatPlaybookArtifact{
				Enable: true,
			},
		},
	}

	data, err := yaml.Marshal(&settingsFormat)
	if err != nil {
		return "", fmt.Errorf("failed to build ansible-navigator settings file, %w", err)
	}

	return string(data), nil
}

func CreateNavigatorSettingsFile(tempRunDir string, settingsContents string) error {
	path := filepath.Join(tempRunDir, navigatorSettingsFilename)

	err := writeFile(path, settingsContents)
	if err != nil {
		return fmt.Errorf("failed to create %s settings file, %w", NavigatorProgram, err)
	}

	return nil
}
