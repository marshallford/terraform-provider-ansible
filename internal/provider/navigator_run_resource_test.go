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
	navigatorRunResource = "ansible_navigator_run.test"
)

func TestAccNavigatorRunResource_ansible_options(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testTerraformConfig(t, filepath.Join("navigator_run_resource", "ansible_options")),
				ConfigVariables: testDefaultConfigVariables(t),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						navigatorRunResource,
						tfjsonpath.New("command"),
						knownvalue.StringRegexp(regexp.MustCompile("--force-handlers --skip-tags tag1,tag2 --start-at-task task name --limit host1,host2 --tags tag3,tag4")),
					),
				},
			},
		},
	})
}

func TestAccNavigatorRunResource_artifact_queries(t *testing.T) {
	t.Parallel()

	commandValueDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformConfig(t, filepath.Join("navigator_run_resource", "artifact_queries")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"file_contents": config.StringVariable(testString),
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					commandValueDiffer.AddStateValue(navigatorRunResource, tfjsonpath.New("command")),
					statecheck.ExpectKnownValue(
						navigatorRunResource,
						tfjsonpath.New("artifact_queries").AtMapKey("stdout").AtMapKey("results").AtSliceIndex(0),
						knownvalue.StringRegexp(regexp.MustCompile("ok=3")),
					),
					statecheck.ExpectKnownOutputValue("file_contents", knownvalue.StringExact(testString)),
				},
			},
			{
				Config: testTerraformConfig(t, filepath.Join("navigator_run_resource", "artifact_queries")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"file_contents": config.StringVariable(testUpdateString),
				}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction(navigatorRunResource, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					commandValueDiffer.AddStateValue(navigatorRunResource, tfjsonpath.New("command")),
					statecheck.ExpectKnownValue(
						navigatorRunResource,
						tfjsonpath.New("artifact_queries").AtMapKey("stdout").AtMapKey("results").AtSliceIndex(0),
						knownvalue.StringRegexp(regexp.MustCompile("ok=3")),
					),
					statecheck.ExpectKnownOutputValue("file_contents", knownvalue.StringExact(testUpdateString)),
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
				Config:          testTerraformConfig(t, filepath.Join("navigator_run_resource", "basic")),
				ConfigVariables: testDefaultConfigVariables(t),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(navigatorRunResource, tfjsonpath.New("playbook"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunResource, tfjsonpath.New("inventory"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunResource, tfjsonpath.New("working_directory"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunResource, tfjsonpath.New("execution_environment"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunResource, tfjsonpath.New("ansible_navigator_binary"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunResource, tfjsonpath.New("ansible_options"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunResource, tfjsonpath.New("timezone"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunResource, tfjsonpath.New("run_on_destroy"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunResource, tfjsonpath.New("triggers"), knownvalue.Null()),
					statecheck.ExpectKnownValue(navigatorRunResource, tfjsonpath.New("artifact_queries"), knownvalue.Null()),
					statecheck.ExpectKnownValue(navigatorRunResource, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunResource, tfjsonpath.New("command"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(navigatorRunResource, tfjsonpath.New("timeouts"), knownvalue.Null()),
					statecheck.ExpectKnownValue(navigatorRunResource, tfjsonpath.New("ansible_options").AtMapKey("known_hosts"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(navigatorRunResource, tfjsonpath.New("execution_environment").AtMapKey("container_engine"), knownvalue.StringExact("auto")),
				},
			},
			{
				Config:          testTerraformConfig(t, filepath.Join("navigator_run_resource", "basic")),
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
				Config:          testTerraformConfig(t, filepath.Join("navigator_run_resource", "binary_in_path")),
				ConfigVariables: testDefaultConfigVariables(t),
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
				Config:          testTerraformConfig(t, filepath.Join("navigator_run_resource", "ee_disabled")),
				ConfigVariables: testDefaultConfigVariables(t),
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
				Config: testTerraformConfig(t, filepath.Join("navigator_run_resource", "env_vars")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"operation": config.StringVariable("create"),
				}),
			},
			{
				Config: testTerraformConfig(t, filepath.Join("navigator_run_resource", "env_vars")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"operation": config.StringVariable("update"),
				}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction(navigatorRunResource, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccNavigatorRunResource_extra_vars(t *testing.T) { //nolint:paralleltest
	for _, test := range EETestCases() { //nolint:paralleltest
		t.Run(test.name, func(t *testing.T) {
			test.setup(t)

			variables := config.Variables{}
			if test.variables != nil {
				variables = test.variables(t)
			}

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testTerraformConfig(t, filepath.Join("navigator_run_resource", "extra_vars")),
						ConfigVariables: testConfigVariables(t, variables, config.Variables{
							"revision": config.IntegerVariable(1),
						}),
					},
					{
						Config: testTerraformConfig(t, filepath.Join("navigator_run_resource", "extra_vars")),
						ConfigVariables: testConfigVariables(t, variables, config.Variables{
							"revision": config.IntegerVariable(1),
						}),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectEmptyPlan(),
							},
						},
					},
					{
						Config: testTerraformConfig(t, filepath.Join("navigator_run_resource", "extra_vars")),
						ConfigVariables: testConfigVariables(t, variables, config.Variables{
							"revision": config.IntegerVariable(2),
						}),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectNonEmptyPlan(),
								plancheck.ExpectResourceAction(navigatorRunResource, plancheck.ResourceActionUpdate),
							},
						},
					},
				},
			})
		})
	}
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
				Config: testTerraformConfig(t, filepath.Join("navigator_run_resource", "known_hosts")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"ssh_port": config.IntegerVariable(port),
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						navigatorRunResource,
						tfjsonpath.New("ansible_options").AtMapKey("known_hosts").AtSliceIndex(0),
						knownvalue.StringRegexp(regexp.MustCompile(regexp.QuoteMeta(serverPublicKey))),
					),
				},
			},
		},
	})
}

