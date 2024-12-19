//nolint:dupl
package provider_test

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccNavigatorRunEphemeralResource_errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		variables func(*testing.T) config.Variables
		expected  *regexp.Regexp
	}{
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
					tfversion.SkipBelow(tfversion.Version1_10_0),
				},
				Steps: []resource.TestStep{
					{
						Config:          testTerraformConfig(t, filepath.Join("navigator_run_ephemeral_resource", "errors", test.name)),
						ConfigVariables: testConfigVariables(t, variables),
						ExpectError:     test.expected,
					},
				},
			})
		})
	}
}
