package provider_test

import (
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccNavigatorRunAction_basic(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.14.0"))), // TODO replace with tfversion.Version1_14_0 when new plugin-testing version is released
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