func TestAccNavigatorRunResource_previous_inventory(t *testing.T) { //nolint:paralleltest
	for _, test := range EETestCases() { //nolint:paralleltest
		t.Run(test.name, func(t *testing.T) {
			test.setup(t)

			variables := config.Variables{}
			if test.variables != nil {
				variables = test.variables(t)
			}

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testTerraformConfig(t, filepath.Join("navigator_run_resource", "previous_inventory")),
						ConfigVariables: testConfigVariables(t, variables, config.Variables{
							"inventory_file": config.StringVariable(filepath.Join("testdata", "navigator_run_resource", "previous_inventory", "create_inventory.yaml")),
						}),
					},
					{
						Config: testTerraformConfig(t, filepath.Join("navigator_run_resource", "previous_inventory")),
						ConfigVariables: testConfigVariables(t, variables, config.Variables{
							"inventory_file": config.StringVariable(filepath.Join("testdata", "navigator_run_resource", "previous_inventory", "update_inventory.yaml")),
						}),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectNonEmptyPlan(),
								plancheck.ExpectResourceAction(navigatorRunResource, plancheck.ResourceActionUpdate),
							},
						},
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownOutputValue("previous_hosts", knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("a"),
								knownvalue.StringExact("b"),
								knownvalue.StringExact("c"),
							})),
						},
					},
				},
			})
		})
	}
}

//nolint:dupl
func TestAccNavigatorRunResource_private_keys(t *testing.T) { //nolint:paralleltest
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
						Config:          testTerraformConfig(t, filepath.Join("navigator_run_resource", "private_keys")),
						ConfigVariables: testConfigVariables(t, variables),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue(
								navigatorRunResource,
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

func TestAccNavigatorRunResource_pull_args(t *testing.T) {
	t.Parallel()

	arg := "--tls-verify=true"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformConfig(t, filepath.Join("navigator_run_resource", "pull_args")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"pull_args": config.ListVariable(config.StringVariable(arg)),
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("pull_arg", knownvalue.StringExact(arg)),
				},
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
				Config: testTerraformConfig(t, filepath.Join("navigator_run_resource", "role")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					// https://github.com/hashicorp/terraform-plugin-testing/issues/277
					"working_directory": config.StringVariable(filepath.Join("testdata", "navigator_run_resource", "role-working-dir")),
				}),
			},
		},
	})
}

func TestAccNavigatorRunResource_skip_run(t *testing.T) {
	t.Parallel()

	commandValueSame := statecheck.CompareValue(compare.ValuesSame())
	queryResultValueSame := statecheck.CompareValue(compare.ValuesSame())
	queryResultPath := tfjsonpath.New("artifact_queries").AtMapKey("test").AtMapKey("results").AtSliceIndex(0)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testTerraformConfig(t, filepath.Join("navigator_run_resource", "skip_run")),
				ConfigVariables: testDefaultConfigVariables(t),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(navigatorRunResource, tfjsonpath.New("command"), knownvalue.NotNull()),
					commandValueSame.AddStateValue(navigatorRunResource, tfjsonpath.New("command")),
					statecheck.ExpectKnownValue(navigatorRunResource, queryResultPath, knownvalue.NotNull()),
					queryResultValueSame.AddStateValue(navigatorRunResource, queryResultPath),
				},
			},
			{
				Config:          testTerraformConfig(t, filepath.Join("navigator_run_resource", "skip_run_update")),
				ConfigVariables: testDefaultConfigVariables(t),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction(navigatorRunResource, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(navigatorRunResource, tfjsonpath.New("command"), knownvalue.NotNull()),
					commandValueSame.AddStateValue(navigatorRunResource, tfjsonpath.New("command")),
					statecheck.ExpectKnownValue(navigatorRunResource, queryResultPath, knownvalue.NotNull()),
					queryResultValueSame.AddStateValue(navigatorRunResource, queryResultPath),
				},
			},
		},
	})
}

