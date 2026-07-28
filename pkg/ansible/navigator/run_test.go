package navigator

import (
	"context"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
	"github.com/spf13/afero"
)

const testHostDir = "/tmp/ansible-navigator-run-test"

func testConfig(eeEnabled bool) *RunConfig {
	return &RunConfig{
		WorkingDir: "/work",
		Binary:     "/usr/bin/ansible-navigator",
		Playbook:   "- hosts: all\n",
		Inventories: []ansible.Inventory{
			{Name: "hosts", Contents: "all:\n"},
			{Name: "previous-hosts", Contents: "all:\n", Exclude: true},
		},
		ExtraVars:       []ansible.ExtraVarsFile{{Name: "vars.yaml", Contents: "key: value\n"}},
		PrivateKeys:     []ansible.PrivateKey{{Name: "key", Data: "PRIVATE KEY\n"}},
		KnownHosts:      []ansible.KnownHost{"example.com ssh-ed25519 AAAA"},
		UseKnownHosts:   true,
		HostKeyChecking: true,
		Options: &ansible.PlaybookOptions{
			ForceHandlers: true,
			SkipTags:      []string{"skip-me"},
			StartAtTask:   "task name",
			Limit:         []string{"host1", "host2"},
			Tags:          []string{"tag1"},
		},
		Settings: &Settings{
			Timeout:                  10 * time.Minute,
			EEEnabled:                eeEnabled,
			ContainerEngine:          ContainerEngineAuto,
			EnvironmentVariablesPass: []string{"SSH_AUTH_SOCK"},
			EnvironmentVariablesSet:  map[string]string{"SET_VAR": "set-value"},
			Image:                    "ghcr.io/ansible/community-ansible-dev-tools:v26.7.1",
			PullArguments:            []string{"--tls-verify=false"},
			PullPolicy:               "tag",
			ContainerOptions:         []string{"--userns=host"},
			Timezone:                 "UTC",
		},
		Env: map[string]string{"EXAMPLE_VAR": "example-value"},
	}
}

func newTestRun(t *testing.T, eeEnabled bool) (*Run, *fakeExecutor) {
	t.Helper()

	memFs := afero.NewMemMapFs()
	// createDirs uses Mkdir, not MkdirAll.
	if err := memFs.MkdirAll("/tmp", dirPermissions); err != nil {
		t.Fatalf("failed to create tmp directory: %v", err)
	}

	if err := memFs.MkdirAll("/work", dirPermissions); err != nil {
		t.Fatalf("failed to create working directory: %v", err)
	}

	exec := newFakeExecutor().
		withProgram("podman", "docker", ansible.PlaybookProgram, Program).
		withResponse("ansible-navigator --version", "ansible-navigator 26.6.0", nil).
		withResponse("ansible-playbook --version", "ansible-playbook [core 2.19.0]", nil)

	return NewRun(testHostDir, testConfig(eeEnabled), WithFs(memFs), WithExecutor(exec)), exec
}

func assertLines(t *testing.T, name string, got []string, want []string) {
	t.Helper()

	if !slices.Equal(got, want) {
		t.Errorf("%s mismatch\nwant: %q\ngot:  %q", name, want, got)
	}
}

func TestPreflightCommands(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		eeEnabled bool
		want      []string
	}{
		"host": {
			eeEnabled: false,
			want: []string{
				"ansible-playbook --version",
				"/usr/bin/ansible-navigator --version",
			},
		},
		"ee": {
			eeEnabled: true,
			want: []string{
				"podman info",
				"/usr/bin/ansible-navigator --version",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			run, exec := newTestRun(t, test.eeEnabled)

			if err := run.Preflight(context.Background()); err != nil {
				t.Fatalf("preflight failed: %v", err)
			}

			assertLines(t, "commands", exec.commandStrings(), test.want)
		})
	}
}

