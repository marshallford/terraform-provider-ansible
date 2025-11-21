package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

const (
	navigatorRunName                   = "terraform"
	navigatorRunPrevInventoryName      = "previous-terraform"
	navigatorRunDir                    = "tf-ansible-navigator-run"
	navigatorRunOperationEnvVar        = "ANSIBLE_TF_OPERATION"
	navigatorRunInventoryEnvVar        = "ANSIBLE_TF_INVENTORY"
	navigatorRunPrevInventoryEnvVar    = "ANSIBLE_TF_PREVIOUS_INVENTORY"
	navigatorRunTimeoutOverhead        = 5 * time.Second
	defaultNavigatorRunWorkingDir      = "."
	defaultNavigatorRunTimeout         = 10 * time.Minute
	defaultNavigatorRunContainerEngine = ansible.ContainerEngineAuto
	defaultNavigatorRunEEEnabled       = true
	defaultNavigatorRunImage           = "ghcr.io/ansible/community-ansible-dev-tools:v25.10.0"
	defaultNavigatorRunPullPolicy      = "tag"
	defaultNavigatorRunTimezone        = "UTC"
	defaultNavigatorRunOnDestroy       = false
)

type (
	getKey func(ctx context.Context, key string) ([]byte, diag.Diagnostics)
	setKey func(ctx context.Context, key string, value []byte) diag.Diagnostics
)

func setRuns(ctx context.Context, diags *diag.Diagnostics, setKey setKey, runs uint32) {
	runsBytes, err := json.Marshal(runs)
	if addError(diags, "Failed to set 'runs' private state", err) {
		return
	}

	setKey(ctx, "runs", runsBytes)
}

func incrementRuns(ctx context.Context, diags *diag.Diagnostics, getKey getKey, setKey setKey) uint32 {
	runsBytes, newDiags := getKey(ctx, "runs")
	diags.Append(newDiags...)

	runs := uint32(0)
	if runsBytes != nil {
		err := json.Unmarshal(runsBytes, &runs)
		if addError(diags, "Failed to get 'runs' private state", err) {
			return runs
		}
	}

	runs++

	runsBytes, err := json.Marshal(runs)
	if addError(diags, "Failed to set 'runs' private state", err) {
		return runs
	}

	setKey(ctx, "runs", runsBytes)

	return runs
}

type navigatorRun struct {
	hostDir           string
	persistDir        bool
	playbook          string
	inventories       []ansible.Inventory
	workingDir        string
	navigatorBinary   string
	options           ansible.Options
	navigatorSettings ansible.NavigatorSettings
	privateKeys       []ansible.PrivateKey
	knownHosts        []ansible.KnownHost
	artifactQueries   map[string]ansible.ArtifactQuery
	command           string
}