func TestAccNavigatorRunResource_trigger_known_hosts(t *testing.T) {
	t.Parallel()

	_, serverPrivateKeyA := testSSHKeygen(t)
	_, serverPrivateKeyB := testSSHKeygen(t)
	portA := testSSHServer(t, "", serverPrivateKeyA)
	portB := testSSHServer(t, "", serverPrivateKeyB)

	knownHostsValueSame := statecheck.CompareValue(compare.ValuesSame())
	knownHostsValueDiffer := statecheck.CompareValue(compare.ValuesDiffer())
	knownHostsPath := tfjsonpath.New("ansible_options").AtMapKey("known_hosts")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformConfig(t, filepath.Join("navigator_run_resource", "trigger_known_hosts")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"ssh_port": config.IntegerVariable(portA),
					"trigger":  config.StringVariable("a"),
				}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectUnknownValue(navigatorRunResource, knownHostsPath),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(navigatorRunResource, knownHostsPath, knownvalue.NotNull()),
					knownHostsValueSame.AddStateValue(navigatorRunResource, knownHostsPath),
					knownHostsValueDiffer.AddStateValue(navigatorRunResource, knownHostsPath),
				},
			},
			{
				Config: testTerraformConfig(t, filepath.Join("navigator_run_resource", "trigger_known_hosts")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"ssh_port": config.IntegerVariable(portB),
					"trigger":  config.StringVariable("a"),
				}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction(navigatorRunResource, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(navigatorRunResource, knownHostsPath, knownvalue.NotNull()),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(navigatorRunResource, knownHostsPath, knownvalue.NotNull()),
					knownHostsValueSame.AddStateValue(navigatorRunResource, knownHostsPath),
				},
			},
			{
				Config: testTerraformConfig(t, filepath.Join("navigator_run_resource", "trigger_known_hosts")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"ssh_port": config.IntegerVariable(portB),
					"trigger":  config.StringVariable("b"),
				}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction(navigatorRunResource, plancheck.ResourceActionUpdate),
						plancheck.ExpectUnknownValue(navigatorRunResource, knownHostsPath),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(navigatorRunResource, knownHostsPath, knownvalue.NotNull()),
					knownHostsValueDiffer.AddStateValue(navigatorRunResource, knownHostsPath),
				},
			},
		},
	})
}

func TestAccNavigatorRunResource_trigger_run(t *testing.T) {
	t.Parallel()

	commandValueDiffer := statecheck.CompareValue(compare.ValuesDiffer())
	queryResultValueDiffer := statecheck.CompareValue(compare.ValuesDiffer())
	queryResultPath := tfjsonpath.New("artifact_queries").AtMapKey("test").AtMapKey("results").AtSliceIndex(0)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformConfig(t, filepath.Join("navigator_run_resource", "trigger_run")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"trigger": config.StringVariable("a"),
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(navigatorRunResource, tfjsonpath.New("command"), knownvalue.NotNull()),
					commandValueDiffer.AddStateValue(navigatorRunResource, tfjsonpath.New("command")),
					statecheck.ExpectKnownValue(navigatorRunResource, queryResultPath, knownvalue.NotNull()),
					queryResultValueDiffer.AddStateValue(navigatorRunResource, queryResultPath),
				},
			},
			{
				Config: testTerraformConfig(t, filepath.Join("navigator_run_resource", "trigger_run")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"trigger": config.StringVariable("b"),
				}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction(navigatorRunResource, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(navigatorRunResource, tfjsonpath.New("command"), knownvalue.NotNull()),
					commandValueDiffer.AddStateValue(navigatorRunResource, tfjsonpath.New("command")),
					statecheck.ExpectKnownValue(navigatorRunResource, queryResultPath, knownvalue.NotNull()),
					queryResultValueDiffer.AddStateValue(navigatorRunResource, queryResultPath),
				},
			},
		},
	})
}
