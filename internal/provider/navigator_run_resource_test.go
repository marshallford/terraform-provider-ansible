package provider_test

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const navigatorRunResource = "ansible_navigator_run.test"

func TestAccNavigatorRun_ansible_options(t *testing.T) {
	t.Parallel()

	workingDirectory := t.TempDir()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResource(t, filepath.Join("navigator_run", "ansible_options"), testAccAbsProgramPath(t), workingDirectory),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(navigatorRunResource, "command", regexp.MustCompile("--force-handlers --skip-tags tag1,tag2 --start-at-task task name --limit host1,host2 --tags tag3,tag4")),
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
				Config: testAccResource(t, filepath.Join("navigator_run", "artifact_queries"), testAccAbsProgramPath(t), workingDirectory),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(navigatorRunResource, "artifact_queries.stdout.result", regexp.MustCompile("ok=3")),
					resource.TestCheckResourceAttr(navigatorRunResource, "artifact_queries.file.result", "YWNj"),
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
				Config: testAccResource(t, filepath.Join("navigator_run", "basic_path"), workingDirectory),
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

	workingDirectory := t.TempDir()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResource(t, filepath.Join("navigator_run", "basic"), testAccAbsProgramPath(t), workingDirectory),
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

func TestAccNavigatorRun_env_vars(t *testing.T) {
	t.Parallel()

	workingDirectory := t.TempDir()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResource(t, filepath.Join("navigator_run", "env_vars"), testAccAbsProgramPath(t), workingDirectory),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(navigatorRunResource, "id"),
				),
			},
		},
	})
}

func TestAccNavigatorRun_private_keys(t *testing.T) {
	t.Parallel()

	workingDirectory := t.TempDir()

	publicKey, privateKey := sshKeygen(t)
	port := sshServer(t, publicKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResource(t, filepath.Join("navigator_run", "private_keys"), testAccAbsProgramPath(t), workingDirectory, port, privateKey),
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
		name           string
		resourceFormat func(*testing.T) []any
		expected       *regexp.Regexp
	}{
		{
			name: "artifact_query",
			resourceFormat: func(t *testing.T) []any { //nolint:thelper
				return []any{testAccAbsProgramPath(t), t.TempDir()}
			},
			expected: regexp.MustCompile("Playbook artifact queries failed"),
		},
		{
			name: "env_var_name",
			resourceFormat: func(t *testing.T) []any { //nolint:thelper
				return []any{testAccAbsProgramPath(t), t.TempDir()}
			},
			expected: regexp.MustCompile("must not be empty|must consist only of printable ASCII characters"),
		},
		{
			name: "navigator_preflight",
			resourceFormat: func(t *testing.T) []any { //nolint:thelper
				return []any{testAccLookPath(t, "docker"), t.TempDir()}
			},
			expected: regexp.MustCompile("Ansible navigator preflight check"),
		},
		{
			name: "playbook_yaml",
			resourceFormat: func(t *testing.T) []any { //nolint:thelper
				return []any{testAccAbsProgramPath(t), t.TempDir()}
			},
			expected: regexp.MustCompile("Not valid YAML"),
		},
		{
			name: "playbook",
			resourceFormat: func(t *testing.T) []any { //nolint:thelper
				return []any{testAccAbsProgramPath(t), t.TempDir()}
			},
			expected: regexp.MustCompile("Ansible navigator run failed"),
		},
		{
			name: "private_keys",
			resourceFormat: func(t *testing.T) []any { //nolint:thelper
				return []any{testAccAbsProgramPath(t), t.TempDir()}
			},
			expected: regexp.MustCompile("Not a SSH private key|Not an unencrypted SSH private key"),
		},
		{
			name: "timeout",
			resourceFormat: func(t *testing.T) []any { //nolint:thelper
				return []any{testAccAbsProgramPath(t), t.TempDir()}
			},
			expected: regexp.MustCompile("Ansible navigator run timed out"),
		},
		{
			name: "working_directory",
			resourceFormat: func(t *testing.T) []any { //nolint:thelper
				return []any{testAccAbsProgramPath(t), filepath.Join(t.TempDir(), "non-existent")}
			},
			expected: regexp.MustCompile("Working directory preflight check"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      testAccResource(t, filepath.Join("navigator_run", "errors", test.name), test.resourceFormat(t)...),
						ExpectError: test.expected,
					},
				},
			})
		})
	}
}
