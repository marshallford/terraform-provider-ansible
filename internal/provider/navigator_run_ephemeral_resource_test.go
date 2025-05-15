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
	navigatorRunEphemeralResource = "echo.test"
)

func TestAccNavigatorRunEphemeralResource_artifact_queries(t *testing.T) {
	t.Parallel()

	commandValueDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	commandPath := tfjsonpath.New("data").AtMapKey("ephemeral_resource").AtMapKey("command")
	stdoutArtifactQueryPath := tfjsonpath.New("data").AtMapKey("ephemeral_resource").AtMapKey("artifact_queries").AtMapKey("stdout").AtMapKey("results").AtSliceIndex(0)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformConfig(
					t,
					filepath.Join("navigator_run_ephemeral_resource", "artifact_queries"),
					filepath.Join("navigator_run_ephemeral_resource", "artifact_queries_echo_first"),
				),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"file_contents": config.StringVariable(testString),
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.first_test", commandPath, knownvalue.NotNull()),
					commandValueDiffer.AddStateValue("echo.first_test", commandPath),
					statecheck.ExpectKnownValue(
						"echo.first_test",
						stdoutArtifactQueryPath,
						knownvalue.StringRegexp(regexp.MustCompile("ok=3")),
					),
					statecheck.ExpectKnownValue(
						"echo.first_test",
						tfjsonpath.New("data").AtMapKey("file_contents"),
						knownvalue.StringExact(testString),
					),
				},
			},
			{
				Config: testTerraformConfig(
					t,
					filepath.Join("navigator_run_ephemeral_resource", "artifact_queries"),
					filepath.Join("navigator_run_ephemeral_resource", "artifact_queries_echo_second"),
				),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"file_contents": config.StringVariable(testUpdateString),
				}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					commandValueDiffer.AddStateValue("echo.second_test", commandPath),
					statecheck.ExpectKnownValue(
						"echo.second_test",
						stdoutArtifactQueryPath,
						knownvalue.StringRegexp(regexp.MustCompile("ok=3")),
					),
					statecheck.ExpectKnownValue(
						"echo.second_test",
						tfjsonpath.New("data").AtMapKey("file_contents"),
						knownvalue.StringExact(testUpdateString),
					),
				},
			},
		},
	})
}

func TestAccNavigatorRunEphemeralResource_basic(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformConfig(
					t,
					filepath.Join("navigator_run_ephemeral_resource", "basic"),
					filepath.Join("navigator_run_ephemeral_resource", "_echo"),
				),
				ConfigVariables: testDefaultConfigVariables(t),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(navigatorRunEphemeralResource, tfjsonpath.New("data").AtMapKey("playbook"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunEphemeralResource, tfjsonpath.New("data").AtMapKey("inventory"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunEphemeralResource, tfjsonpath.New("data").AtMapKey("working_directory"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunEphemeralResource, tfjsonpath.New("data").AtMapKey("execution_environment"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunEphemeralResource, tfjsonpath.New("data").AtMapKey("ansible_navigator_binary"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunEphemeralResource, tfjsonpath.New("data").AtMapKey("ansible_options"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunEphemeralResource, tfjsonpath.New("data").AtMapKey("timezone"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunEphemeralResource, tfjsonpath.New("data").AtMapKey("artifact_queries"), knownvalue.Null()),
					statecheck.ExpectKnownValue(navigatorRunEphemeralResource, tfjsonpath.New("data").AtMapKey("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunEphemeralResource, tfjsonpath.New("data").AtMapKey("command"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunEphemeralResource, tfjsonpath.New("data").AtMapKey("timeouts"), knownvalue.Null()),
					statecheck.ExpectKnownValue(navigatorRunEphemeralResource, tfjsonpath.New("data").AtMapKey("ansible_options").AtMapKey("known_hosts"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(navigatorRunEphemeralResource, tfjsonpath.New("data").AtMapKey("execution_environment").AtMapKey("container_engine"), knownvalue.StringExact("auto")),
				},
			},
		},
	})
}

func TestAccNavigatorRunEphemeralResource_env_vars(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testTerraformConfig(t, filepath.Join("navigator_run_ephemeral_resource", "env_vars")),
				ConfigVariables: testDefaultConfigVariables(t),
			},
		},
	})
}

func TestAccNavigatorRunEphemeralResource_known_hosts(t *testing.T) {
	t.Parallel()

	serverPublicKey, serverPrivateKey := testSSHKeygen(t)
	port := testSSHServer(t, "", serverPrivateKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformConfig(
					t,
					filepath.Join("navigator_run_ephemeral_resource", "known_hosts"),
					filepath.Join("navigator_run_ephemeral_resource", "_echo"),
				),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"ssh_port": config.IntegerVariable(port),
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						navigatorRunEphemeralResource,
						tfjsonpath.New("data").AtMapKey("ansible_options").AtMapKey("known_hosts").AtSliceIndex(0),
						knownvalue.StringRegexp(regexp.MustCompile(regexp.QuoteMeta(serverPublicKey))),
					),
				},
			},
		},
	})
}

func TestAccNavigatorRunEphemeralResource_private_keys(t *testing.T) { //nolint:paralleltest
	for _, test := range EETestCases() { //nolint:paralleltest
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
						Config: testTerraformConfig(
							t,
							filepath.Join("navigator_run_ephemeral_resource", "private_keys"),
							filepath.Join("navigator_run_ephemeral_resource", "_echo"),
						),
						ConfigVariables: testConfigVariables(t, variables),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue(
								navigatorRunEphemeralResource,
								tfjsonpath.New("data").AtMapKey("command"),
								knownvalue.StringRegexp(regexp.MustCompile(fmt.Sprintf("--private-key(?s)(.*)--extra-vars %s", ansible.SSHKnownHostsFileVar))),
							),
						},
					},
				},
			})
		})
	}
}
