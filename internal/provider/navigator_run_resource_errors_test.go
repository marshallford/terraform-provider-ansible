package provider_test

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNavigatorRunResource_errors_command_output(t *testing.T) {
	t.Setenv("ANSIBLE_NAVIGATOR_CONTAINER_ENGINE", "test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testTerraformConfig(t, filepath.Join("navigator_run_resource", "errors", "command_output")),
				ConfigVariables: testDefaultConfigVariables(t),
				ExpectError:     regexp.MustCompile("Ansible navigator run failed"),
			},
		},
	})
}

func TestAccNavigatorRunResource_errors_host_key_checking(t *testing.T) {
	t.Parallel()

	_, serverPrivateKey := testSSHKeygen(t)
	port := testSSHServer(t, "", serverPrivateKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformConfig(t, filepath.Join("navigator_run_resource", "errors", "host_key_checking")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"ssh_port": config.IntegerVariable(port),
				}),
				ExpectError: regexp.MustCompile("Host key verification failed"),
			},
		},
	})
}

func TestAccNavigatorRunResource_errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		variables func(*testing.T) config.Variables
		expected  *regexp.Regexp
	}{
		{
			name:     "artifact_query",
			expected: regexp.MustCompile("failed to parse JQ filter"),
		},
		{
			name:     "env_var_name_empty",
			expected: regexp.MustCompile(`must(\s)not(\s)be(\s)empty`),
		},
		{
			name:     "env_var_name_invalid",
			expected: regexp.MustCompile(`must(\s)consist(\s)only(\s)of(\s)printable(\s)ASCII`),
		},
		{
			name:     "image",
			expected: regexp.MustCompile("failed to parse container image"),
		},
		{
			name:     "known_hosts",
			expected: regexp.MustCompile("(?s)SSH known host must not be empty(.*)failed to parse SSH known host(.*)must not include multiple"),
		},
		{
			name: "navigator_preflight",
			variables: func(t *testing.T) config.Variables { //nolint:thelper
				return config.Variables{
					"ansible_navigator_binary": config.StringVariable(testLookPath(t, "docker")),
				}
			},
			expected: regexp.MustCompile("Ansible navigator preflight check"),
		},
		{
			name:     "playbook_yaml",
			expected: regexp.MustCompile("Not valid YAML"),
		},
		{
			name:     "playbook",
			expected: regexp.MustCompile("Ansible navigator run failed"),
		},
		{
			name:     "private_keys",
			expected: regexp.MustCompile(`(?s)SSH private key must be a(.*)key(\s)must(\s)be(\s)unencrypted(.*)key(\s)name(\s)can(\s)only(\s)contain`),
		},
		{
			name:     "timeout",
			expected: regexp.MustCompile("Ansible navigator run timed out"),
		},
		{
			name:     "timezone_empty",
			expected: regexp.MustCompile("IANA time zone must not be empty"),
		},
		{
			name:     "timezone_invalid",
			expected: regexp.MustCompile("IANA time zone not found"),
		},
		{
			name: "working_directory",
			variables: func(t *testing.T) config.Variables { //nolint:thelper
				return config.Variables{
					"working_directory": config.StringVariable(filepath.Join(t.TempDir(), "non-existent")),
				}
			},
			expected: regexp.MustCompile("Working directory preflight check"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			variables := config.Variables{}
			if test.variables != nil {
				variables = test.variables(t)
			}

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:          testTerraformConfig(t, filepath.Join("navigator_run_resource", "errors", test.name)),
						ConfigVariables: testConfigVariables(t, variables),
						ExpectError:     test.expected,
					},
				},
			})
		})
	}
}
