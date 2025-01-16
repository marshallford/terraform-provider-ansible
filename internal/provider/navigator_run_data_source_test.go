package provider_test

import (
	"fmt"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

const (
	navigatorRunDataSource = "data.ansible_navigator_run.test"
)

func TestAccNavigatorRunDataSource_artifact_queries(t *testing.T) {
	t.Parallel()

	commandValueDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformConfig(t, filepath.Join("navigator_run_data_source", "artifact_queries")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"file_contents": config.StringVariable(testString),
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(navigatorRunDataSource, tfjsonpath.New("command"), knownvalue.NotNull()),
					commandValueDiffer.AddStateValue(navigatorRunDataSource, tfjsonpath.New("command")),
					statecheck.ExpectKnownValue(
						navigatorRunDataSource,
						tfjsonpath.New("artifact_queries").AtMapKey("stdout").AtMapKey("results").AtSliceIndex(0),
						knownvalue.StringRegexp(regexp.MustCompile("ok=3")),
					),
					statecheck.ExpectKnownOutputValue("file_contents", knownvalue.StringExact(testString)),
				},
			},
			{
				Config: testTerraformConfig(t, filepath.Join("navigator_run_data_source", "artifact_queries")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"file_contents": config.StringVariable(testUpdateString),
				}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					commandValueDiffer.AddStateValue(navigatorRunDataSource, tfjsonpath.New("command")),
					statecheck.ExpectKnownValue(
						navigatorRunDataSource,
						tfjsonpath.New("artifact_queries").AtMapKey("stdout").AtMapKey("results").AtSliceIndex(0),
						knownvalue.StringRegexp(regexp.MustCompile("ok=3")),
					),
					statecheck.ExpectKnownOutputValue("file_contents", knownvalue.StringExact(testUpdateString)),
				},
			},
		},
	})
}

func TestAccNavigatorRunDataSource_basic(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testTerraformConfig(t, filepath.Join("navigator_run_data_source", "basic")),
				ConfigVariables: testDefaultConfigVariables(t),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(navigatorRunDataSource, tfjsonpath.New("playbook"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunDataSource, tfjsonpath.New("inventory"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunDataSource, tfjsonpath.New("working_directory"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunDataSource, tfjsonpath.New("execution_environment"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunDataSource, tfjsonpath.New("ansible_navigator_binary"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunDataSource, tfjsonpath.New("ansible_options"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunDataSource, tfjsonpath.New("timezone"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunDataSource, tfjsonpath.New("artifact_queries"), knownvalue.Null()),
					statecheck.ExpectKnownValue(navigatorRunDataSource, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunDataSource, tfjsonpath.New("command"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunDataSource, tfjsonpath.New("timeouts"), knownvalue.Null()),
					statecheck.ExpectKnownValue(navigatorRunDataSource, tfjsonpath.New("ansible_options").AtMapKey("known_hosts"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(navigatorRunDataSource, tfjsonpath.New("execution_environment").AtMapKey("container_engine"), knownvalue.StringExact("auto")),
				},
			},
		},
	})
}

func TestAccNavigatorRunDataSource_env_vars(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testTerraformConfig(t, filepath.Join("navigator_run_data_source", "env_vars")),
				ConfigVariables: testDefaultConfigVariables(t),
			},
		},
	})
}

func TestAccNavigatorRunDataSource_known_hosts(t *testing.T) {
	t.Parallel()

	serverPublicKey, serverPrivateKey := testSSHKeygen(t)
	port := testSSHServer(t, "", serverPrivateKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformConfig(t, filepath.Join("navigator_run_data_source", "known_hosts")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"ssh_port": config.IntegerVariable(port),
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						navigatorRunDataSource,
						tfjsonpath.New("ansible_options").AtMapKey("known_hosts").AtSliceIndex(0),
						knownvalue.StringRegexp(regexp.MustCompile(regexp.QuoteMeta(serverPublicKey))),
					),
				},
			},
		},
	})
}

//nolint:dupl
func TestAccNavigatorRunDataSource_private_keys(t *testing.T) { //nolint:paralleltest
	for _, test := range privateKeyTestCases() { //nolint:paralleltest
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
						Config:          testTerraformConfig(t, filepath.Join("navigator_run_data_source", "private_keys")),
						ConfigVariables: testConfigVariables(t, variables),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue(
								navigatorRunDataSource,
								tfjsonpath.New("command"),
								knownvalue.StringRegexp(regexp.MustCompile(fmt.Sprintf("--private-key(?s)(.*)--extra-vars %s", ansible.SSHKnownHostsFileVar))),
							),
						},
					},
				},
			})
		})
	}
}
