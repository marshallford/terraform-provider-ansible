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
	defaultNavigatorRunImage           = "ghcr.io/ansible/community-ansible-dev-tools:v25.8.3"
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
	dir               string
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

func run(ctx context.Context, diags *diag.Diagnostics, timeout time.Duration, operation terraformOp, run *navigatorRun) { //nolint:cyclop
	var err error

	ctx = tflog.SetField(ctx, "dir", run.dir)
	ctx = tflog.SetField(ctx, "workingDir", run.workingDir)
	tflog.Debug(ctx, "starting run")

	tflog.Trace(ctx, "directory preflight")
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

	err = ansible.CreateRunDir(run.dir)
	addError(diags, "Run directory not created", err)

	err = ansible.CreatePlaybook(run.dir, run.playbook)
	addError(diags, "Ansible playbook not created", err)

	err = ansible.CreateInventories(run.dir, run.inventories, &run.navigatorSettings)
	addError(diags, "Ansible inventories not created", err)

	if len(run.privateKeys) > 0 {
		err = ansible.CreatePrivateKeys(run.dir, run.privateKeys, &run.navigatorSettings)
		addError(diags, "Private keys not created", err)
	}

	if run.options.KnownHosts {
		err = ansible.CreateKnownHosts(run.dir, run.knownHosts, &run.navigatorSettings)
		addError(diags, "Known hosts not created", err)
	}

	run.navigatorSettings.EnvironmentVariablesSet[navigatorRunOperationEnvVar] = operation.String()
	run.navigatorSettings.EnvironmentVariablesSet[navigatorRunInventoryEnvVar] = ansible.InventoryPath(
		run.dir,
		navigatorRunName,
		run.navigatorSettings.EEEnabled,
		false,
	)
	if operation == terraformOpUpdate {
		run.navigatorSettings.EnvironmentVariablesSet[navigatorRunPrevInventoryEnvVar] = ansible.InventoryPath(
			run.dir,
			navigatorRunPrevInventoryName,
			run.navigatorSettings.EEEnabled,
			true,
		)
	}

	run.navigatorSettings.Timeout = timeout

	navigatorSettingsContents, err := ansible.GenerateNavigatorSettings(&run.navigatorSettings)
	addError(diags, "Ansible navigator settings not generated", err)

	err = ansible.CreateNavigatorSettingsFile(run.dir, navigatorSettingsContents)
	addError(diags, "Ansible navigator settings file not created", err)

	if diags.HasError() {
		if !run.persistDir {
			err = ansible.RemoveRunDir(run.dir)
			addWarning(diags, "Run directory not removed", err)
		}

		return
	}

	command := ansible.GenerateNavigatorRunCommand(
		ctx,
		run.dir,
		run.workingDir,
		binary,
		run.navigatorSettings.EEEnabled,
		&run.options,
	)

	run.command = command.String()

	commandOutput, err := ansible.ExecNavigatorRunCommand(command)
	if err != nil {
		output, _ := ansible.GetStdoutFromPlaybookArtifact(run.dir)
		if output == "" {
			output = commandOutput
		}

		status, _ := ansible.GetStatusFromPlaybookArtifact(run.dir)
		switch status {
		case "timeout":
			addError(diags, "Ansible navigator run timed out", fmt.Errorf("%w\n\nOutput:\n%s", err, output))
		default:
			addError(diags, "Ansible navigator run failed", fmt.Errorf("%w\n\nOutput:\n%s", err, output))
		}
	}

	if !diags.HasError() {
		err = ansible.QueryPlaybookArtifact(run.dir, run.artifactQueries)
		addPathError(diags, path.Root("artifact_queries"), "Playbook artifact queries failed", err)

		if run.options.KnownHosts {
			knownHosts, err := ansible.GetKnownHosts(run.dir)
			addPathError(diags, path.Root("ansible_options").AtMapKey("known_hosts"), "Failed to get known hosts", err)
			run.knownHosts = knownHosts
		}
	}

	if !run.persistDir {
		err = ansible.RemoveRunDir(run.dir)
		addWarning(diags, "Run directory not removed", err)
	}
}

func runDir(baseRunDirectory string, id string, runs uint32) string {
	return filepath.Join(baseRunDirectory, fmt.Sprintf("%s-%s-%d", navigatorRunDir, id, runs))
}
