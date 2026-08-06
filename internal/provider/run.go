package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible/navigator"
)

const (
	navigatorRunName                   = "terraform"
	navigatorRunExtraVarsFileName      = "terraform.yaml"
	navigatorRunPrevInventoryName      = "previous-terraform"
	navigatorRunDir                    = "tf-ansible-navigator-run"
	navigatorRunOperationEnvVar        = "ANSIBLE_TF_OPERATION"
	navigatorRunInventoryEnvVar        = "ANSIBLE_TF_INVENTORY"
	navigatorRunPrevInventoryEnvVar    = "ANSIBLE_TF_PREVIOUS_INVENTORY"
	navigatorRunTimeoutOverhead        = 5 * time.Second
	defaultNavigatorRunWorkingDir      = "."
	defaultNavigatorRunTimeout         = 10 * time.Minute
	defaultNavigatorRunContainerEngine = string(navigator.ContainerEngineAuto)
	defaultNavigatorRunEEEnabled       = true
	defaultNavigatorRunImage           = "ghcr.io/ansible/community-ansible-dev-tools:v26.7.1"
	defaultNavigatorRunPullPolicy      = string(navigator.PullPolicyTag)
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

type navigatorRunData struct {
	hostDir                 string
	config                  navigator.RunConfig
	operation               terraformOp
	persistDir              bool
	playbookArtifactQueries map[string]ansible.PlaybookArtifactQuery
	userArtifactQueries     bool
	knownHosts              []ansible.KnownHost
	command                 string
}

func (rd *navigatorRunData) Load(ctx context.Context, common NavigatorRunCommonModel) diag.Diagnostics {
	var diags diag.Diagnostics

	rd.config.WorkingDir = common.WorkingDirectory.ValueString()
	rd.config.Binary = common.AnsibleNavigatorBinary.ValueString()
	rd.config.Playbook = common.Playbook.ValueString()
	rd.config.Inventories = []ansible.Inventory{{Name: navigatorRunName, Contents: common.Inventory.ValueString()}}
	rd.config.Settings.Timezone = common.Timezone.ValueString()

	var eeModel ExecutionEnvironmentModel
	diags.Append(common.ExecutionEnvironment.As(ctx, &eeModel, basetypes.ObjectAsOptions{})...)

	diags.Append(eeModel.Value(ctx, &rd.config.Settings.ExecutionEnvironment)...)

	var optsModel AnsibleOptionsModel
	diags.Append(common.AnsibleOptions.As(ctx, &optsModel, basetypes.ObjectAsOptions{})...)

	diags.Append(optsModel.Value(ctx, &rd.config.Options)...)

	if !optsModel.ExtraVars.IsNull() {
		rd.config.ExtraVars = []ansible.ExtraVarsFile{{Name: navigatorRunExtraVarsFileName, Contents: optsModel.ExtraVars.ValueString()}}
	}

	var privateKeysModel []PrivateKeyModel
	if !optsModel.PrivateKeys.IsNull() {
		diags.Append(optsModel.PrivateKeys.ElementsAs(ctx, &privateKeysModel, false)...)
	}

	rd.config.PrivateKeys = make([]ansible.PrivateKey, 0, len(privateKeysModel))
	for _, model := range privateKeysModel {
		var key ansible.PrivateKey

		diags.Append(model.Value(ctx, &key)...)
		rd.config.PrivateKeys = append(rd.config.PrivateKeys, key)
	}

	var knownHosts []string
	if !optsModel.KnownHosts.IsUnknown() {
		diags.Append(optsModel.KnownHosts.ElementsAs(ctx, &knownHosts, false)...)
	}

	rd.config.KnownHosts = knownHosts

	rd.config.UseKnownHosts = optsModel.KnownHosts.IsUnknown() || len(optsModel.KnownHosts.Elements()) > 0

	rd.config.HostKeyChecking = optsModel.HostKeyChecking.ValueBool()
	if optsModel.HostKeyChecking.IsNull() {
		rd.config.HostKeyChecking = ansible.RunnerDefaultHostKeyChecking
	}

	return diags
}

func (rd navigatorRunData) Store(ctx context.Context, command *types.String, ansibleOpts *types.Object, artifactQueries *types.Map) diag.Diagnostics {
	var diags diag.Diagnostics

	*command = types.StringValue(rd.command)

	var optsModel AnsibleOptionsModel
	diags.Append(ansibleOpts.As(ctx, &optsModel, basetypes.ObjectAsOptions{})...)
	diags.Append(optsModel.Set(ctx, rd)...)

	optsResults, newDiags := types.ObjectValueFrom(ctx, AnsibleOptionsModel{}.AttrTypes(), optsModel)
	diags.Append(newDiags...)
	*ansibleOpts = optsResults

	var queriesModel map[string]ArtifactQueryModel
	diags.Append(artifactQueries.ElementsAs(ctx, &queriesModel, false)...)

	for name, model := range queriesModel {
		diags.Append(model.Set(ctx, rd.playbookArtifactQueries[name])...)
		queriesModel[name] = model
	}

	queriesValue, newDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: ArtifactQueryModel{}.AttrTypes()}, queriesModel)
	diags.Append(newDiags...)
	*artifactQueries = queriesValue

	return diags
}

