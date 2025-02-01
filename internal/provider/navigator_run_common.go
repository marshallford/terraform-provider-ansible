package provider

import (
	"context"
	"fmt"

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

func navigatorRunDescriptions() map[string]attrDescription {
	return map[string]attrDescription{
		"playbook": {
			Description:         "Ansible playbook contents.",
			MarkdownDescription: "Ansible [playbook](https://docs.ansible.com/ansible/latest/playbook_guide/playbooks_intro.html) contents.",
		},
		"inventory": {
			Description:         "Ansible inventory contents.",
			MarkdownDescription: "Ansible [inventory](https://docs.ansible.com/ansible/latest/getting_started/get_started_inventory.html) contents.",
		},
		"working_directory": {
			Description:         fmt.Sprintf("Directory which '%s' is run from. Recommended to be the root Ansible content directory (sometimes called the project directory), which is likely to contain 'ansible.cfg', 'roles/', etc. Defaults to '%s'.", ansible.NavigatorProgram, defaultNavigatorRunWorkingDir),
			MarkdownDescription: fmt.Sprintf("Directory which `%s` is run from. Recommended to be the root Ansible [content directory](https://docs.ansible.com/ansible/latest/tips_tricks/sample_setup.html#sample-directory-layout) (sometimes called the project directory), which is likely to contain `ansible.cfg`, `roles/`, etc. Defaults to `%s`.", ansible.NavigatorProgram, defaultNavigatorRunWorkingDir),
		},
		"execution_environment": {
			Description:         "Execution environment (EE) related configuration.",
			MarkdownDescription: "[Execution environment](https://ansible.readthedocs.io/en/latest/getting_started_ee/index.html) (EE) related configuration.",
		},
		"ansible_navigator_binary": {
			Description:         fmt.Sprintf("Path to the '%s' binary. By default '$PATH' is searched.", ansible.NavigatorProgram),
			MarkdownDescription: fmt.Sprintf("Path to the `%s` binary. By default `$PATH` is searched.", ansible.NavigatorProgram),
		},
		"ansible_options": {
			Description:         "Ansible playbook run related configuration.",
			MarkdownDescription: "Ansible [playbook](https://docs.ansible.com/ansible/latest/cli/ansible-playbook.html) run related configuration.",
		},
		"timezone": {
			Description:         fmt.Sprintf("IANA time zone, use 'local' for the system time zone. Defaults to '%s'.", defaultNavigatorRunTimezone),
			MarkdownDescription: fmt.Sprintf("IANA time zone, use `local` for the system time zone. Defaults to `%s`.", defaultNavigatorRunTimezone),
		},
		"artifact_queries": {
			Description:         "Query the Ansible playbook artifact with 'jq' syntax. The playbook artifact contains detailed information about every play and task, as well as the stdout from the playbook run.",
			MarkdownDescription: "Query the Ansible playbook artifact with [`jq`](https://jqlang.github.io/jq/) syntax. The [playbook artifact](https://access.redhat.com/documentation/en-us/red_hat_ansible_automation_platform/2.0-ea/html/ansible_navigator_creator_guide/assembly-troubleshooting-navigator_ansible-navigator#proc-review-artifact_troubleshooting-navigator) contains detailed information about every play and task, as well as the stdout from the playbook run.",
		},
		"id": {
			Description: "UUID.",
		},
		"command": {
			Description:         fmt.Sprintf("Generated '%s' run command. Useful for troubleshooting.", ansible.NavigatorProgram),
			MarkdownDescription: fmt.Sprintf("Generated `%s` run command. Useful for troubleshooting.", ansible.NavigatorProgram),
		},
	}
}

func (ExecutionEnvironmentModel) descriptions() map[string]attrDescription {
	return map[string]attrDescription{
		"container_engine": {
			Description:         fmt.Sprintf("Container engine responsible for running the execution environment container image. Options: %s. Defaults to '%s'.", wrapElementsJoin(ansible.ContainerEngineOptions(true), "'"), defaultNavigatorRunContainerEngine),
			MarkdownDescription: fmt.Sprintf("[Container engine](https://ansible.readthedocs.io/projects/navigator/settings/#container-engine) responsible for running the execution environment container image. Options: %s. Defaults to `%s`.", wrapElementsJoin(ansible.ContainerEngineOptions(true), "`"), defaultNavigatorRunContainerEngine),
		},
		"enabled": {
			Description:         fmt.Sprintf("Enable or disable the use of an execution environment. Disabling requires '%s' and is only recommended when without a container engine. Defaults to '%t'.", ansible.PlaybookProgram, defaultNavigatorRunEEEnabled),
			MarkdownDescription: fmt.Sprintf("Enable or disable the use of an execution environment. Disabling requires `%s` and is only recommended when without a container engine. Defaults to `%t`.", ansible.PlaybookProgram, defaultNavigatorRunEEEnabled),
		},
		"environment_variables_pass": {
			Description:         "Existing environment variables to be passed through to and set within the execution environment.",
			MarkdownDescription: "Existing environment variables to be [passed](https://ansible.readthedocs.io/projects/navigator/settings/#pass-environment-variable) through to and set within the execution environment.",
		},
		"environment_variables_set": {
			Description:         "Environment variables to be set within the execution environment.",
			MarkdownDescription: "Environment variables to be [set](https://ansible.readthedocs.io/projects/navigator/settings/#set-environment-variable) within the execution environment.",
		},
		"image": {
			Description:         fmt.Sprintf("Name of the execution environment container image. Defaults to '%s'.", defaultNavigatorRunImage),
			MarkdownDescription: fmt.Sprintf("Name of the execution environment container [image](https://ansible.readthedocs.io/projects/navigator/settings/#execution-environment-image). Defaults to `%s`.", defaultNavigatorRunImage),
		},
		"pull_arguments": {
			Description:         "Additional parameters that should be added to the pull command when pulling an execution environment container image from a container registry.",
			MarkdownDescription: "Additional [parameters](https://ansible.readthedocs.io/projects/navigator/settings/#pull-arguments) that should be added to the pull command when pulling an execution environment container image from a container registry.",
		},
		"pull_policy": {
			Description:         fmt.Sprintf("Container image pull policy. Defaults to '%s'.", defaultNavigatorRunPullPolicy),
			MarkdownDescription: fmt.Sprintf("Container image [pull policy](https://ansible.readthedocs.io/projects/navigator/settings/#pull-policy). Defaults to `%s`.", defaultNavigatorRunPullPolicy),
		},
		"container_options": {
			Description:         "Extra parameters passed to the container engine command.",
			MarkdownDescription: "[Extra parameters](https://ansible.readthedocs.io/projects/navigator/settings/#container-options) passed to the container engine command.",
		},
	}
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

func (AnsibleOptionsModel) descriptions() map[string]attrDescription {
	return map[string]attrDescription{
		"force_handlers": {
			Description: "Run handlers even if a task fails.",
		},
		"skip_tags": {
			Description: "Only run plays and tasks whose tags do not match these values.",
		},
		"start_at_task": {
			Description: "Start the playbook at the task matching this name.",
		},
		"limit": {
			Description: "Further limit selected hosts to an additional pattern.",
		},
		"tags": {
			Description: "Only run plays and tasks tagged with these values.",
		},
		"private_keys": {
			Description:         "SSH private keys used for authentication in addition to the automatically mounted default named keys and SSH agent socket path.",
			MarkdownDescription: "SSH private keys used for authentication in addition to the [automatically mounted](https://ansible.readthedocs.io/projects/navigator/faq/#how-do-i-use-my-ssh-keys-with-an-execution-environment) default named keys and SSH agent socket path.",
		},
		"known_hosts": {
			Description:         fmt.Sprintf("SSH known host entries. Ansible variable '%s' set to path of 'known_hosts' file and SSH option 'UserKnownHostsFile' must be configured to said path. Defaults to all of the 'known_hosts' entries recorded.", ansible.SSHKnownHostsFileVar),
			MarkdownDescription: fmt.Sprintf("SSH known host entries. Ansible variable `%s` set to path of `known_hosts` file and SSH option `UserKnownHostsFile` must be configured to said path. Defaults to all of the `known_hosts` entries recorded.", ansible.SSHKnownHostsFileVar),
		},
		"host_key_checking": {
			Description:         fmt.Sprintf("SSH host key checking. Can help protect against man-in-the-middle attacks by verifying the identity of hosts. Ansible runner (library used by '%s') defaults this option to '%t' explicitly.", ansible.NavigatorProgram, ansible.RunnerDefaultHostKeyChecking),
			MarkdownDescription: fmt.Sprintf("SSH host key checking. Can help protect against man-in-the-middle attacks by verifying the identity of hosts. Ansible runner (library used by `%s`) defaults this option to `%t` explicitly.", ansible.NavigatorProgram, ansible.RunnerDefaultHostKeyChecking),
		},
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

func (m *AnsibleOptionsModel) Set(ctx context.Context, run navigatorRun) diag.Diagnostics {
	var diags diag.Diagnostics

	if m.KnownHosts.IsUnknown() {
		knownHostsValue, newDiags := types.ListValueFrom(ctx, types.StringType, run.knownHosts)
		diags.Append(newDiags...)
		m.KnownHosts = knownHostsValue
	}

	return diags
}

func (PrivateKeyModel) descriptions() map[string]attrDescription {
	return map[string]attrDescription{
		"name": {
			Description: "Key name.",
		},
		"data": {
			Description: "Key data.",
		},
	}
}

func (PrivateKeyModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name": types.StringType,
		"data": types.StringType,
	}
}

func (m PrivateKeyModel) Value(_ context.Context, key *ansible.PrivateKey) diag.Diagnostics {
	var diags diag.Diagnostics

	key.Name = m.Name.ValueString()
	key.Data = m.Data.ValueString()

	return diags
}

func (ArtifactQueryModel) descriptions() map[string]attrDescription {
	return map[string]attrDescription{
		"jq_filter": {
			Description:         "'jq' filter. Example: '.status, .stdout'.",
			MarkdownDescription: "`jq` filter. Example: `.status, .stdout`.",
		},
		"results": {
			Description:         "Results of the 'jq' filter in JSON format.",
			MarkdownDescription: "Results of the `jq` filter in JSON format.",
		},
	}
}

func (ArtifactQueryModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"jq_filter": types.StringType,
		"results":   types.ListType{ElemType: types.StringType},
	}
}

func (m ArtifactQueryModel) Value(_ context.Context, query *ansible.ArtifactQuery) diag.Diagnostics {
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
