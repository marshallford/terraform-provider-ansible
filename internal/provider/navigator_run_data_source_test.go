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
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

const (
	navigatorRunDataSource = "data.ansible_navigator_run.test"
)

func TestAccNavigatorRunDataSource_artifact_queries(t *testing.T) {
	t.Parallel()

	fileContents := "acc"
	fileContentsUpdate := "acc_update"
	var dataSourceCommand, dataSourceCommandUpdate string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testTerraformFile(t, filepath.Join("navigator_run_data_source", "artifact_queries")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"file_contents": config.StringVariable(fileContents),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(navigatorRunDataSource, "artifact_queries.stdout.results.0", regexp.MustCompile("ok=3")),
					testExtractResourceAttr(navigatorRunDataSource, "command", &dataSourceCommand),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("file_contents", knownvalue.StringExact(fileContents)),
				},
			},
			{
				Config: testTerraformFile(t, filepath.Join("navigator_run_data_source", "artifact_queries")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"file_contents": config.StringVariable(fileContentsUpdate),
				}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(navigatorRunDataSource, "artifact_queries.stdout.results.0", regexp.MustCompile("ok=3")),
					testExtractResourceAttr(navigatorRunDataSource, "command", &dataSourceCommandUpdate),
					testCheckAttributeValuesDiffer(&dataSourceCommand, &dataSourceCommandUpdate),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("file_contents", knownvalue.StringExact(fileContentsUpdate)),
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
				Config:          testTerraformFile(t, filepath.Join("navigator_run_data_source", "basic")),
				ConfigVariables: testDefaultConfigVariables(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunDataSource, "playbook"),
					resource.TestCheckResourceAttrSet(navigatorRunDataSource, "inventory"),
					resource.TestCheckResourceAttrSet(navigatorRunDataSource, "working_directory"),
					// resource.TestCheckResourceAttrSet(navigatorRunDataSource, "execution_environment"), TODO check elements
					resource.TestCheckResourceAttrSet(navigatorRunDataSource, "ansible_navigator_binary"),
					// resource.TestCheckNoResourceAttr(navigatorRunDataSource, "ansible_options"), TODO check elements
					resource.TestCheckResourceAttr(navigatorRunDataSource, "ansible_options.known_hosts.#", "0"),
					resource.TestCheckResourceAttrSet(navigatorRunDataSource, "timezone"),
					resource.TestCheckNoResourceAttr(navigatorRunDataSource, "triggers"),
					resource.TestCheckNoResourceAttr(navigatorRunDataSource, "replacement_triggers"),
					resource.TestCheckNoResourceAttr(navigatorRunDataSource, "artifact_queries"),
					resource.TestCheckResourceAttrSet(navigatorRunDataSource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunDataSource, "command"),
					resource.TestCheckNoResourceAttr(navigatorRunDataSource, "timeouts"),
				),
			},
		},
	})
}

func TestAccNavigatorRunDataSource_ee_defaults(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testTerraformFile(t, filepath.Join("navigator_run_data_source", "ee_defaults")),
				ConfigVariables: testDefaultConfigVariables(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunDataSource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunDataSource, "command"),
				),
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
				Config:          testTerraformFile(t, filepath.Join("navigator_run_data_source", "env_vars")),
				ConfigVariables: testDefaultConfigVariables(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunDataSource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunDataSource, "command"),
				),
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
				Config: testTerraformFile(t, filepath.Join("navigator_run_data_source", "known_hosts")),
				ConfigVariables: testConfigVariables(t, config.Variables{
					"ssh_port": config.IntegerVariable(port),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunDataSource, "id"),
					resource.TestCheckResourceAttrSet(navigatorRunDataSource, "command"),
					resource.TestMatchResourceAttr(navigatorRunDataSource, "ansible_options.known_hosts.0", regexp.MustCompile(regexp.QuoteMeta(serverPublicKey))),
				),
			},
		},
	})
}

//nolint:dupl //TODO fix
func TestAccNavigatorRunDataSource_private_keys(t *testing.T) { //nolint:paralleltest
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
						Config:          testTerraformFile(t, filepath.Join("navigator_run_data_source", "private_keys")),
						ConfigVariables: testConfigVariables(t, variables),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttrSet(navigatorRunDataSource, "id"),
							resource.TestCheckResourceAttrSet(navigatorRunDataSource, "command"),
							resource.TestMatchResourceAttr(navigatorRunDataSource, "command", regexp.MustCompile(fmt.Sprintf("--private-key(?s)(.*)--extra-vars %s", ansible.SSHKnownHostsFileVar))),
						),
					},
				},
			})
		})
	}
}