// MemMapFs creates parent directories implicitly on write, so assert the exact
// tree rather than relying on writes failing.
func TestSetupCreatesRunDirectory(t *testing.T) {
	t.Parallel()

	want := []string{
		testHostDir + "/",
		testHostDir + "/extra-vars/",
		testHostDir + "/extra-vars/vars.yaml",
		testHostDir + "/inventories/",
		testHostDir + "/inventories/hosts",
		testHostDir + "/inventories/previous-hosts",
		testHostDir + "/known-hosts/",
		testHostDir + "/known-hosts/known_hosts",
		testHostDir + "/playbook.yaml",
		testHostDir + "/private-keys/",
		testHostDir + "/private-keys/key",
	}

	for name, eeEnabled := range map[string]bool{"host": false, "ee": true} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			run, _ := newTestRun(t, eeEnabled)

			if err := run.Setup(); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			var paths []string

			err := afero.Walk(run.fs, testHostDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() {
					path += "/"
				}

				paths = append(paths, path)

				return nil
			})
			if err != nil {
				t.Fatalf("failed to walk run directory: %v", err)
			}

			slices.Sort(paths)
			assertLines(t, "run directory", paths, want)
		})
	}
}

func TestGenerateSettings(t *testing.T) {
	t.Parallel()

	for name, eeEnabled := range map[string]bool{"host": false, "ee": true} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			run, _ := newTestRun(t, eeEnabled)

			if err := run.Setup(); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			if err := run.writeGeneratedSettings(); err != nil {
				t.Fatalf("failed to write settings: %v", err)
			}

			contents, err := afero.ReadFile(run.fs, run.hostJoin(navigatorSettingsFilename))
			if err != nil {
				t.Fatalf("failed to read settings: %v", err)
			}

			assertGolden(t, "settings/"+name+".yaml", string(contents))
		})
	}
}

func TestGenerateCommand(t *testing.T) {
	t.Parallel()

	wantEnv := []string{
		"ANSIBLE_NAVIGATOR_CONFIG=" + testHostDir + "/ansible-navigator.yaml",
		"EXAMPLE_VAR=example-value",
		"ANSIBLE_HOST_KEY_CHECKING=true",
	}

	for name, eeEnabled := range map[string]bool{"host": false, "ee": true} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			run, exec := newTestRun(t, eeEnabled)

			if err := run.Preflight(context.Background()); err != nil {
				t.Fatalf("preflight failed: %v", err)
			}

			if err := run.Setup(); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			run.generateCommand(context.Background())

			cmd := exec.lastCommand()

			assertGoldenLines(t, "command/"+name+".txt", cmd.argv)
			assertLines(t, "env", cmd.envDelta(), wantEnv)

			if cmd.dir != "/work" {
				t.Errorf("expected command dir %q, got %q", "/work", cmd.dir)
			}
		})
	}
}

func TestPathScopes(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		eeEnabled     bool
		resolvedDir   string
		inventoryPath string
		resolvedJoin  string
	}{
		"host": {
			eeEnabled:     false,
			resolvedDir:   testHostDir,
			inventoryPath: testHostDir + "/inventories/hosts",
			resolvedJoin:  testHostDir + "/extra-vars/vars.yaml",
		},
		"ee": {
			eeEnabled:     true,
			resolvedDir:   containerRunDir,
			inventoryPath: containerRunDir + "/inventories/hosts",
			resolvedJoin:  containerRunDir + "/extra-vars/vars.yaml",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			run, _ := newTestRun(t, test.eeEnabled)

			if err := run.Setup(); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			if got := run.HostDir(); got != testHostDir {
				t.Errorf("hostDir: want %q, got %q", testHostDir, got)
			}

			if got := run.ResolvedDir(); got != test.resolvedDir {
				t.Errorf("resolvedDir: want %q, got %q", test.resolvedDir, got)
			}

			if got := run.InventoryPath("hosts"); got != test.inventoryPath {
				t.Errorf("inventoryPath: want %q, got %q", test.inventoryPath, got)
			}

			want := testHostDir + "/extra-vars/vars.yaml"
			if got := run.hostJoin(extraVarsDir, "vars.yaml"); got != want {
				t.Errorf("hostJoin: want %q, got %q", want, got)
			}

			if got := run.resolvedJoin(extraVarsDir, "vars.yaml"); got != test.resolvedJoin {
				t.Errorf("resolvedJoin: want %q, got %q", test.resolvedJoin, got)
			}
		})
	}
}

func TestCleanupRemovesRunDirectory(t *testing.T) {
	t.Parallel()

	run, _ := newTestRun(t, false)

	if err := run.Setup(); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	if err := run.Cleanup(); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	exists, err := afero.DirExists(run.fs, testHostDir)
	if err != nil {
		t.Fatalf("failed to stat run directory: %v", err)
	}

	if exists {
		t.Error("expected run directory to be removed")
	}
}
