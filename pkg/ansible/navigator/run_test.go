package navigator

import (
	"context"
	"os"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
	"github.com/spf13/afero"
)

const testHostDir = "/tmp/ansible-navigator-run-test"

func testConfig(eeEnabled bool) RunConfig {
	return RunConfig{
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
		Options: ansible.PlaybookOptions{
			ForceHandlers: true,
			SkipTags:      []string{"skip-me"},
			StartAtTask:   "task name",
			Limit:         []string{"host1", "host2"},
			Tags:          []string{"tag1"},
		},
		Settings: Settings{
			Timeout:  10 * time.Minute,
			Timezone: "UTC",
			ExecutionEnvironment: ExecutionEnvironment{
				Enabled:         eeEnabled,
				ContainerEngine: ContainerEngineAuto,
				Image:           "ghcr.io/ansible/community-ansible-dev-tools:v26.7.1",
				Pull: Pull{
					Arguments: []string{"--tls-verify=false"},
					Policy:    PullPolicyTag,
				},
				EnvironmentVariables: EnvironmentVariables{
					Pass: []string{"SSH_AUTH_SOCK"},
					Set:  map[string]string{"SET_VAR": "set-value"},
				},
				ContainerOptions: []string{"--userns=host"},
			},
		},
	}
}

func newTestRun(t *testing.T, eeEnabled bool) (*Run, *fakeExecutor) {
	t.Helper()

	memFs := afero.NewMemMapFs()
	if err := memFs.MkdirAll("/tmp", dirPermissions); err != nil {
		t.Fatalf("failed to create tmp directory: %v", err)
	}

	if err := memFs.MkdirAll("/work", dirPermissions); err != nil {
		t.Fatalf("failed to create working directory: %v", err)
	}

	exec := newFakeExecutor().
		withProgram(ContainerEnginePodman.String(), ContainerEngineDocker.String(), ansible.PlaybookProgram, Program).
		withResponse(Program+" --version", Program+" 26.6.0", nil).
		withResponse(ansible.PlaybookProgram+" --version", ansible.PlaybookProgram+" [core 2.19.0]", nil)

	run := NewRun(testHostDir, testConfig(eeEnabled), WithFs(memFs), WithExecutor(exec))

	// Not in sorted order, so the goldens pin that the run sorts them.
	run.SetEnv("ZULU_VAR", "zulu-value")
	run.SetEnv("ALPHA_VAR", "alpha-value")
	run.SetEnv("EXAMPLE_VAR", "example-value")

	return run, exec
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

func TestSetupCreatesRunDirectory(t *testing.T) {
	t.Parallel()

	want := []string{
		testHostDir + "/",
		testHostDir + "/ansible-navigator.yaml",
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

			var got []string

			err := afero.Walk(run.fs, testHostDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() {
					path += "/"
				}

				got = append(got, path)

				return nil
			})
			if err != nil {
				t.Fatalf("failed to walk run directory: %v", err)
			}

			slices.Sort(got)
			assertLines(t, "run directory", got, want)
		})
	}
}

func TestSettingsGenerate(t *testing.T) {
	t.Parallel()

	for name, eeEnabled := range map[string]bool{"host": false, "ee": true} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			run, _ := newTestRun(t, eeEnabled)

			// Replays the real order, so a future dependency on preflight
			// cannot slip in unnoticed.
			if err := run.Preflight(context.Background()); err != nil {
				t.Fatalf("preflight failed: %v", err)
			}

			if err := run.Setup(); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			contents, err := afero.ReadFile(run.fs, run.hostJoin(navigatorSettingsFilename))
			if err != nil {
				t.Fatalf("failed to read settings: %v", err)
			}

			assertGolden(t, "settings/"+name+".yaml", string(contents))
		})
	}
}

func TestNavigatorCommand(t *testing.T) {
	t.Parallel()

	wantEnv := []string{
		"ANSIBLE_NAVIGATOR_CONFIG=" + testHostDir + "/ansible-navigator.yaml",
		"ALPHA_VAR=alpha-value",
		"EXAMPLE_VAR=example-value",
		"ZULU_VAR=zulu-value",
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

			command := run.navigatorCommand()

			assertGoldenLines(t, "command/"+name+".txt", append([]string{command.Name}, command.Args...))
			assertLines(t, "env", exec.envDelta(command), wantEnv)

			if command.Dir != "/work" {
				t.Errorf("expected command dir %q, got %q", "/work", command.Dir)
			}
		})
	}
}

