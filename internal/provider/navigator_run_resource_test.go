package provider_test

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

const (
	navigatorRunResource = "ansible_navigator_run.test"
)

func TestAccNavigatorRun_ansible_options(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testTerraformFile(t, filepath.Join("navigator_run", "ansible_options")),
				ConfigVariables: testDefaultConfigVariables(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(navigatorRunResource, "command", regexp.MustCompile("--force-handlers --skip-tags tag1,tag2 --start-at-task task name --limit host1,host2 --tags tag3,tag4")),
				),
			},
		},
	})
}

func TestAccNavigatorRun_artifact_queries(t *testing.T) {
	t.Parallel()

	var resourceCommand, resourceCommandUpdate string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testTerraformFile(t, filepath.Join("navigator_run", "artifact_queries")),
				ConfigVariables: testDefaultConfigVariables(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(navigatorRunResource, "artifact_queries.stdout.result", regexp.MustCompile("ok=3")),
					testExtractResourceAttr(navigatorRunResource, "command", &resourceCommand),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("file_contents", knownvalue.StringExact("acc")),
				},
			},
			{
				Config:          testTerraformFile(t, filepath.Join("navigator_run", "artifact_queries_update")),
				ConfigVariables: testDefaultConfigVariables(t),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectUnknownValue(navigatorRunResource, tfjsonpath.New("artifact_queries").AtMapKey("file_contents").AtMapKey("result")),
						plancheck.ExpectUnknownValue(navigatorRunResource, tfjsonpath.New("command")),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(navigatorRunResource, "artifact_queries.stdout.result", regexp.MustCompile("ok=3")),
					testExtractResourceAttr(navigatorRunResource, "command", &resourceCommandUpdate),
					testCheckAttributeValuesDiffer(&resourceCommand, &resourceCommandUpdate),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("file_contents", knownvalue.StringExact("acc_update")),
				},
			},
		},
	})
}

func TestAccNavigatorRun_basic_binary_path(t *testing.T) { //nolint:paralleltest
	testPrependProgramsToPath(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testTerraformFile(t, filepath.Join("navigator_run", "basic_binary_path")),
				ConfigVariables: testDefaultConfigVariables(t),
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

func TestAccNavigatorRun_basic(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testTerraformFile(t, filepath.Join("navigator_run", "basic")),
				ConfigVariables: testDefaultConfigVariables(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "command"),
					resource.TestCheckNoResourceAttr(navigatorRunResource, "ansible_options"),
					resource.TestCheckNoResourceAttr(navigatorRunResource, "triggers"),
					resource.TestCheckNoResourceAttr(navigatorRunResource, "replacement_triggers"),
					resource.TestCheckNoResourceAttr(navigatorRunResource, "artifact_queries"),
				),
			},
			{
				Config:          testTerraformFile(t, filepath.Join("navigator_run", "basic")),
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

func TestAccNavigatorRun_env_vars(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testTerraformFile(t, filepath.Join("navigator_run", "env_vars")),
				ConfigVariables: testDefaultConfigVariables(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "command"),
				),
			},
			{
				Config:          testTerraformFile(t, filepath.Join("navigator_run", "env_vars_update")),
				ConfigVariables: testDefaultConfigVariables(t),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccNavigatorRun_private_keys(t *testing.T) {
	t.Parallel()

	publicKey, privateKey := testSSHKeygen(t)
	port := testSSHServer(t, publicKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformFile(t, filepath.Join("navigator_run", "private_keys")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"private_key_data": config.StringVariable(privateKey),
					"ssh_port":         config.IntegerVariable(port),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "command"),
				),
			},
		},
	})
}

func TestAccNavigatorRun_pull_args(t *testing.T) {
	t.Parallel()

	arg := "--tls-verify=true"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformFile(t, filepath.Join("navigator_run", "pull_args")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"pull_arguments": config.ListVariable(config.StringVariable(arg)),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(navigatorRunResource, "artifact_queries.pull_args.result", regexp.MustCompile(arg)),
				),
			},
		},
	})
}

func TestAccNavigatorRun_relative_binary(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformFile(t, filepath.Join("navigator_run", "relative_binary")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"working_directory": config.StringVariable(t.TempDir()),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "command"),
				),
			},
		},
	})
}

func TestAccNavigatorRun_role(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformFile(t, filepath.Join("navigator_run", "role")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					// https://github.com/hashicorp/terraform-plugin-testing/issues/277
					"working_directory": config.StringVariable(filepath.Join("testdata", "navigator_run", "role-working-dir")),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "command"),
				),
			},
		},
	})
}

func TestAccNavigatorRun_skip_run(t *testing.T) {
	t.Parallel()

	var resourceCommand, resourceCommandUpdate string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testTerraformFile(t, filepath.Join("navigator_run", "skip_run")),
				ConfigVariables: testDefaultConfigVariables(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "command"),
					testExtractResourceAttr(navigatorRunResource, "command", &resourceCommand),
				),
			},
			{
				Config: testTerraformFile(t, filepath.Join("navigator_run", "skip_run_update")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"ansible_navigator_binary": config.StringVariable(acctest.RandString(8)),
				}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunResource, "command"),
					testExtractResourceAttr(navigatorRunResource, "command", &resourceCommandUpdate),
					testCheckAttributeValuesEqual(&resourceCommand, &resourceCommandUpdate),
				),
			},
		},
	})
}

func TestAccNavigatorRun_errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		variables func(*testing.T) config.Variables
		expected  *regexp.Regexp
	}{
		{
			name:     "artifact_query",
			expected: regexp.MustCompile("Not a valid JSONPath expression"),
		},
		{
			name:     "env_var_name",
			expected: regexp.MustCompile("must not be empty|must consist only of printable ASCII characters"),
		},
		{
			name: "navigator_preflight",
			variables: func(t *testing.T) config.Variables { //nolint:thelper
				return config.Variables{
					"ansible_navigator_binary": config.StringVariable(testLookPath(t, "docker")),
				}
			},
			expected: regexp.MustCompile("Ansible navigator preflight check"),
		},
		{
			name:     "playbook_yaml",
			expected: regexp.MustCompile("Not valid YAML"),
		},
		{
			name:     "playbook",
			expected: regexp.MustCompile("Ansible navigator run failed"),
		},
		{
			name:     "private_keys",
			expected: regexp.MustCompile("Not a SSH private key|Not an unencrypted SSH private key"),
		},
		{
			name:     "timeout",
			expected: regexp.MustCompile("Ansible navigator run timed out"),
		},
		{
			name:     "timezone",
			expected: regexp.MustCompile("Not a valid IANA time zone"),
		},
		{
			name: "working_directory",
			variables: func(t *testing.T) config.Variables { //nolint:thelper
				return config.Variables{
					"working_directory": config.StringVariable(filepath.Join(t.TempDir(), "non-existent")),
				}
			},
			expected: regexp.MustCompile("Working directory preflight check"),
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
				Steps: []resource.TestStep{
					{
						Config:          testTerraformFile(t, filepath.Join("navigator_run", "errors", test.name)),
						ConfigVariables: testConfigVariables(t, variables),
						ExpectError:     test.expected,
					},
				},
			})
		})
	}
}
