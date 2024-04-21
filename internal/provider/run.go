package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

type (
	GetKey func(ctx context.Context, key string) ([]byte, diag.Diagnostics)
	SetKey func(ctx context.Context, key string, value []byte) diag.Diagnostics
)

func SetRuns(ctx context.Context, diags *diag.Diagnostics, setKey SetKey, runs uint32) {
	runsBytes, err := json.Marshal(runs)
	if addError(diags, "Failed to set 'runs' private state", err) {
		return
	}

	setKey(ctx, "runs", runsBytes)
}

func IncrementRuns(ctx context.Context, diags *diag.Diagnostics, getKey GetKey, setKey SetKey) uint32 {
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

func (r *NavigatorRunResource) Run(ctx context.Context, diags *diag.Diagnostics, data *NavigatorRunResourceModel, runs uint32, operation TerraformOperation) { //nolint:cyclop
	var err error

	var timeout time.Duration
	var newDiags diag.Diagnostics

	switch operation {
	case terraformOperationCreate:
		timeout, newDiags = data.Timeouts.Create(ctx, defaultNavigatorRunTimeout)
	case terraformOperationUpdate:
		timeout, newDiags = data.Timeouts.Update(ctx, defaultNavigatorRunTimeout)
	case terraformOperationDelete:
		timeout, newDiags = data.Timeouts.Delete(ctx, defaultNavigatorRunTimeout)
	}

	diags.Append(newDiags...)
	if diags.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)

	defer cancel()

	runDir := filepath.Join(r.opts.BaseRunDirectory, fmt.Sprintf("%s-%s-%d", navigatorRunDir, data.ID.ValueString(), runs))

	var eeModel ExecutionEnvironmentModel
	diags.Append(data.ExecutionEnvironment.As(ctx, &eeModel, basetypes.ObjectAsOptions{})...)

	var navigatorSettings ansible.NavigatorSettings
	navigatorSettings.Timezone = data.Timezone.ValueString()
	diags.Append(eeModel.Value(ctx, &navigatorSettings)...)

	var optsModel AnsibleOptionsModel
	diags.Append(data.AnsibleOptions.As(ctx, &optsModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)

	var ansibleOptions ansible.RunOptions
	diags.Append(optsModel.Value(ctx, &ansibleOptions)...)

	var privateKeysModel []PrivateKeyModel
	if !optsModel.PrivateKeys.IsNull() {
		diags.Append(optsModel.PrivateKeys.ElementsAs(ctx, &privateKeysModel, false)...)
	}

	privateKeys := make([]ansible.PrivateKey, 0, len(privateKeysModel))
	for _, model := range privateKeysModel {
		var key ansible.PrivateKey

		diags.Append(model.Value(ctx, &key)...)
		privateKeys = append(privateKeys, key)
	}

	var queriesModel map[string]ArtifactQueryModel
	diags.Append(data.ArtifactQueries.ElementsAs(ctx, &queriesModel, false)...)

	artifactQueries := map[string]ansible.ArtifactQuery{}
	for name, model := range queriesModel {
		var query ansible.ArtifactQuery

		diags.Append(model.Value(ctx, &query)...)
		artifactQueries[name] = query
	}

	if diags.HasError() {
		return
	}

	err = ansible.DirectoryPreflight(data.WorkingDirectory.ValueString())
	addPathError(diags, path.Root("working_directory"), "Working directory preflight check", err)

	err = ansible.ContainerEnginePreflight(navigatorSettings.ContainerEngine)
	addPathError(diags, path.Root("execution_environment").AtMapKey("container_engine"), "Container engine preflight check", err)

	ansibleNavigatorBinary, err := ansible.NavigatorPath(data.AnsibleNavigatorBinary.ValueString())
	addError(diags, "Ansible navigator not found", err)

	err = ansible.NavigatorPreflight(ansibleNavigatorBinary)
	addPathError(diags, path.Root("ansible_navigator_binary"), "Ansible navigator preflight check", err)

	err = ansible.CreateRunDir(runDir)
	addError(diags, "Run directory not created", err)

	err = ansible.CreatePlaybookFile(runDir, data.Playbook.ValueString())
	addError(diags, "Ansible playbook file not created", err)

	err = ansible.CreateInventoryFile(runDir, data.Inventory.ValueString())
	addError(diags, "Ansible inventory file not created", err)

	err = ansible.CreatePrivateKeys(runDir, privateKeys, &navigatorSettings)
	addError(diags, "Private keys not created", err)

	navigatorSettings.EnvironmentVariablesSet[navigatorRunOperationEnvVar] = operation.String()
	navigatorSettings.Timeout = timeout

	navigatorSettingsContents, err := ansible.GenerateNavigatorSettings(&navigatorSettings)
	addError(diags, "Ansible navigator settings not generated", err)

	err = ansible.CreateNavigatorSettingsFile(runDir, navigatorSettingsContents)
	addError(diags, "Ansible navigator settings file not created", err)

	command := ansible.GenerateNavigatorRunCommand(
		data.WorkingDirectory.ValueString(),
		ansibleNavigatorBinary,
		runDir,
		&ansibleOptions,
	)

	if diags.HasError() {
		if !r.opts.PersistRunDirectory {
			err = ansible.RemoveRunDir(runDir)
			addWarning(diags, "Run directory not removed", err)
		}

		return
	}

	data.Command = types.StringValue(command.String())

	commandOutput, err := ansible.ExecNavigatorRunCommand(command)
	if err != nil {
		output, _ := ansible.GetStdoutFromPlaybookArtifact(runDir)
		if output == "" {
			output = commandOutput
		}

		status, _ := ansible.GetStatusFromPlaybookArtifact(runDir)
		switch status {
		case "timeout":
			addError(diags, "Ansible navigator run timed out", fmt.Errorf("%w\n\noutput:\n%s", err, output))
		default:
			addError(diags, "Ansible navigator run failed", fmt.Errorf("%w\n\noutput:\n%s", err, output))
		}
	}

	// TODO skip on destroy?
	err = ansible.QueryPlaybookArtifact(runDir, artifactQueries)
	addPathError(diags, path.Root("artifact_queries"), "Playbook artifact queries failed ", err)

	for name, model := range queriesModel {
		diags.Append(model.Set(ctx, artifactQueries[name])...)
		queriesModel[name] = model
	}

	newQueriesModel, newDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: ArtifactQueryModel{}.AttrTypes()}, queriesModel)
	diags.Append(newDiags...)
	data.ArtifactQueries = newQueriesModel

	if !r.opts.PersistRunDirectory {
		err = ansible.RemoveRunDir(runDir)
		addWarning(diags, "Run directory not removed", err)
	}
}

func (*NavigatorRunResource) ShouldRun(plan *NavigatorRunResourceModel, state *NavigatorRunResourceModel) bool {
	attributeChanges := []bool{
		plan.Playbook.Equal(state.Playbook),
		plan.Inventory.Equal(state.Inventory),
		plan.WorkingDirectory.Equal(state.WorkingDirectory),
		plan.ExecutionEnvironment.Equal(state.ExecutionEnvironment),
		plan.AnsibleOptions.Equal(state.AnsibleOptions), // TODO check nested attrs
		plan.Triggers.Equal(state.Triggers),
		plan.ArtifactQueries.Equal(state.ArtifactQueries),
	}

	for _, attributeChange := range attributeChanges {
		if !attributeChange {
			return true
		}
	}

	return false
}