func (rd navigatorRunData) artifactQueryPath(name string) path.Path {
	if !rd.userArtifactQueries {
		return path.Empty()
	}

	return path.Root("artifact_queries").AtMapKey(name)
}

func preflightCheckPath(check navigator.PreflightCheck) path.Path {
	switch check {
	case navigator.CheckWorkingDir:
		return path.Root("working_directory")
	case navigator.CheckContainerEngine:
		return path.Root("execution_environment").AtName("container_engine")
	case navigator.CheckPlaybook:
		return path.Root("execution_environment").AtName("enabled")
	case navigator.CheckNavigatorResolve, navigator.CheckNavigatorBinary:
		return path.Root("ansible_navigator_binary")
	}

	return path.Empty()
}

func setupStepPath(step navigator.SetupStep) path.Path {
	switch step {
	case navigator.SetupPlaybook:
		return path.Root("playbook")
	case navigator.SetupInventories:
		return path.Root("inventory")
	case navigator.SetupExtraVars:
		return path.Root("ansible_options").AtName("extra_vars")
	case navigator.SetupPrivateKeys:
		return path.Root("ansible_options").AtName("private_keys")
	case navigator.SetupKnownHosts:
		return path.Root("ansible_options").AtName("known_hosts")
	case navigator.SetupDir, navigator.SetupSettings:
		return path.Empty()
	}

	return path.Empty()
}

//nolint:cyclop
func run(ctx context.Context, diags *diag.Diagnostics, runData *navigatorRunData) {
	navRun := navigator.NewRun(runData.hostDir, runData.config)

	ctx = tflog.SetField(ctx, "operation", runData.operation.String())
	ctx = tflog.SetField(ctx, "mode", navRun.Mode().String())
	ctx = tflog.SetField(ctx, "workingDir", runData.config.WorkingDir)
	ctx = tflog.SetField(ctx, "hostDir", navRun.HostDir())
	ctx = tflog.SetField(ctx, "playbookDir", navRun.PlaybookDir())

	tflog.Debug(ctx, "starting run")

	defer func() {
		if !runData.persistDir {
			err := navRun.Cleanup()
			addWarning(diags, "Run not cleaned up", err)
		}
	}()

	navRun.SetEnv(navigatorRunOperationEnvVar, runData.operation.String())
	navRun.SetEnv(navigatorRunInventoryEnvVar, navRun.InventoryPath(navigatorRunName))

	if runData.operation == terraformOpUpdate {
		navRun.SetEnv(navigatorRunPrevInventoryEnvVar, navRun.InventoryPath(navigatorRunPrevInventoryName))
	}

	tflog.Trace(ctx, "running preflight checks")

	if err := navRun.Preflight(ctx); err != nil {
		for _, preflightErr := range unwrapJoinedErrors(err) {
			var typed *navigator.PreflightError
			if errors.As(preflightErr, &typed) {
				addPathError(diags, preflightCheckPath(typed.Check), "Preflight check failed", typed)

				continue
			}

			addError(diags, "Preflight check failed", preflightErr)
		}
	}

	tflog.Trace(ctx, "setting up run directory")

	if err := navRun.Setup(); err != nil {
		for _, setupErr := range unwrapJoinedErrors(err) {
			var typed *navigator.SetupError
			if errors.As(setupErr, &typed) {
				addPathError(diags, setupStepPath(typed.Step), "Setup failed", typed)

				continue
			}

			addError(diags, "Setup failed", setupErr)
		}
	}

	if diags.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("executing %s", navigator.Program))

	if err := navRun.Execute(ctx); err != nil {
		runData.command = navRun.Command.String()

		summary := "Ansible navigator run failed"
		if navRun.Status == ansible.StatusTimeout {
			summary = "Ansible navigator run timed out"
		}

		addError(diags, summary, fmt.Errorf("%w\n\nOutput:\n%s", err, navRun.Output))

		return
	}

	runData.command = navRun.Command.String()

	tflog.Trace(ctx, "querying playbook artifact")

	if err := navRun.Query(runData.playbookArtifactQueries); err != nil {
		for _, queryErr := range unwrapJoinedErrors(err) {
			var typed *navigator.QueryError
			if errors.As(queryErr, &typed) {
				addPathError(diags, runData.artifactQueryPath(typed.Name), "Playbook artifact query failed", typed)

				continue
			}

			addError(diags, "Playbook artifact queries failed", queryErr)
		}
	}

	if runData.config.UseKnownHosts {
		tflog.Trace(ctx, "reading known hosts")

		knownHosts, err := navRun.ReadKnownHosts()
		if err != nil {
			addPathError(diags, path.Root("ansible_options").AtName("known_hosts"), "Failed to read known hosts", err)
		}
		runData.knownHosts = knownHosts
	}

	tflog.Debug(ctx, "run complete")
}

func unwrapJoinedErrors(err error) []error {
	if err == nil {
		return nil
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		return joined.Unwrap()
	}

	return []error{err}
}

func navigatorRunDirPath(baseRunDirectory string, id string, runs uint32) string {
	return filepath.Join(baseRunDirectory, fmt.Sprintf("%s-%s-%d", navigatorRunDir, id, runs))
}