func TestRunDirs(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		eeEnabled     bool
		playbookDir   string
		inventoryPath string
		playbookJoin  string
	}{
		"host": {
			eeEnabled:     false,
			playbookDir:   testHostDir,
			inventoryPath: testHostDir + "/inventories/hosts",
			playbookJoin:  testHostDir + "/extra-vars/vars.yaml",
		},
		"ee": {
			eeEnabled:     true,
			playbookDir:   containerRunDir,
			inventoryPath: containerRunDir + "/inventories/hosts",
			playbookJoin:  containerRunDir + "/extra-vars/vars.yaml",
		},
	}

	// No Setup call: NewRun resolves these.
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			run, _ := newTestRun(t, test.eeEnabled)

			if got := run.HostDir(); got != testHostDir {
				t.Errorf("hostDir: want %q, got %q", testHostDir, got)
			}

			if got := run.NavigatorDir(); got != testHostDir {
				t.Errorf("navigatorDir: want %q, got %q", testHostDir, got)
			}

			if got := run.PlaybookDir(); got != test.playbookDir {
				t.Errorf("playbookDir: want %q, got %q", test.playbookDir, got)
			}

			if got := run.InventoryPath("hosts"); got != test.inventoryPath {
				t.Errorf("inventoryPath: want %q, got %q", test.inventoryPath, got)
			}

			want := testHostDir + "/extra-vars/vars.yaml"
			if got := run.hostJoin(extraVarsDir, "vars.yaml"); got != want {
				t.Errorf("hostJoin: want %q, got %q", want, got)
			}

			if got := run.navigatorJoin(extraVarsDir, "vars.yaml"); got != want {
				t.Errorf("navigatorJoin: want %q, got %q", want, got)
			}

			if got := run.playbookJoin(extraVarsDir, "vars.yaml"); got != test.playbookJoin {
				t.Errorf("playbookJoin: want %q, got %q", test.playbookJoin, got)
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

func TestResolveNavigatorBinary(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		binary    string
		onPath    []string
		want      string
		expectErr bool
	}{
		"looked_up_when_unset": {
			onPath: []string{Program},
			want:   "/usr/bin/ansible-navigator",
		},
		"error_when_unset_and_not_on_path": {
			expectErr: true,
		},
		"absolute_path_kept": {
			binary: "/opt/bin/ansible-navigator",
			want:   "/opt/bin/ansible-navigator",
		},
		"relative_path_made_absolute": {
			binary: "./venv/bin/ansible-navigator",
			want:   "/abs/venv/bin/ansible-navigator",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config := testConfig(false)
			config.Binary = test.binary

			exec := newFakeExecutor().withProgram(test.onPath...)
			run := NewRun(testHostDir, config, WithFs(afero.NewMemMapFs()), WithExecutor(exec))

			err := run.resolveNavigatorBinary()

			if test.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if run.resolved.navigatorBinary != "" {
					t.Errorf("expected no binary recorded, got %q", run.resolved.navigatorBinary)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if run.resolved.navigatorBinary != test.want {
				t.Errorf("want %q, got %q", test.want, run.resolved.navigatorBinary)
			}
		})
	}
}

// A Set key colliding with a run variable, and a non-empty mount list, so that
// a missing clone writes through to the caller.
func aliasProneConfig() RunConfig {
	config := testConfig(true)

	execEnv := &config.Settings.ExecutionEnvironment
	execEnv.EnvironmentVariables.Set = map[string]string{"RUN_VAR": "caller-value"}
	execEnv.VolumeMounts = []VolumeMount{{Src: "/caller", Dest: "/caller"}}

	return config
}

func TestRunDoesNotMutateConfig(t *testing.T) {
	t.Parallel()

	config := aliasProneConfig()

	memFs := afero.NewMemMapFs()
	for _, dir := range []string{"/tmp", "/work"} {
		if err := memFs.MkdirAll(dir, dirPermissions); err != nil {
			t.Fatalf("failed to create %s: %v", dir, err)
		}
	}

	exec := newFakeExecutor().
		withProgram(ContainerEnginePodman.String(), ContainerEngineDocker.String(), ansible.PlaybookProgram, Program).
		withResponse(Program+" --version", Program+" 26.6.0", nil)

	run := NewRun(testHostDir, config, WithFs(memFs), WithExecutor(exec))
	run.SetEnv("RUN_VAR", "run-value")

	if err := run.Preflight(context.Background()); err != nil {
		t.Fatalf("preflight failed: %v", err)
	}

	if err := run.Setup(); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	if !reflect.DeepEqual(config, aliasProneConfig()) {
		t.Errorf("run mutated the caller config:\ngot:  %+v\nwant: %+v", config, testConfig(true))
	}
}
