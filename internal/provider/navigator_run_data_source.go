package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

var (
	_ datasource.DataSource              = &NavigatorRunDataSource{}
	_ datasource.DataSourceWithConfigure = &NavigatorRunDataSource{}
)

func NewNavigatorRunDataSource() datasource.DataSource { //nolint:ireturn
	return &NavigatorRunDataSource{}
}

type NavigatorRunDataSource struct {
	opts *providerOptions
}

type NavigatorRunDataSourceModel struct {
	Playbook               types.String   `tfsdk:"playbook"`
	Inventory              types.String   `tfsdk:"inventory"`
	WorkingDirectory       types.String   `tfsdk:"working_directory"`
	ExecutionEnvironment   types.Object   `tfsdk:"execution_environment"`
	AnsibleNavigatorBinary types.String   `tfsdk:"ansible_navigator_binary"`
	AnsibleOptions         types.Object   `tfsdk:"ansible_options"`
	Timezone               types.String   `tfsdk:"timezone"`
	ArtifactQueries        types.Map      `tfsdk:"artifact_queries"`
	ID                     types.String   `tfsdk:"id"`
	Command                types.String   `tfsdk:"command"`
	Timeouts               timeouts.Value `tfsdk:"timeouts"`
}

func (m NavigatorRunDataSourceModel) Value(ctx context.Context, run *navigatorRun, opts *providerOptions) diag.Diagnostics {
	var diags diag.Diagnostics

	run.dir = runDir(opts.BaseRunDirectory, m.ID.ValueString(), 0)
	run.persistDir = opts.PersistRunDirectory
	run.playbook = m.Playbook.ValueString()
	run.inventory = m.Inventory.ValueString()
	run.workingDir = m.WorkingDirectory.ValueString()
	run.navigatorBinary = m.AnsibleNavigatorBinary.ValueString()

	var eeModel ExecutionEnvironmentModel
	diags.Append(m.ExecutionEnvironment.As(ctx, &eeModel, basetypes.ObjectAsOptions{})...)

	run.navigatorSettings.Timezone = m.Timezone.ValueString()

	diags.Append(eeModel.Value(ctx, &run.navigatorSettings)...)

	var optsModel AnsibleOptionsModel
	diags.Append(m.AnsibleOptions.As(ctx, &optsModel, basetypes.ObjectAsOptions{})...)

	diags.Append(optsModel.Value(ctx, &run.options)...)

	var privateKeysModel []PrivateKeyModel
	if !optsModel.PrivateKeys.IsNull() {
		diags.Append(optsModel.PrivateKeys.ElementsAs(ctx, &privateKeysModel, false)...)
	}

	run.privateKeys = make([]ansible.PrivateKey, 0, len(privateKeysModel))
	for _, model := range privateKeysModel {
		var key ansible.PrivateKey

		diags.Append(model.Value(ctx, &key)...)
		run.privateKeys = append(run.privateKeys, key)
	}

	var knownHosts []string
	if !optsModel.KnownHosts.IsUnknown() {
		diags.Append(optsModel.KnownHosts.ElementsAs(ctx, &knownHosts, false)...)
	}

	run.knownHosts = knownHosts

	var queriesModel map[string]ArtifactQueryModel
	diags.Append(m.ArtifactQueries.ElementsAs(ctx, &queriesModel, false)...)

	run.artifactQueries = map[string]ansible.ArtifactQuery{}
	for name, model := range queriesModel {
		var query ansible.ArtifactQuery

		diags.Append(model.Value(ctx, &query)...)
		run.artifactQueries[name] = query
	}

	return diags
}

//nolint:dupl
func (m *NavigatorRunDataSourceModel) Set(ctx context.Context, run navigatorRun) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Command = types.StringValue(run.command)

	var optsModel AnsibleOptionsModel
	diags.Append(m.AnsibleOptions.As(ctx, &optsModel, basetypes.ObjectAsOptions{})...)
	diags.Append(optsModel.Set(ctx, &run)...)

	optsResults, newDiags := types.ObjectValueFrom(ctx, AnsibleOptionsModel{}.AttrTypes(), optsModel)
	diags.Append(newDiags...)
	m.AnsibleOptions = optsResults

	var queriesModel map[string]ArtifactQueryModel
	diags.Append(m.ArtifactQueries.ElementsAs(ctx, &queriesModel, false)...)

	for name, model := range queriesModel {
		diags.Append(model.Set(ctx, run.artifactQueries[name])...)
		queriesModel[name] = model
	}

	queriesValue, newDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: ArtifactQueryModel{}.AttrTypes()}, queriesModel)
	diags.Append(newDiags...)
	m.ArtifactQueries = queriesValue

	return diags
}

