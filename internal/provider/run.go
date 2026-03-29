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
	defaultNavigatorRunContainerEngine = navigator.ContainerEngineAuto
	defaultNavigatorRunEEEnabled       = true
	defaultNavigatorRunImage           = "ghcr.io/ansible/community-ansible-dev-tools:v26.1.0"
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

type navigatorRunData struct {
	hostDir                 string
	config                  navigator.RunConfig
	operation               terraformOp
	timeout                 time.Duration
	persistDir              bool
	playbookArtifactQueries map[string]ansible.PlaybookArtifactQuery
	knownHosts              []ansible.KnownHost
	command                 string
}

func (rd *navigatorRunData) Load(ctx context.Context, ee types.Object, timezone string, ansibleOpts types.Object) diag.Diagnostics {
	var diags diag.Diagnostics

	var eeModel ExecutionEnvironmentModel
	diags.Append(ee.As(ctx, &eeModel, basetypes.ObjectAsOptions{})...)

	rd.config.Settings.Timezone = timezone

	diags.Append(eeModel.Value(ctx, rd.config.Settings)...)

	var optsModel AnsibleOptionsModel
	diags.Append(ansibleOpts.As(ctx, &optsModel, basetypes.ObjectAsOptions{})...)

	diags.Append(optsModel.Value(ctx, rd.config.Options)...)

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

func preflightCheckPath(check navigator.PreflightCheckID) path.Path {
	switch check {
	case navigator.CheckWorkingDir:
		return path.Root("working_directory")
	case navigator.CheckContainerEngine:
		return path.Root("execution_environment").AtMapKey("container_engine")
	case navigator.CheckPlaybook:
		return path.Root("execution_environment").AtMapKey("enabled")
	case navigator.CheckNavigatorResolve, navigator.CheckNavigatorBinary:
		return path.Root("ansible_navigator_binary")
	default:
		return path.Empty()
	}
}

//nolint:cyclop
func run(ctx context.Context, diags *diag.Diagnostics, runData *navigatorRunData) {
	tflog.Debug(ctx, "starting run")

	navRun := navigator.NewRun(runData.hostDir, &runData.config)
	defer func() {
		if !runData.persistDir {
			err := navRun.Cleanup()
			addWarning(diags, "Run not cleaned up", err)
		}
	}()

	tflog.Trace(ctx, "preflight checks")
	ctx = tflog.SetField(ctx, "workingDir", runData.config.WorkingDir)
	if err := navRun.Preflight(ctx); err != nil {
		for _, e := range unwrapJoinedErrors(err) {
			var pe *navigator.PreflightError
			if errors.As(e, &pe) {
				addPathError(diags, preflightCheckPath(pe.Check), "Preflight check failed", pe)
			}
		}
	}

	tflog.Trace(ctx, "creating directories and files")
	if err := navRun.Setup(); err != nil {
		for _, e := range unwrapJoinedErrors(err) {
			var se *navigator.SetupError
			if errors.As(e, &se) {
				addError(diags, "Setup failed", se)
			}
		}
	}

	ctx = tflog.SetField(ctx, "hostRunDir", navRun.HostDir())
	ctx = tflog.SetField(ctx, "resolvedRunDir", navRun.ResolvedDir())

	if diags.HasError() {
		return
	}

	runData.config.Env[navigatorRunOperationEnvVar] = runData.operation.String()
	runData.config.Env[navigatorRunInventoryEnvVar] = navRun.InventoryPath(navigatorRunName)
	if runData.operation == terraformOpUpdate {
		runData.config.Env[navigatorRunPrevInventoryEnvVar] = navRun.InventoryPath(navigatorRunPrevInventoryName)
	}

	runData.config.Settings.Timeout = runData.timeout

	tflog.Trace(ctx, "executing ansible-navigator")
	if err := navRun.Execute(ctx); err != nil {
		runData.command = navRun.Command
		switch navRun.Status {
		case "timeout":
			addError(diags, "Ansible navigator run timed out", fmt.Errorf("%w\n\nOutput:\n%s", err, navRun.Output))
		default:
			addError(diags, "Ansible navigator run failed", fmt.Errorf("%w\n\nOutput:\n%s", err, navRun.Output))
		}

		return
	}

	runData.command = navRun.Command

	if err := navRun.Query(runData.playbookArtifactQueries); err != nil {
		addPathError(diags, path.Root("artifact_queries"), "Playbook artifact queries failed", err)
	}

	if runData.config.UseKnownHosts {
		knownHosts, err := navRun.ReadKnownHosts()
		if err != nil {
			addPathError(diags, path.Root("ansible_options").AtMapKey("known_hosts"), "Failed to read known hosts", err)
		}
		runData.knownHosts = knownHosts
	}
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
