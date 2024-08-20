package provider_test

import (
	"fmt"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

const (
	navigatorRunResource = "ansible_navigator_run.test"
)

func TestAccNavigatorRunResource_ansible_options(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testTerraformFile(t, filepath.Join("navigator_run_resource", "ansible_options")),
				ConfigVariables: testDefaultConfigVariables(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(navigatorRunResource, "command", regexp.MustCompile("--force-handlers --skip-tags tag1,tag2 --start-at-task task name --limit host1,host2 --tags tag3,tag4")),
				),
			},
		},
	})
}

func TestAccNavigatorRunResource_artifact_queries(t *testing.T) {
	t.Parallel()

	fileContents := "acc"
	fileContentsUpdate := "acc_update"
	var resourceCommand, resourceCommandUpdate string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformFile(t, filepath.Join("navigator_run_resource", "artifact_queries")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"file_contents": config.StringVariable(fileContents),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(navigatorRunResource, "artifact_queries.stdout.results.0", regexp.MustCompile("ok=3")),
					testExtractResourceAttr(navigatorRunResource, "command", &resourceCommand),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("file_contents", knownvalue.StringExact(fileContents)),
				},
			},
			{
				Config: testTerraformFile(t, filepath.Join("navigator_run_resource", "artifact_queries")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"file_contents": config.StringVariable(fileContentsUpdate),
				}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectUnknownValue(navigatorRunResource, tfjsonpath.New("artifact_queries").AtMapKey("file_contents").AtMapKey("results")),
						plancheck.ExpectUnknownValue(navigatorRunResource, tfjsonpath.New("command")),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(navigatorRunResource, "artifact_queries.stdout.results.0", regexp.MustCompile("ok=3")),
					testExtractResourceAttr(navigatorRunResource, "command", &resourceCommandUpdate),
					testCheckAttributeValuesDiffer(&resourceCommand, &resourceCommandUpdate),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("file_contents", knownvalue.StringExact(fileContentsUpdate)),
				},
			},
		},
	})
}

func TestAccNavigatorRunResource_basic(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testTerraformFile(t, filepath.Join("navigator_run_resource", "basic")),
				ConfigVariables: testDefaultConfigVariables(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "playbook"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "inventory"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "working_directory"),
					// resource.TestCheckResourceAttrSet(navigatorRunResource, "execution_environment"), TODO check elements
					resource.TestCheckResourceAttrSet(navigatorRunResource, "ansible_navigator_binary"),
					// resource.TestCheckNoResourceAttr(navigatorRunResource, "ansible_options"), TODO check elements
					resource.TestCheckResourceAttr(navigatorRunResource, "ansible_options.known_hosts.#", "0"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "timezone"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "run_on_destroy"),
					resource.TestCheckNoResourceAttr(navigatorRunResource, "triggers"),
					resource.TestCheckNoResourceAttr(navigatorRunResource, "replacement_triggers"),
					resource.TestCheckNoResourceAttr(navigatorRunResource, "artifact_queries"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "command"),
					resource.TestCheckNoResourceAttr(navigatorRunResource, "timeouts"),
				),
			},
			{
				Config:          testTerraformFile(t, filepath.Join("navigator_run_resource", "basic")),
				ConfigVariables: testDefaultConfigVariables(t),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccNavigatorRunResource_binary_in_path(t *testing.T) { //nolint:paralleltest
	testPrependNavigatorToPath(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testTerraformFile(t, filepath.Join("navigator_run_resource", "binary_in_path")),
				ConfigVariables: testDefaultConfigVariables(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "command"),
				),
			},
		},
	})
}

func TestAccNavigatorRunResource_ee_disabled(t *testing.T) { //nolint:paralleltest
	testPrependPlaybookToPath(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testTerraformFile(t, filepath.Join("navigator_run_resource", "ee_disabled")),
				ConfigVariables: testDefaultConfigVariables(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "command"),
				),
			},
		},
	})
}

func TestAccNavigatorRunResource_env_vars(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformFile(t, filepath.Join("navigator_run_resource", "env_vars")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"operation": config.StringVariable("create"),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "command"),
				),
			},
			{
				Config: testTerraformFile(t, filepath.Join("navigator_run_resource", "env_vars")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"operation": config.StringVariable("update"),
				}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccNavigatorRunResource_known_hosts(t *testing.T) {
	t.Parallel()

	serverPublicKey, serverPrivateKey := testSSHKeygen(t)
	port := testSSHServer(t, "", serverPrivateKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformFile(t, filepath.Join("navigator_run_resource", "known_hosts")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"ssh_port": config.IntegerVariable(port),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "command"),
					resource.TestMatchResourceAttr(navigatorRunResource, "ansible_options.known_hosts.0", regexp.MustCompile(regexp.QuoteMeta(serverPublicKey))),
				),
			},
		},
	})
}