func (m *NavigatorRunDataSourceModel) SetDefaults(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics

	if m.WorkingDirectory.IsNull() {
		m.WorkingDirectory = types.StringValue(defaultNavigatorRunWorkingDir)
	}

	if m.ExecutionEnvironment.IsNull() {
		m.ExecutionEnvironment = ExecutionEnvironmentModel{}.Defaults()
	}

	var eeModel ExecutionEnvironmentModel
	diags.Append(m.ExecutionEnvironment.As(ctx, &eeModel, basetypes.ObjectAsOptions{})...)

	if eeModel.ContainerEngine.IsNull() {
		eeModel.ContainerEngine = types.StringValue(defaultNavigatorRunContainerEngine)
	}

	if eeModel.Enabled.IsNull() {
		eeModel.Enabled = types.BoolValue(defaultNavigatorRunEEEnabled)
	}

	if eeModel.Image.IsNull() {
		eeModel.Image = types.StringValue(defaultNavigatorRunImage)
	}

	if eeModel.PullPolicy.IsNull() {
		eeModel.PullPolicy = types.StringValue(defaultNavigatorRunPullPolicy)
	}

	eeValue, newDiags := types.ObjectValueFrom(ctx, ExecutionEnvironmentModel{}.AttrTypes(), eeModel)
	diags.Append(newDiags...)
	m.ExecutionEnvironment = eeValue

	if m.AnsibleOptions.IsNull() {
		m.AnsibleOptions = AnsibleOptionsModel{}.Defaults()
	}

	var optsModel AnsibleOptionsModel
	diags.Append(m.AnsibleOptions.As(ctx, &optsModel, basetypes.ObjectAsOptions{})...)

	if optsModel.KnownHosts.IsNull() {
		optsModel.KnownHosts = types.ListUnknown(types.StringType)
	}

	optsResults, newDiags := types.ObjectValueFrom(ctx, AnsibleOptionsModel{}.AttrTypes(), optsModel)
	diags.Append(newDiags...)
	m.AnsibleOptions = optsResults

	if m.Timezone.IsNull() {
		m.Timezone = types.StringValue(defaultNavigatorRunTimezone)
	}

	return diags
}

func (d *NavigatorRunDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_navigator_run", req.ProviderTypeName)
}

