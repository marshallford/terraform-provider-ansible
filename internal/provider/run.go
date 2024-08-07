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
	navigatorRunDir                    = "tf-ansible-navigator-run"
	navigatorRunOperationEnvVar        = "ANSIBLE_TF_OPERATION"
	defaultNavigatorRunWorkingDir      = "."
	defaultNavigatorRunTimeout         = 10 * time.Minute
	defaultNavigatorRunContainerEngine = ansible.ContainerEngineAuto
	defaultNavigatorRunEEEnabled       = true
	defaultNavigatorRunImage           = "ghcr.io/ansible/community-ansible-dev-tools:v24.7.2"
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
	inventory         string
	workingDir        string
	navigatorBinary   string
	options           ansible.Options
	navigatorSettings ansible.NavigatorSettings
	privateKeys       []ansible.PrivateKey
	artifactQueries   map[string]ansible.ArtifactQuery
	command           string
}

func run(ctx context.Context, diags *diag.Diagnostics, timeout time.Duration, operation terraformOperation, run *navigatorRun) {
	var err error

	ctx = tflog.SetField(ctx, "dir", run.dir)
	ctx = tflog.SetField(ctx, "workingDir", run.workingDir)
	tflog.Debug(ctx, "starting run")

	tflog.Trace(ctx, "directory preflight")
	err = ansible.DirectoryPreflight(run.workingDir)
	addPathError(diags, path.Root("working_directory"), "Working directory preflight check", err)

	if run.navigatorSettings.EEEnabled {
		tflog.Trace(ctx, "container engine preflight")
		err = ansible.ContainerEnginePreflight(run.navigatorSettings.ContainerEngine)
		addPathError(diags, path.Root("execution_environment").AtMapKey("container_engine"), "Container engine preflight check", err)
	} else {
		tflog.Trace(ctx, "playbook preflight")
		err = ansible.PlaybookPreflight()
		addPathError(diags, path.Root("execution_environment").AtMapKey("enabled"), "Ansible playbook preflight check", err)
	}

	tflog.Trace(ctx, "navigator path preflight")
	binary, err := ansible.NavigatorPathPreflight(run.navigatorBinary)
	addPathError(diags, path.Root("ansible_navigator_binary"), "Ansible navigator not found", err)

	tflog.Trace(ctx, "navigator preflight")
	err = ansible.NavigatorPreflight(binary)
	addPathError(diags, path.Root("ansible_navigator_binary"), "Ansible navigator preflight check", err)

	tflog.Trace(ctx, "creating directories and files")

	err = ansible.CreateRunDir(run.dir)
	addError(diags, "Run directory not created", err)

	err = ansible.CreatePlaybookFile(run.dir, run.playbook)
	addError(diags, "Ansible playbook file not created", err)

	err = ansible.CreateInventoryFile(run.dir, run.inventory)
	addError(diags, "Ansible inventory file not created", err)

	err = ansible.CreatePrivateKeys(run.dir, run.privateKeys, &run.navigatorSettings)
	addError(diags, "Private keys not created", err)

	run.navigatorSettings.EnvironmentVariablesSet[navigatorRunOperationEnvVar] = operation.String()
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
			addError(diags, "Ansible navigator run timed out", fmt.Errorf("%w\n\noutput:\n%s", err, output))
		default:
			addError(diags, "Ansible navigator run failed", fmt.Errorf("%w\n\noutput:\n%s", err, output))
		}
	}

	err = ansible.QueryPlaybookArtifact(run.dir, run.artifactQueries)
	addPathError(diags, path.Root("artifact_queries"), "Playbook artifact queries failed ", err)

	if !run.persistDir {
		err = ansible.RemoveRunDir(run.dir)
		addWarning(diags, "Run directory not removed", err)
	}
}

func runDir(baseRunDirectory string, id string, runs uint32) string {
	return filepath.Join(baseRunDirectory, fmt.Sprintf("%s-%s-%d", navigatorRunDir, id, runs))
}