//nolint:cyclop
func run(ctx context.Context, diags *diag.Diagnostics, timeout time.Duration, operation terraformOp, run *navigatorRun) {
	var err error

	tflog.Debug(ctx, "starting run")

	tflog.Trace(ctx, "directory preflight")
	ctx = tflog.SetField(ctx, "workingDir", run.workingDir)
	err = ansible.DirectoryPreflight(run.workingDir)
	addPathError(diags, path.Root("working_directory"), "Working directory preflight check", err)

	if run.navigatorSettings.EEEnabled {
		tflog.Trace(ctx, "container engine preflight")
		err = ansible.ContainerEnginePreflight(ctx, run.navigatorSettings.ContainerEngine)
		addPathError(diags, path.Root("execution_environment").AtMapKey("container_engine"), "Container engine preflight check", err)
	} else {
		tflog.Trace(ctx, "playbook preflight")
		err = ansible.PlaybookPreflight(ctx)
		addPathError(diags, path.Root("execution_environment").AtMapKey("enabled"), "Ansible playbook preflight check", err)
	}

	tflog.Trace(ctx, "navigator path preflight")
	binary, err := ansible.NavigatorPathPreflight(run.navigatorBinary)
	addPathError(diags, path.Root("ansible_navigator_binary"), "Ansible navigator not found", err)

	tflog.Trace(ctx, "navigator preflight")
	err = ansible.NavigatorPreflight(ctx, binary)
	addPathError(diags, path.Root("ansible_navigator_binary"), "Ansible navigator preflight check", err)

	tflog.Trace(ctx, "creating directories and files")

	runDir, err := ansible.CreateRunDir(run.hostDir, &run.navigatorSettings)
	ctx = tflog.SetField(ctx, "hostRunDir", runDir.Host)
	ctx = tflog.SetField(ctx, "resolvedRunDir", runDir.Resolved)
	addError(diags, "Run directory not created", err)

	err = ansible.CreatePlaybook(runDir, run.playbook)
	addError(diags, "Ansible playbook not created", err)

	err = ansible.CreateInventories(runDir, run.inventories)
	addError(diags, "Ansible inventories not created", err)

	if len(run.privateKeys) > 0 {
		err = ansible.CreatePrivateKeys(runDir, run.privateKeys)
		addError(diags, "Private keys not created", err)
	}

	if run.options.KnownHosts {
		err = ansible.CreateKnownHosts(runDir, run.knownHosts)
		addError(diags, "Known hosts not created", err)
	}

	inventoryPaths := ansible.ResolvedInventoryPaths(runDir, run.inventories)

	run.navigatorSettings.EnvironmentVariablesSet[navigatorRunOperationEnvVar] = operation.String()
	run.navigatorSettings.EnvironmentVariablesSet[navigatorRunInventoryEnvVar] = inventoryPaths[navigatorRunName]
	if operation == terraformOpUpdate {
		run.navigatorSettings.EnvironmentVariablesSet[navigatorRunPrevInventoryEnvVar] = inventoryPaths[navigatorRunPrevInventoryName]
	}

	run.navigatorSettings.Timeout = timeout

	navigatorSettingsContents, err := ansible.GenerateNavigatorSettings(&run.navigatorSettings)
	addError(diags, "Ansible navigator settings not generated", err)

	err = ansible.CreateNavigatorSettingsFile(runDir, navigatorSettingsContents)
	addError(diags, "Ansible navigator settings file not created", err)

	if diags.HasError() {
		if !run.persistDir {
			err = runDir.Remove()
			addWarning(diags, "Run directory not removed", err)
		}

		return
	}

	command := ansible.GenerateNavigatorRunCommand(
		ctx,
		runDir,
		run.workingDir,
		binary,
		&run.options,
	)

	run.command = command.String()

	commandOutput, err := ansible.ExecNavigatorRunCommand(command)
	if err != nil {
		output, _ := ansible.GetStdoutFromPlaybookArtifact(runDir)
		if output == "" {
			output = commandOutput
		}

		status, _ := ansible.GetStatusFromPlaybookArtifact(runDir)
		switch status {
		case "timeout":
			addError(diags, "Ansible navigator run timed out", fmt.Errorf("%w\n\nOutput:\n%s", err, output))
		default:
			addError(diags, "Ansible navigator run failed", fmt.Errorf("%w\n\nOutput:\n%s", err, output))
		}
	}

	if !diags.HasError() {
		err = ansible.QueryPlaybookArtifact(runDir, run.artifactQueries)
		addPathError(diags, path.Root("artifact_queries"), "Playbook artifact queries failed", err)

		if run.options.KnownHosts {
			knownHosts, err := ansible.GetKnownHosts(runDir)
			addPathError(diags, path.Root("ansible_options").AtMapKey("known_hosts"), "Failed to get known hosts", err)
			run.knownHosts = knownHosts
		}
	}

	if !run.persistDir {
		err = runDir.Remove()
		addWarning(diags, "Run directory not removed", err)
	}
}

func runDir(baseRunDirectory string, id string, runs uint32) string {
	return filepath.Join(baseRunDirectory, fmt.Sprintf("%s-%s-%d", navigatorRunDir, id, runs))
}
