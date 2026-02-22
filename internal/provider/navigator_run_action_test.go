package provider_test

import (
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccNavigatorRunAction_basic(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		// TODO use TestCheckProgressMessageContains when new plugin-testing version is released
		Steps: []resource.TestStep{
			{
				Config:          testTerraformConfig(t, filepath.Join("navigator_run_action", "basic")),
				ConfigVariables: testDefaultConfigVariables(t),
			},
		},
	})
}
