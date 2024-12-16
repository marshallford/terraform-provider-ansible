package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

type ExecutionEnvironmentModel struct {
	ContainerEngine          types.String `tfsdk:"container_engine"`
	Enabled                  types.Bool   `tfsdk:"enabled"`
	EnvironmentVariablesPass types.List   `tfsdk:"environment_variables_pass"`
	EnvironmentVariablesSet  types.Map    `tfsdk:"environment_variables_set"`
	Image                    types.String `tfsdk:"image"`
	PullArguments            types.List   `tfsdk:"pull_arguments"`
	PullPolicy               types.String `tfsdk:"pull_policy"`
	ContainerOptions         types.List   `tfsdk:"container_options"`
}

type AnsibleOptionsModel struct {
	ForceHandlers   types.Bool   `tfsdk:"force_handlers"`
	SkipTags        types.List   `tfsdk:"skip_tags"`
	StartAtTask     types.String `tfsdk:"start_at_task"`
	Limit           types.List   `tfsdk:"limit"`
	Tags            types.List   `tfsdk:"tags"`
	PrivateKeys     types.List   `tfsdk:"private_keys"`
	KnownHosts      types.List   `tfsdk:"known_hosts"`
	HostKeyChecking types.Bool   `tfsdk:"host_key_checking"`
}

type PrivateKeyModel struct {
	Name types.String `tfsdk:"name"`
	Data types.String `tfsdk:"data"`
}

type ArtifactQueryModel struct {
	JQFilter types.String `tfsdk:"jq_filter"`
	Results  types.List   `tfsdk:"results"`
}

func (ExecutionEnvironmentModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"container_engine":           types.StringType,
		"enabled":                    types.BoolType,
		"environment_variables_pass": types.ListType{ElemType: types.StringType},
		"environment_variables_set":  types.MapType{ElemType: types.StringType},
		"image":                      types.StringType,
		"pull_arguments":             types.ListType{ElemType: types.StringType},
		"pull_policy":                types.StringType,
		"container_options":          types.ListType{ElemType: types.StringType},
	}
}

func (ExecutionEnvironmentModel) Defaults() basetypes.ObjectValue {
	return types.ObjectValueMust(
		ExecutionEnvironmentModel{}.AttrTypes(),
		map[string]attr.Value{
			"container_engine":           types.StringValue(defaultNavigatorRunContainerEngine),
			"enabled":                    types.BoolValue(defaultNavigatorRunEEEnabled),
			"environment_variables_pass": types.ListNull(types.StringType),
			"environment_variables_set":  types.MapNull(types.StringType),
			"image":                      types.StringValue(defaultNavigatorRunImage),
			"pull_arguments":             types.ListNull(types.StringType),
			"pull_policy":                types.StringValue(defaultNavigatorRunPullPolicy),
			"container_options":          types.ListNull(types.StringType),
		},
	)
}

func (m ExecutionEnvironmentModel) Value(ctx context.Context, settings *ansible.NavigatorSettings) diag.Diagnostics {
	var diags diag.Diagnostics

	settings.ContainerEngine = m.ContainerEngine.ValueString()

	settings.EEEnabled = m.Enabled.ValueBool()

	var envVarsPass []string
	if !m.EnvironmentVariablesPass.IsNull() {
		diags.Append(m.EnvironmentVariablesPass.ElementsAs(ctx, &envVarsPass, false)...)
	}

	settings.EnvironmentVariablesPass = envVarsPass

	envVarsSet := map[string]string{}
	if !m.EnvironmentVariablesSet.IsNull() {
		diags.Append(m.EnvironmentVariablesSet.ElementsAs(ctx, &envVarsSet, false)...)
	}

	settings.EnvironmentVariablesSet = envVarsSet

	settings.Image = m.Image.ValueString()

	var pullArguments []string
	if !m.PullArguments.IsNull() {
		diags.Append(m.PullArguments.ElementsAs(ctx, &pullArguments, false)...)
	}
	settings.PullArguments = pullArguments

	settings.PullPolicy = m.PullPolicy.ValueString()

	var containerOptions []string
	if !m.ContainerOptions.IsNull() {
		diags.Append(m.ContainerOptions.ElementsAs(ctx, &containerOptions, false)...)
	}
	settings.ContainerOptions = containerOptions

	return diags
}

func (PrivateKeyModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name": types.StringType,
		"data": types.StringType,
	}
}

