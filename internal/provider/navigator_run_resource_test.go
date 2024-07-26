package provider_test

import (
	"maps"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

const (
	navigatorRunResource = "ansible_navigator_run.test"
)

func TestAccNavigatorRun_ansible_options(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testAccFile(t, filepath.Join("navigator_run", "ansible_options")),
				ConfigVariables: testAccDefaultConfigVariables(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(navigatorRunResource, "command", regexp.MustCompile("--force-handlers --skip-tags tag1,tag2 --start-at-task task name --limit host1,host2 --tags tag3,tag4")),
				),
			},
		},
	})
}

func TestAccNavigatorRun_artifact_queries(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testAccFile(t, filepath.Join("navigator_run", "artifact_queries")),
				ConfigVariables: testAccDefaultConfigVariables(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(navigatorRunResource, "artifact_queries.stdout.result", regexp.MustCompile("ok=3")),
					resource.TestCheckResourceAttr(navigatorRunResource, "artifact_queries.file.result", "YWNj"),
				),
			},
		},
	})
}

func TestAccNavigatorRun_basic_binary_path(t *testing.T) { //nolint:paralleltest
	testAccPrependProgramsToPath(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testAccFile(t, filepath.Join("navigator_run", "basic_binary_path")),
				ConfigVariables: testAccDefaultConfigVariables(t),
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testAccFile(t, filepath.Join("navigator_run", "basic")),
				ConfigVariables: testAccDefaultConfigVariables(t),
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
				Config:          testAccFile(t, filepath.Join("navigator_run", "basic")),
				ConfigVariables: testAccDefaultConfigVariables(t),
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testAccFile(t, filepath.Join("navigator_run", "env_vars")),
				ConfigVariables: testAccDefaultConfigVariables(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
				),
			},
			{
				Config:          testAccFile(t, filepath.Join("navigator_run", "env_vars_update")),
				ConfigVariables: testAccDefaultConfigVariables(t),
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

	publicKey, privateKey := sshKeygen(t)
	port := sshServer(t, publicKey)
	variables := config.Variables{
		"private_key_data": config.StringVariable(privateKey),
		"ssh_port":         config.IntegerVariable(port),
	}
	maps.Copy(variables, testAccDefaultConfigVariables(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testAccFile(t, filepath.Join("navigator_run", "private_keys")),
				ConfigVariables: variables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
				),
			},
		},
	})
}

func TestAccNavigatorRun_pull_args(t *testing.T) {
	t.Parallel()

	arg := "--tls-verify=true"
	variables := config.Variables{
		"pull_arguments": config.ListVariable(config.StringVariable(arg)),
	}
	maps.Copy(variables, testAccDefaultConfigVariables(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testAccFile(t, filepath.Join("navigator_run", "pull_args")),
				ConfigVariables: variables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(navigatorRunResource, "artifact_queries.pull_args.result", regexp.MustCompile(arg)),
				),
			},
		},
	})
}

func TestAccNavigatorRun_relative_binary(t *testing.T) {
	t.Parallel()

	variables := config.Variables{
		"working_directory": config.StringVariable(t.TempDir()),
	}
	maps.Copy(variables, testAccDefaultConfigVariables(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testAccFile(t, filepath.Join("navigator_run", "relative_binary")),
				ConfigVariables: variables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
				),
			},
		},
	})
}

func TestAccNavigatorRun_role(t *testing.T) {
	t.Parallel()

	variables := config.Variables{
		// https://github.com/hashicorp/terraform-plugin-testing/issues/277
		"working_directory": config.StringVariable(filepath.Join("testdata", "navigator_run", "role-working-dir")),
	}
	maps.Copy(variables, testAccDefaultConfigVariables(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          testAccFile(t, filepath.Join("navigator_run", "role")),
				ConfigVariables: variables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
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
					"ansible_navigator_binary": config.StringVariable(testAccLookPath(t, "docker")),
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

			variables := testAccDefaultConfigVariables(t)
			if test.variables != nil {
				maps.Copy(variables, test.variables(t))
			}

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:          testAccFile(t, filepath.Join("navigator_run", "errors", test.name)),
						ConfigVariables: variables,
						ExpectError:     test.expected,
					},
				},
			})
		})
	}
}
