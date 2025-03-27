//nolint:dupl
package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
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
	run.inventories = []ansible.Inventory{{Name: navigatorRunName, Contents: m.Inventory.ValueString()}}
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

func (m *NavigatorRunDataSourceModel) Set(ctx context.Context, run navigatorRun) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Command = types.StringValue(run.command)

	var optsModel AnsibleOptionsModel
	diags.Append(m.AnsibleOptions.As(ctx, &optsModel, basetypes.ObjectAsOptions{})...)
	diags.Append(optsModel.Set(ctx, run)...)

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

func (d *NavigatorRunDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_navigator_run", req.ProviderTypeName)
}

//nolint:dupl
func (d *NavigatorRunDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         fmt.Sprintf("Run an Ansible playbook as a means to gather information. It is recommended to only run playbooks without observable side-effects. Requires '%s' and a container engine to run within an execution environment (EE).", ansible.NavigatorProgram),
		MarkdownDescription: fmt.Sprintf("Run an Ansible playbook as a means to gather information. It is recommended to only run playbooks without observable side-effects. Requires `%s` and a container engine to run within an execution environment (EE).", ansible.NavigatorProgram),
		Attributes: map[string]schema.Attribute{
			// required
			"playbook": schema.StringAttribute{
				Description:         navigatorRunDescriptions()["playbook"].Description,
				MarkdownDescription: navigatorRunDescriptions()["playbook"].MarkdownDescription,
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringIsYAML(),
				},
			},
			"inventory": schema.StringAttribute{
				Description:         navigatorRunDescriptions()["inventory"].Description,
				MarkdownDescription: navigatorRunDescriptions()["inventory"].MarkdownDescription,
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			// optional
			"working_directory": schema.StringAttribute{
				Description:         navigatorRunDescriptions()["working_directory"].Description,
				MarkdownDescription: navigatorRunDescriptions()["working_directory"].MarkdownDescription,
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"execution_environment": schema.SingleNestedAttribute{
				Description:         navigatorRunDescriptions()["execution_environment"].Description,
				MarkdownDescription: navigatorRunDescriptions()["execution_environment"].MarkdownDescription,
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"container_engine": schema.StringAttribute{
						Description:         ExecutionEnvironmentModel{}.descriptions()["container_engine"].Description,
						MarkdownDescription: ExecutionEnvironmentModel{}.descriptions()["container_engine"].MarkdownDescription,
						Optional:            true,
						Computed:            true,
						Validators: []validator.String{
							stringvalidator.OneOf(ansible.ContainerEngineOptions(true)...),
						},
					},
					"enabled": schema.BoolAttribute{
						Description:         ExecutionEnvironmentModel{}.descriptions()["enabled"].Description,
						MarkdownDescription: ExecutionEnvironmentModel{}.descriptions()["enabled"].MarkdownDescription,
						Optional:            true,
						Computed:            true,
					},
					"environment_variables_pass": schema.ListAttribute{
						Description:         ExecutionEnvironmentModel{}.descriptions()["environment_variables_pass"].Description,
						MarkdownDescription: ExecutionEnvironmentModel{}.descriptions()["environment_variables_pass"].MarkdownDescription,
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							listvalidator.ValueStringsAre(stringIsEnvVarName()),
						},
					},
					"environment_variables_set": schema.MapAttribute{
						Description:         fmt.Sprintf("%s '%s' is automatically set to '%s'.", ExecutionEnvironmentModel{}.descriptions()["environment_variables_set"].Description, navigatorRunOperationEnvVar, terraformOp(terraformOpRead)),
						MarkdownDescription: fmt.Sprintf("%s `%s` is automatically set to `%s`.", ExecutionEnvironmentModel{}.descriptions()["environment_variables_set"].MarkdownDescription, navigatorRunOperationEnvVar, terraformOp(terraformOpRead)),
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.Map{
							mapvalidator.KeysAre(stringIsEnvVarName()),
						},
					},
					"image": schema.StringAttribute{
						Description:         ExecutionEnvironmentModel{}.descriptions()["image"].Description,
						MarkdownDescription: ExecutionEnvironmentModel{}.descriptions()["image"].MarkdownDescription,
						Optional:            true,
						Computed:            true,
						Validators: []validator.String{
							stringIsContainerImageName(),
						},
					},
					"pull_arguments": schema.ListAttribute{
						Description:         ExecutionEnvironmentModel{}.descriptions()["pull_arguments"].Description,
						MarkdownDescription: ExecutionEnvironmentModel{}.descriptions()["pull_arguments"].MarkdownDescription,
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
						},
					},
					"pull_policy": schema.StringAttribute{
						Description:         ExecutionEnvironmentModel{}.descriptions()["pull_policy"].Description,
						MarkdownDescription: ExecutionEnvironmentModel{}.descriptions()["pull_policy"].MarkdownDescription,
						Optional:            true,
						Computed:            true,
						Validators: []validator.String{
							stringvalidator.OneOf(ansible.PullPolicyOptions()...),
						},
					},
					"container_options": schema.ListAttribute{
						Description:         ExecutionEnvironmentModel{}.descriptions()["container_options"].Description,
						MarkdownDescription: ExecutionEnvironmentModel{}.descriptions()["container_options"].MarkdownDescription,
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
						},
					},
				},
			},
			"ansible_navigator_binary": schema.StringAttribute{
				Description:         navigatorRunDescriptions()["ansible_navigator_binary"].Description,
				MarkdownDescription: navigatorRunDescriptions()["ansible_navigator_binary"].MarkdownDescription,
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"ansible_options": schema.SingleNestedAttribute{
				Description:         navigatorRunDescriptions()["ansible_options"].Description,
				MarkdownDescription: navigatorRunDescriptions()["ansible_options"].MarkdownDescription,
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"force_handlers": schema.BoolAttribute{
						Description: AnsibleOptionsModel{}.descriptions()["force_handlers"].Description,
						Optional:    true,
					},
					"skip_tags": schema.ListAttribute{
						Description: AnsibleOptionsModel{}.descriptions()["skip_tags"].Description,
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
						},
					},
					"start_at_task": schema.StringAttribute{
						Description: AnsibleOptionsModel{}.descriptions()["start_at_task"].Description,
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					"limit": schema.ListAttribute{
						Description: AnsibleOptionsModel{}.descriptions()["limit"].Description,
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
						},
					},
					"tags": schema.ListAttribute{
						Description: AnsibleOptionsModel{}.descriptions()["tags"].Description,
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
						},
					},
					"private_keys": schema.ListNestedAttribute{
						Description:         AnsibleOptionsModel{}.descriptions()["private_keys"].Description,
						MarkdownDescription: AnsibleOptionsModel{}.descriptions()["private_keys"].MarkdownDescription,
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: PrivateKeyModel{}.descriptions()["name"].Description,
									Required:    true,
									Validators: []validator.String{
										stringIsSSHPrivateKeyName(),
									},
								},
								"data": schema.StringAttribute{
									Description: PrivateKeyModel{}.descriptions()["data"].Description,
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
						Description:         AnsibleOptionsModel{}.descriptions()["known_hosts"].Description,
						MarkdownDescription: AnsibleOptionsModel{}.descriptions()["known_hosts"].MarkdownDescription,
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							listvalidator.ValueStringsAre(stringIsSSHKnownHost()),
						},
					},
					"host_key_checking": schema.BoolAttribute{
						Description:         AnsibleOptionsModel{}.descriptions()["host_key_checking"].Description,
						MarkdownDescription: AnsibleOptionsModel{}.descriptions()["host_key_checking"].MarkdownDescription,
						Optional:            true,
					},
				},
			},
			"timezone": schema.StringAttribute{
				Description:         navigatorRunDescriptions()["timezone"].Description,
				MarkdownDescription: navigatorRunDescriptions()["timezone"].MarkdownDescription,
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringIsIANATimezone(),
				},
			},
			"artifact_queries": schema.MapNestedAttribute{
				Description:         navigatorRunDescriptions()["artifact_queries"].Description,
				MarkdownDescription: navigatorRunDescriptions()["artifact_queries"].MarkdownDescription,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"jq_filter": schema.StringAttribute{
							Description:         ArtifactQueryModel{}.descriptions()["jq_filter"].Description,
							MarkdownDescription: ArtifactQueryModel{}.descriptions()["jq_filter"].MarkdownDescription,
							Required:            true,
							Validators: []validator.String{
								stringIsJQFilter(),
							},
						},
						"results": schema.ListAttribute{ // TODO switch to a dynamic attribute when supported as an element in a collection
							Description:         ArtifactQueryModel{}.descriptions()["results"].Description,
							MarkdownDescription: ArtifactQueryModel{}.descriptions()["results"].MarkdownDescription,
							Computed:            true,
							ElementType:         jsontypes.NormalizedType{},
						},
					},
				},
			},
			"id": schema.StringAttribute{
				Description: navigatorRunDescriptions()["id"].Description,
				Computed:    true,
			},
			"command": schema.StringAttribute{
				Description:         navigatorRunDescriptions()["command"].Description,
				MarkdownDescription: navigatorRunDescriptions()["command"].MarkdownDescription,
				Computed:            true,
			},
			// TODO include defaultNavigatorRunTimeout in description
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *NavigatorRunDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	run(ctx, &resp.Diagnostics, timeout, terraformOpRead, &navigatorRun)
	resp.Diagnostics.Append(data.Set(ctx, navigatorRun)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
