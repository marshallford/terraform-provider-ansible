package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TODO
// 1. Test failure cases
//   a. GenerateNavigatorSettings
//   b. WorkingDirectoryPreflight [x]
//   c. ContainerEnginePreflight [requires PATH modification]
//   d. ansible-navigator not in path [requires PATH modification]
//   e. NavigatorPreflight [x]
//   f. base_run_directory unwritable
//   g. timeout
//   h. query error?
//   i. playbook error [x]
// 2. Test provider options
// 3. Test ansible options [x]
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
				Config: testAccFixture(t, "basic", testAccAbsProgramPath(t), workingDirectory),
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
				Config: testAccFixture(t, "basic_path", workingDirectory),
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

func TestAccNavigatorRun_env_vars(t *testing.T) {
	t.Parallel()

	workingDirectory := t.TempDir()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFixture(t, "env_vars", testAccAbsProgramPath(t), workingDirectory),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
				),
			},
		},
	})
}

func TestAccNavigatorRun_ansible_options(t *testing.T) {
	t.Parallel()

	workingDirectory := t.TempDir()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFixture(t, "ansible_options", testAccAbsProgramPath(t), workingDirectory),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(navigatorRunResource, "command", regexp.MustCompile("--force-handlers --limit host1,host2 --tags tag1,tag2")),
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
				Config: testAccFixture(t, "artifact_queries", testAccAbsProgramPath(t), workingDirectory),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(navigatorRunResource, "artifact_queries.stdout.result", regexp.MustCompile("ok=3")),
					resource.TestCheckResourceAttr(navigatorRunResource, "artifact_queries.file.result", "YWNj"),
				),
			},
		},
	})
}

func TestAccNavigatorRun_errors(t *testing.T) {
	t.Parallel()

	testTable := map[string]struct {
		config   func(*testing.T, string) string
		expected *regexp.Regexp
	}{
		"playbook": {
			config: func(t *testing.T, workingDirectory string) string {
				t.Helper()

				return testAccFixture(t, "playbook_error", testAccAbsProgramPath(t), workingDirectory)
			},
			expected: regexp.MustCompile("ansible-navigator run command failed|failed=1"),
		},
		"working_directory": {
			config: func(t *testing.T, workingDirectory string) string {
				t.Helper()

				return testAccFixture(t, "working_directory_error", testAccAbsProgramPath(t), fmt.Sprintf("%s/non-existent-dir", workingDirectory))
			},
			expected: regexp.MustCompile("Working directory preflight check|directory is not valid"),
		},
		"ansible_navigator": {
			config: func(t *testing.T, workingDirectory string) string {
				t.Helper()

				return testAccFixture(t, "ansible_navigator_error", testAccLookPath(t, "docker"), workingDirectory)
			},
			expected: regexp.MustCompile("Ansible navigator preflight check|ansible-navigator is not functional"),
		},
		"env_var_name": {
			config: func(t *testing.T, workingDirectory string) string {
				t.Helper()

				return testAccFixture(t, "env_var_name_error", testAccAbsProgramPath(t), workingDirectory)
			},
			expected: regexp.MustCompile("Not a environment variable name"),
		},
		"playbook_yaml": {
			config: func(t *testing.T, workingDirectory string) string {
				t.Helper()

				return testAccFixture(t, "playbook_yaml_error", testAccAbsProgramPath(t), workingDirectory)
			},
			expected: regexp.MustCompile("Not valid YAML"),
		},
	}

	for name, test := range testTable {
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			workingDirectory := t.TempDir()

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      test.config(t, workingDirectory),
						ExpectError: test.expected,
					},
				},
			})
		})
	}
}