//nolint:dupl
func (d *NavigatorRunDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         fmt.Sprintf("Run an Ansible playbook as a means to gather information. It is recommended to only run playbooks without observable side-effects. Requires '%s' and a container engine to run within an execution environment (EE).", ansible.NavigatorProgram),
		MarkdownDescription: fmt.Sprintf("Run an Ansible playbook as a means to gather information. It is recommended to only run playbooks without observable side-effects. Requires `%s` and a container engine to run within an execution environment (EE).", ansible.NavigatorProgram),
		Attributes: map[string]schema.Attribute{
			// required
			"playbook": schema.StringAttribute{
				Description:         "Ansible playbook contents.",
				MarkdownDescription: "Ansible [playbook](https://docs.ansible.com/ansible/latest/playbook_guide/playbooks_intro.html) contents.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringIsYAML(),
				},
			},
			"inventory": schema.StringAttribute{
				Description:         "Ansible inventory contents.",
				MarkdownDescription: "Ansible [inventory](https://docs.ansible.com/ansible/latest/getting_started/get_started_inventory.html) contents.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			// optional
			"working_directory": schema.StringAttribute{
				Description:         fmt.Sprintf("Directory which '%s' is run from. Recommended to be the root Ansible content directory (sometimes called the project directory), which is likely to contain 'ansible.cfg', 'roles/', etc. Defaults to '%s'.", ansible.NavigatorProgram, defaultNavigatorRunWorkingDir),
				MarkdownDescription: fmt.Sprintf("Directory which `%s` is run from. Recommended to be the root Ansible [content directory](https://docs.ansible.com/ansible/latest/tips_tricks/sample_setup.html#sample-directory-layout) (sometimes called the project directory), which is likely to contain `ansible.cfg`, `roles/`, etc. Defaults to `%s`.", ansible.NavigatorProgram, defaultNavigatorRunWorkingDir),
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"execution_environment": schema.SingleNestedAttribute{
				Description:         "Execution environment (EE) related configuration.",
				MarkdownDescription: "[Execution environment](https://ansible.readthedocs.io/en/latest/getting_started_ee/index.html) (EE) related configuration.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"container_engine": schema.StringAttribute{
						Description:         fmt.Sprintf("Container engine responsible for running the execution environment container image. Options: %s. Defaults to '%s'.", wrapElementsJoin(ansible.ContainerEngineOptions(true), "'"), defaultNavigatorRunContainerEngine),
						MarkdownDescription: fmt.Sprintf("[Container engine](https://ansible.readthedocs.io/projects/navigator/settings/#container-engine) responsible for running the execution environment container image. Options: %s. Defaults to `%s`.", wrapElementsJoin(ansible.ContainerEngineOptions(true), "`"), defaultNavigatorRunContainerEngine),
						Optional:            true,
						Computed:            true,
						Validators: []validator.String{
							stringvalidator.OneOf(ansible.ContainerEngineOptions(true)...),
						},
					},
					"enabled": schema.BoolAttribute{
						Description:         fmt.Sprintf("Enable or disable the use of an execution environment. Disabling requires '%s' and is only recommended when without a container engine. Defaults to '%t'.", ansible.PlaybookProgram, defaultNavigatorRunEEEnabled),
						MarkdownDescription: fmt.Sprintf("Enable or disable the use of an execution environment. Disabling requires `%s` and is only recommended when without a container engine. Defaults to `%t`.", ansible.PlaybookProgram, defaultNavigatorRunEEEnabled),
						Optional:            true,
						Computed:            true,
					},
					"environment_variables_pass": schema.ListAttribute{
						Description:         "Existing environment variables to be passed through to and set within the execution environment.",
						MarkdownDescription: "Existing environment variables to be [passed](https://ansible.readthedocs.io/projects/navigator/settings/#pass-environment-variable) through to and set within the execution environment.",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							listvalidator.ValueStringsAre(stringIsEnvVarName()),
						},
					},
					"environment_variables_set": schema.MapAttribute{
						Description:         fmt.Sprintf("Environment variables to be set within the execution environment. By default '%s' is set to '%s'.", navigatorRunOperationEnvVar, terraformOperation(terraformOperationRead).String()),
						MarkdownDescription: fmt.Sprintf("Environment variables to be [set](https://ansible.readthedocs.io/projects/navigator/settings/#set-environment-variable) within the execution environment. By default `%s` is set to `%s`.", navigatorRunOperationEnvVar, terraformOperation(terraformOperationRead).String()),
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.Map{
							mapvalidator.KeysAre(stringIsEnvVarName()),
						},
					},
					"image": schema.StringAttribute{
						Description:         fmt.Sprintf("Name of the execution environment container image. Defaults to '%s'.", defaultNavigatorRunImage),
						MarkdownDescription: fmt.Sprintf("Name of the execution environment container [image](https://ansible.readthedocs.io/projects/navigator/settings/#execution-environment-image). Defaults to `%s`.", defaultNavigatorRunImage),
						Optional:            true,
						Computed:            true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					"pull_arguments": schema.ListAttribute{
						Description:         "Additional parameters that should be added to the pull command when pulling an execution environment container image from a container registry.",
						MarkdownDescription: "Additional [parameters](https://ansible.readthedocs.io/projects/navigator/settings/#pull-arguments) that should be added to the pull command when pulling an execution environment container image from a container registry.",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
						},
					},
					"pull_policy": schema.StringAttribute{
						Description:         fmt.Sprintf("Container image pull policy. Defaults to '%s'.", defaultNavigatorRunPullPolicy),
						MarkdownDescription: fmt.Sprintf("Container image [pull policy](https://ansible.readthedocs.io/projects/navigator/settings/#pull-policy). Defaults to `%s`.", defaultNavigatorRunPullPolicy),
						Optional:            true,
						Computed:            true,
						Validators: []validator.String{
							stringvalidator.OneOf(ansible.PullPolicyOptions()...),
						},
					},
					"container_options": schema.ListAttribute{
						Description:         "Extra parameters passed to the container engine command.",
						MarkdownDescription: "[Extra parameters](https://ansible.readthedocs.io/projects/navigator/settings/#container-options) passed to the container engine command.",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
						},
					},
				},
			},
			"ansible_navigator_binary": schema.StringAttribute{
				Description:         fmt.Sprintf("Path to the '%s' binary. By default '$PATH' is searched.", ansible.NavigatorProgram),
				MarkdownDescription: fmt.Sprintf("Path to the `%s` binary. By default `$PATH` is searched.", ansible.NavigatorProgram),
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"ansible_options": schema.SingleNestedAttribute{
				Description:         "Ansible playbook run related configuration.",
				MarkdownDescription: "Ansible [playbook](https://docs.ansible.com/ansible/latest/cli/ansible-playbook.html) run related configuration.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"force_handlers": schema.BoolAttribute{
						Description: "Run handlers even if a task fails.",
						Optional:    true,
					},
					"skip_tags": schema.ListAttribute{
						Description: "Only run plays and tasks whose tags do not match these values.",
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
						},
					},
					"start_at_task": schema.StringAttribute{
						Description: "Start the playbook at the task matching this name.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					"limit": schema.ListAttribute{
						Description: "Further limit selected hosts to an additional pattern.",
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
						},
					},
					"tags": schema.ListAttribute{
						Description: "Only run plays and tasks tagged with these values.",
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
						},
					},
					"private_keys": schema.ListNestedAttribute{
						Description:         "SSH private keys used for authentication in addition to the automatically mounted default named keys and SSH agent socket path.",
						MarkdownDescription: "SSH private keys used for authentication in addition to the [automatically mounted](https://ansible.readthedocs.io/projects/navigator/faq/#how-do-i-use-my-ssh-keys-with-an-execution-environment) default named keys and SSH agent socket path.",
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "Key name.",
									Required:    true,
									Validators: []validator.String{
										stringvalidator.RegexMatches(
											regexp.MustCompile(`^[a-zA-Z0-9]+(-[a-zA-Z0-9]+)*$`),
											"Must only contain dashes and alphanumeric characters",
										),
									},
								},
								"data": schema.StringAttribute{
									Description: "Key data.",
									Required:    true,
									Sensitive:   true,
									Validators: []validator.String{
										stringIsSSHPrivateKey(),
									},
								},
							},
						},
					},
					"known_hosts": schema.ListAttribute{
						Description:         fmt.Sprintf("SSH known host entries. Can help protect against man-in-the-middle attacks by verifying the identity of hosts. Ansible variable '%s' set to path of 'known_hosts' file. If unspecified will be set to contents of 'known_hosts' file after run.", ansible.SSHKnownHostsFileVar),
						MarkdownDescription: fmt.Sprintf("SSH known host entries. Can help protect against man-in-the-middle attacks by verifying the identity of hosts. Ansible variable `%s` set to path of `known_hosts` file. If unspecified will be set to contents of `known_hosts` file after run.", ansible.SSHKnownHostsFileVar),
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							listvalidator.ValueStringsAre(stringIsSSHKnownHost()),
						},
					},
				},
			},
			"timezone": schema.StringAttribute{
				Description:         fmt.Sprintf("IANA time zone, use 'local' for the system time zone. Defaults to '%s'.", defaultNavigatorRunTimezone),
				MarkdownDescription: fmt.Sprintf("IANA time zone, use `local` for the system time zone. Defaults to `%s`.", defaultNavigatorRunTimezone),
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringIsIANATimezone(),
				},
			},
			"artifact_queries": schema.MapNestedAttribute{
				Description:         "Query the Ansible playbook artifact with 'jq' syntax. The playbook artifact contains detailed information about every play and task, as well as the stdout from the playbook run.",
				MarkdownDescription: "Query the Ansible playbook artifact with [`jq`](https://jqlang.github.io/jq/) syntax. The [playbook artifact](https://access.redhat.com/documentation/en-us/red_hat_ansible_automation_platform/2.0-ea/html/ansible_navigator_creator_guide/assembly-troubleshooting-navigator_ansible-navigator#proc-review-artifact_troubleshooting-navigator) contains detailed information about every play and task, as well as the stdout from the playbook run.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"jq_filter": schema.StringAttribute{
							Description:         "'jq' filter. Example: '.status, .stdout'.",
							MarkdownDescription: "`jq` filter. Example: `.status, .stdout`.",
							Required:            true,
							Validators: []validator.String{
								stringIsIsJQFilter(),
							},
						},
						"results": schema.ListAttribute{ // TODO switch to a dynamic attribute when supported as an element in a collection
							Description:         "Results of the 'jq' filter in JSON format.",
							MarkdownDescription: "Results of the `jq` filter in JSON format.",
							Computed:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
			// computed
			"id": schema.StringAttribute{
				Description: "UUID.",
				Computed:    true,
			},
			"command": schema.StringAttribute{
				Description:         fmt.Sprintf("Generated '%s' run command. Useful for troubleshooting.", ansible.NavigatorProgram),
				MarkdownDescription: fmt.Sprintf("Generated `%s` run command. Useful for troubleshooting.", ansible.NavigatorProgram),
				Computed:            true,
			},
			// timeouts
			// TODO include defaultNavigatorRunTimeout in description
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *NavigatorRunDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	opts, ok := configureDataSourceClient(req, resp)
	if !ok {
		return
	}

	d.opts = opts
}

func (d *NavigatorRunDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *NavigatorRunDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	resp.Diagnostics.Append(data.SetDefaults(ctx)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, newDiags := terraformOperationDataSourceTimeout(ctx, data.Timeouts, defaultNavigatorRunTimeout)
	resp.Diagnostics.Append(newDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	data.ID = types.StringValue(uuid.New().String())

	var navigatorRun navigatorRun
	resp.Diagnostics.Append(data.Value(ctx, &navigatorRun, d.opts)...)

	if resp.Diagnostics.HasError() {
		return
	}

	run(ctx, &resp.Diagnostics, timeout, terraformOperationRead, &navigatorRun)
	resp.Diagnostics.Append(data.Set(ctx, navigatorRun)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