func (AnsibleOptionsModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"force_handlers":    types.BoolType,
		"skip_tags":         types.ListType{ElemType: types.StringType},
		"start_at_task":     types.StringType,
		"limit":             types.ListType{ElemType: types.StringType},
		"tags":              types.ListType{ElemType: types.StringType},
		"private_keys":      types.ListType{ElemType: types.ObjectType{AttrTypes: PrivateKeyModel{}.AttrTypes()}},
		"known_hosts":       types.ListType{ElemType: types.StringType},
		"host_key_checking": types.BoolType,
	}
}

func (AnsibleOptionsModel) Defaults() basetypes.ObjectValue {
	return types.ObjectValueMust(
		AnsibleOptionsModel{}.AttrTypes(),
		map[string]attr.Value{
			"force_handlers":    types.BoolNull(),
			"skip_tags":         types.ListNull(types.StringType),
			"start_at_task":     types.StringNull(),
			"limit":             types.ListNull(types.StringType),
			"tags":              types.ListNull(types.StringType),
			"private_keys":      types.ListNull(types.ObjectType{AttrTypes: PrivateKeyModel{}.AttrTypes()}),
			"known_hosts":       types.ListUnknown(types.StringType),
			"host_key_checking": types.BoolNull(),
		},
	)
}

func (m AnsibleOptionsModel) Value(ctx context.Context, options *ansible.Options) diag.Diagnostics {
	var diags diag.Diagnostics

	options.ForceHandlers = m.ForceHandlers.ValueBool()

	var skipTags []string
	if !m.SkipTags.IsNull() {
		diags.Append(m.SkipTags.ElementsAs(ctx, &skipTags, false)...)
	}
	options.SkipTags = skipTags

	options.StartAtTask = m.StartAtTask.ValueString()

	var limit []string
	if !m.Limit.IsNull() {
		diags.Append(m.Limit.ElementsAs(ctx, &limit, false)...)
	}
	options.Limit = limit

	var tags []string
	if !m.Tags.IsNull() {
		diags.Append(m.Tags.ElementsAs(ctx, &tags, false)...)
	}
	options.Tags = tags

	var privateKeysModel []PrivateKeyModel
	if !m.PrivateKeys.IsNull() {
		diags.Append(m.PrivateKeys.ElementsAs(ctx, &privateKeysModel, false)...)
	}

	privateKeys := make([]string, 0, len(privateKeysModel))
	for _, privateKeyModel := range privateKeysModel {
		privateKeys = append(privateKeys, privateKeyModel.Name.ValueString())
	}
	options.PrivateKeys = privateKeys

	options.KnownHosts = m.KnownHosts.IsUnknown() || len(m.KnownHosts.Elements()) > 0

	options.HostKeyChecking = m.HostKeyChecking.ValueBool()
	if m.HostKeyChecking.IsNull() {
		options.HostKeyChecking = ansible.RunnerDefaultHostKeyChecking
	}

	return diags
}

func (m *AnsibleOptionsModel) Set(ctx context.Context, run *navigatorRun) diag.Diagnostics {
	var diags diag.Diagnostics

	if m.KnownHosts.IsUnknown() {
		knownHostsValue, newDiags := types.ListValueFrom(ctx, types.StringType, run.knownHosts)
		diags.Append(newDiags...)
		m.KnownHosts = knownHostsValue
	}

	return diags
}

func (m PrivateKeyModel) Value(ctx context.Context, key *ansible.PrivateKey) diag.Diagnostics {
	var diags diag.Diagnostics

	key.Name = m.Name.ValueString()
	key.Data = m.Data.ValueString()

	return diags
}

func (ArtifactQueryModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"jq_filter": types.StringType,
		"results":   types.ListType{ElemType: types.StringType},
	}
}

func (m ArtifactQueryModel) Value(ctx context.Context, query *ansible.ArtifactQuery) diag.Diagnostics {
	var diags diag.Diagnostics

	query.JQFilter = m.JQFilter.ValueString()
	query.Results = []string{} // m.Results always unknown when this function is called

	return diags
}

func (m *ArtifactQueryModel) Set(ctx context.Context, query ansible.ArtifactQuery) diag.Diagnostics {
	var diags diag.Diagnostics

	m.JQFilter = types.StringValue(query.JQFilter)

	resultsValue, newDiags := types.ListValueFrom(ctx, types.StringType, query.Results)
	diags.Append(newDiags...)
	m.Results = resultsValue

	return diags
}
