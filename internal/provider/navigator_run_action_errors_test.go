package provider_test

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccNavigatorRunAction_errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		variables func(*testing.T) config.Variables
		expected  *regexp.Regexp
	}{
		{
			name:     "playbook_yaml",
			expected: regexp.MustCompile("Not valid YAML"),
		},
		{
			name:     "playbook",
			expected: regexp.MustCompile("Ansible navigator run failed"),
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
				TerraformVersionChecks: []tfversion.TerraformVersionCheck{
					tfversion.SkipBelow(version.Must(version.NewVersion("1.14.0"))), // TODO replace with tfversion.Version1_14_0 when new plugin-testing version is released
				},
				Steps: []resource.TestStep{
					{
						Config:          testTerraformConfig(t, filepath.Join("navigator_run_action", "errors", test.name)),
						ConfigVariables: testConfigVariables(t, variables),
						ExpectError:     test.expected,
					},
				},
			})
		})
	}
}
