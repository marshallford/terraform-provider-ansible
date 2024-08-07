package provider_test

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNavigatorRunResource_errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		variables func(*testing.T) config.Variables
		expected  *regexp.Regexp
	}{
		{
			name:     "artifact_query",
			expected: regexp.MustCompile("Not a valid JSONPath expression"),
		},
		{
			name:     "env_var_name_empty",
			expected: regexp.MustCompile("must not be empty"),
		},
		{
			name:     "env_var_name_invalid",
			expected: regexp.MustCompile("must consist only of printable ASCII characters"),
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
			expected: regexp.MustCompile("Not a SSH private key(?s)(.*)Not an unencrypted SSH private key"),
		},
		{
			name:     "timeout",
			expected: regexp.MustCompile("Ansible navigator run timed out"),
		},
		{
			name:     "timezone",
			expected: regexp.MustCompile("Not a valid IANA time zone"),
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
						Config:          testTerraformFile(t, filepath.Join("navigator_run_resource", "errors", test.name)),
						ConfigVariables: testConfigVariables(t, variables),
						ExpectError:     test.expected,
					},
				},
			})
		})
	}
}
