package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TODO
// 1. Test failure cases
// 2. Test provider options
// 3. Test ansible options
// 4. Test EE options (env vars and generated navigator config file)
// 5. Test triggers
// 6. Test run on destroy

const navigatorRunResource = "ansible_navigator_run.test"

func TestAccNavigatorRun_basic(t *testing.T) {
	t.Parallel()

	workingDirectory := t.TempDir()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNavigatorRunResourceConfig(t, "basic", workingDirectory),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "command"),
					resource.TestCheckNoResourceAttr(navigatorRunResource, "ansible_options"),
					resource.TestCheckNoResourceAttr(navigatorRunResource, "triggers"),
					resource.TestCheckNoResourceAttr(navigatorRunResource, "replacement_triggers"),
					resource.TestCheckNoResourceAttr(navigatorRunResource, "artifact_queries"),
				),
			},
		},
	})
}

func TestAccNavigatorRun_basic_path(t *testing.T) { //nolint:paralleltest
	workingDirectory := t.TempDir()
	testAccPrependProgramToPath(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNavigatorRunResourceConfigUsePath(t, "basic_path", workingDirectory),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "command"),
					resource.TestCheckNoResourceAttr(navigatorRunResource, "ansible_navigator_binary"),
					resource.TestCheckNoResourceAttr(navigatorRunResource, "ansible_options"),
					resource.TestCheckNoResourceAttr(navigatorRunResource, "triggers"),
					resource.TestCheckNoResourceAttr(navigatorRunResource, "replacement_triggers"),
					resource.TestCheckNoResourceAttr(navigatorRunResource, "artifact_queries"),
				),
			},
		},
	})
}

func TestAccNavigatorRun_artifact_queries(t *testing.T) {
	t.Parallel()

	workingDirectory := t.TempDir()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNavigatorRunResourceConfig(t, "artifact_queries", workingDirectory),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(navigatorRunResource, "artifact_queries.stdout.result", regexp.MustCompile("ok=3")),
					resource.TestCheckResourceAttr(navigatorRunResource, "artifact_queries.file.result", "YWNj"),
				),
			},
		},
	})
}