//nolint:dupl //TODO fix
func TestAccNavigatorRunResource_private_keys(t *testing.T) { //nolint:paralleltest
	tests := []struct {
		name      string
		variables func(*testing.T) config.Variables
		setup     func(*testing.T)
	}{
		{
			name: "ee_enabled",
			variables: func(t *testing.T) config.Variables { //nolint:thelper
				return config.Variables{
					"ee_enabled": config.BoolVariable(true),
				}
			},
			setup: func(t *testing.T) { //nolint:thelper
				t.Parallel()
			},
		},
		{
			name: "ee_disabled",
			variables: func(t *testing.T) config.Variables { //nolint:thelper
				return config.Variables{
					"ee_enabled": config.BoolVariable(false),
				}
			},
			setup: func(t *testing.T) { //nolint:thelper
				testPrependPlaybookToPath(t)
			},
		},
	}

	for _, test := range tests { //nolint:paralleltest
		t.Run(test.name, func(t *testing.T) {
			test.setup(t)

			variables := config.Variables{}
			if test.variables != nil {
				variables = test.variables(t)
			}

			clientPublicKey, clientPrivateKey := testSSHKeygen(t)
			serverPublicKey, serverPrivateKey := testSSHKeygen(t)
			port := testSSHServer(t, clientPublicKey, serverPrivateKey)

			variables["client_private_key_data"] = config.StringVariable(clientPrivateKey)
			variables["server_public_key_data"] = config.StringVariable(serverPublicKey)
			variables["ssh_port"] = config.IntegerVariable(port)

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:          testTerraformFile(t, filepath.Join("navigator_run_resource", "private_keys")),
						ConfigVariables: testConfigVariables(t, variables),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
							resource.TestCheckResourceAttrSet(navigatorRunResource, "command"),
							resource.TestMatchResourceAttr(navigatorRunResource, "command", regexp.MustCompile(fmt.Sprintf("--private-key(?s)(.*)--extra-vars %s", ansible.SSHKnownHostsFileVar))),
						),
					},
				},
			})
		})
	}
}

func TestAccNavigatorRunResource_pull_args(t *testing.T) {
	t.Parallel()

	arg := "--tls-verify=true"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformFile(t, filepath.Join("navigator_run_resource", "pull_args")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"pull_args": config.ListVariable(config.StringVariable(arg)),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(navigatorRunResource, "artifact_queries.pull_args.results.0", regexp.MustCompile(arg)),
				),
			},
		},
	})
}

func TestAccNavigatorRunResource_role(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformFile(t, filepath.Join("navigator_run_resource", "role")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					// https://github.com/hashicorp/terraform-plugin-testing/issues/277
					"working_directory": config.StringVariable(filepath.Join("testdata", "navigator_run_resource", "role-working-dir")),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "command"),
				),
			},
		},
	})
}

func TestAccNavigatorRunResource_skip_run(t *testing.T) {
	t.Parallel()

	var resourceCommand, resourceCommandUpdate string
	var queryResult, queryResultUpdate string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testTerraformFile(t, filepath.Join("navigator_run_resource", "skip_run")),
				ConfigVariables: testDefaultConfigVariables(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "command"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "artifact_queries.test.results.0"),
					testExtractResourceAttr(navigatorRunResource, "command", &resourceCommand),
					testExtractResourceAttr(navigatorRunResource, "artifact_queries.test.results.0", &queryResult),
				),
			},
			{
				Config:          testTerraformFile(t, filepath.Join("navigator_run_resource", "skip_run_update")),
				ConfigVariables: testDefaultConfigVariables(t),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "command"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "artifact_queries.test.results.0"),
					testExtractResourceAttr(navigatorRunResource, "command", &resourceCommandUpdate),
					testCheckAttributeValuesEqual(&resourceCommand, &resourceCommandUpdate),
					testExtractResourceAttr(navigatorRunResource, "artifact_queries.test.results.0", &queryResultUpdate),
					testCheckAttributeValuesEqual(&queryResult, &queryResultUpdate),
				),
			},
		},
	})
}
