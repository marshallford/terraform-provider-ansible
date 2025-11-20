package provider_test

import (
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNavigatorRunAction_basic(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		// TODO use TestCheckProgressMessageContains when new plugin-testing version is released
		Steps: []resource.TestStep{
			{
				Config:          testTerraformConfig(t, filepath.Join("navigator_run_action", "basic")),
				ConfigVariables: testDefaultConfigVariables(t),
			},
		},
	})
}
