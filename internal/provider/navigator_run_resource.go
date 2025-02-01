package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/dynamicplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

var (
	_ resource.Resource               = &NavigatorRunResource{}
	_ resource.ResourceWithModifyPlan = &NavigatorRunResource{}
)

func NewNavigatorRunResource() resource.Resource { //nolint:ireturn
	return &NavigatorRunResource{}
}

type NavigatorRunResource struct {
	opts *providerOptions
}

type NavigatorRunResourceModel struct {
	Playbook               types.String   `tfsdk:"playbook"`
	Inventory              types.String   `tfsdk:"inventory"`
	WorkingDirectory       types.String   `tfsdk:"working_directory"`
	ExecutionEnvironment   types.Object   `tfsdk:"execution_environment"`
	AnsibleNavigatorBinary types.String   `tfsdk:"ansible_navigator_binary"`
	AnsibleOptions         types.Object   `tfsdk:"ansible_options"`
	Timezone               types.String   `tfsdk:"timezone"`
	RunOnDestroy           types.Bool     `tfsdk:"run_on_destroy"`
	Triggers               types.Object   `tfsdk:"triggers"`
	ArtifactQueries        types.Map      `tfsdk:"artifact_queries"`
	ID                     types.String   `tfsdk:"id"`
	Command                types.String   `tfsdk:"command"`
	Timeouts               timeouts.Value `tfsdk:"timeouts"`
}

func (m NavigatorRunResourceModel) Value(ctx context.Context, run *navigatorRun, opts *providerOptions, runs uint32) diag.Diagnostics {
	var diags diag.Diagnostics

	run.dir = runDir(opts.BaseRunDirectory, m.ID.ValueString(), runs)
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
func (m *NavigatorRunResourceModel) Set(ctx context.Context, run navigatorRun) diag.Diagnostics {
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

func (r *NavigatorRunResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_navigator_run", req.ProviderTypeName)
}

//nolint:dupl,maintidx
func (r *NavigatorRunResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         fmt.Sprintf("Run an Ansible playbook. Requires '%s' and a container engine to run within an execution environment (EE).", ansible.NavigatorProgram),
		MarkdownDescription: fmt.Sprintf("Run an Ansible playbook. Requires `%s` and a container engine to run within an execution environment (EE).", ansible.NavigatorProgram),
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
				Default:             stringdefault.StaticString(defaultNavigatorRunWorkingDir),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"execution_environment": schema.SingleNestedAttribute{
				Description:         navigatorRunDescriptions()["execution_environment"].Description,
				MarkdownDescription: navigatorRunDescriptions()["execution_environment"].MarkdownDescription,
				Optional:            true,
				Computed:            true,
				Default:             objectdefault.StaticValue(ExecutionEnvironmentModel{}.Defaults()),
				Attributes: map[string]schema.Attribute{
					"container_engine": schema.StringAttribute{
						Description:         ExecutionEnvironmentModel{}.descriptions()["container_engine"].Description,
						MarkdownDescription: ExecutionEnvironmentModel{}.descriptions()["container_engine"].MarkdownDescription,
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(defaultNavigatorRunContainerEngine),
						Validators: []validator.String{
							stringvalidator.OneOf(ansible.ContainerEngineOptions(true)...),
						},
					},
					"enabled": schema.BoolAttribute{
						Description:         ExecutionEnvironmentModel{}.descriptions()["enabled"].Description,
						MarkdownDescription: ExecutionEnvironmentModel{}.descriptions()["enabled"].MarkdownDescription,
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(defaultNavigatorRunEEEnabled),
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
						Description:         fmt.Sprintf("%s '%s' is automatically set to the current CRUD operation (%s).", ExecutionEnvironmentModel{}.descriptions()["environment_variables_set"].Description, navigatorRunOperationEnvVar, wrapElementsJoin(terraformOps([]terraformOp{terraformOpCreate, terraformOpUpdate, terraformOpDelete}).Strings(), "'")),
						MarkdownDescription: fmt.Sprintf("%s `%s` is automatically set to the current CRUD operation (%s).", ExecutionEnvironmentModel{}.descriptions()["environment_variables_set"].MarkdownDescription, navigatorRunOperationEnvVar, wrapElementsJoin(terraformOps([]terraformOp{terraformOpCreate, terraformOpUpdate, terraformOpDelete}).Strings(), "`")),
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
						Default:             stringdefault.StaticString(defaultNavigatorRunImage),
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
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
						Default:             stringdefault.StaticString(defaultNavigatorRunPullPolicy),
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
				Default:             objectdefault.StaticValue(AnsibleOptionsModel{}.Defaults()),
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
										stringvalidator.RegexMatches(
											regexp.MustCompile(`^[a-zA-Z0-9]+(-[a-zA-Z0-9]+)*$`),
											"Must only contain dashes and alphanumeric characters",
										),
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
				Default:             stringdefault.StaticString(defaultNavigatorRunTimezone),
				Validators: []validator.String{
					stringIsIANATimezone(),
				},
			},
			"run_on_destroy": schema.BoolAttribute{
				Description:         fmt.Sprintf("Run playbook on destroy. The environment variable '%s' is set to '%s' during the run to allow for conditional plays, tasks, etc. Defaults to '%t'.", navigatorRunOperationEnvVar, terraformOp(terraformOpDelete), defaultNavigatorRunOnDestroy),
				MarkdownDescription: fmt.Sprintf("Run playbook on destroy. The environment variable `%s` is set to `%s` during the run to allow for conditional plays, tasks, etc. Defaults to `%t`.", navigatorRunOperationEnvVar, terraformOp(terraformOpDelete), defaultNavigatorRunOnDestroy),
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(defaultNavigatorRunOnDestroy),
			},
			"triggers": schema.SingleNestedAttribute{
				Description: "Trigger various behaviors via arbitrary values.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"run": schema.DynamicAttribute{
						Description: "A value that, when changed, will run the playbook again. Provides a way to initiate a run without changing other attributes such as the inventory or playbook.",
						Optional:    true,
					},
					"replace": schema.DynamicAttribute{
						Description:         "A value that, when changed, will recreate the resource. Serves as an alternative to the native 'replace_triggered_by' lifecycle argument. Will cause 'id' to change. May be useful when combined with 'run_on_destroy'.",
						MarkdownDescription: "A value that, when changed, will recreate the resource. Serves as an alternative to the native [`replace_triggered_by`](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#replace_triggered_by) lifecycle argument. Will cause `id` to change. May be useful when combined with `run_on_destroy`.",
						Optional:            true,
						PlanModifiers: []planmodifier.Dynamic{
							dynamicplanmodifier.RequiresReplace(),
						},
					},
					"known_hosts": schema.DynamicAttribute{
						Description: "A value that, when changed, will reset the computed list of SSH known host entries. Useful when inventory hosts are recreated with the same hostnames/IP addresses, but different SSH keypairs.",
						Optional:    true,
					},
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
								stringIsIsJQFilter(),
							},
						},
						"results": schema.ListAttribute{ // TODO switch to a dynamic attribute when supported as an element in a collection
							Description:         ArtifactQueryModel{}.descriptions()["results"].Description,
							MarkdownDescription: ArtifactQueryModel{}.descriptions()["results"].MarkdownDescription,
							Computed:            true,
							ElementType:         types.StringType,
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"id": schema.StringAttribute{
				Description: navigatorRunDescriptions()["id"].Description,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"command": schema.StringAttribute{
				Description:         navigatorRunDescriptions()["command"].Description,
				MarkdownDescription: navigatorRunDescriptions()["command"].MarkdownDescription,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			// TODO include defaultNavigatorRunTimeout in description
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

// TODO find better solution
func (NavigatorRunResource) TriggersAttr(data *NavigatorRunResourceModel, attribute string) attr.Value { //nolint:ireturn
	if data.Triggers.IsNull() {
		return types.DynamicNull()
	}

	return data.Triggers.Attributes()[attribute]
}

func (r *NavigatorRunResource) ShouldRun(plan *NavigatorRunResourceModel, state *NavigatorRunResourceModel) bool {
	// skip working_directory, ansible_navigator_binary, run_on_destroy, timeouts
	attributeChanges := []bool{
		plan.Playbook.Equal(state.Playbook),
		plan.Inventory.Equal(state.Inventory),
		plan.ExecutionEnvironment.Equal(state.ExecutionEnvironment),
		plan.AnsibleOptions.Equal(state.AnsibleOptions),
		plan.Timezone.Equal(state.Timezone),
		r.TriggersAttr(plan, "run").Equal(r.TriggersAttr(state, "run")),
		plan.ArtifactQueries.Equal(state.ArtifactQueries),
	}

	for _, attributeChange := range attributeChanges {
		if !attributeChange {
			return true
		}
	}

	return false
}

func (r *NavigatorRunResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	opts, ok := configureResourceClient(req, resp)
	if !ok {
		return
	}

	r.opts = opts
}

//nolint:cyclop
func (r *NavigatorRunResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var data, state *NavigatorRunResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if req.Plan.Raw.IsNull() && state.RunOnDestroy.ValueBool() {
		resp.Diagnostics.AddWarning(
			"Resource Destruction Considerations",
			"Applying this resource destruction with 'run_on_destroy' enabled will run the playbook as configured in state. "+
				"The playbook run must complete successfully to remove the resource from Terraform state. ",
		)
	}

	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	defer func() {
		if !resp.Diagnostics.HasError() {
			resp.Diagnostics.Append(resp.Plan.Set(ctx, &data)...)
		}
	}()

	var optsPlanModel, optsStateModel AnsibleOptionsModel
	resp.Diagnostics.Append(data.AnsibleOptions.As(ctx, &optsPlanModel, basetypes.ObjectAsOptions{})...)
	resp.Diagnostics.Append(state.AnsibleOptions.As(ctx, &optsStateModel, basetypes.ObjectAsOptions{})...)

	if optsPlanModel.KnownHosts.IsUnknown() && r.TriggersAttr(data, "known_hosts").Equal(r.TriggersAttr(state, "known_hosts")) {
		optsPlanModel.KnownHosts = optsStateModel.KnownHosts
	}

	optsPlanValue, newDiags := types.ObjectValueFrom(ctx, AnsibleOptionsModel{}.AttrTypes(), optsPlanModel)
	resp.Diagnostics.Append(newDiags...)
	data.AnsibleOptions = optsPlanValue

	if !r.ShouldRun(data, state) {
		return
	}

	data.Command = types.StringUnknown()

	var artifactQueriesPlanModel map[string]ArtifactQueryModel
	resp.Diagnostics.Append(data.ArtifactQueries.ElementsAs(ctx, &artifactQueriesPlanModel, false)...)

	for name, model := range artifactQueriesPlanModel {
		model.Results = types.ListUnknown(types.StringType)
		artifactQueriesPlanModel[name] = model
	}

	artifactQueriesPlanValue, newDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: ArtifactQueryModel{}.AttrTypes()}, artifactQueriesPlanModel)
	resp.Diagnostics.Append(newDiags...)
	data.ArtifactQueries = artifactQueriesPlanValue
}

func (r *NavigatorRunResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *NavigatorRunResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	runs := uint32(1)
	setRuns(ctx, &resp.Diagnostics, resp.Private.SetKey, runs)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.SetField(ctx, "runs", runs)

	timeout, newDiags := terraformOperationResourceTimeout(ctx, terraformOpCreate, data.Timeouts, defaultNavigatorRunTimeout)
	resp.Diagnostics.Append(newDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	data.ID = types.StringValue(uuid.New().String())

	var navigatorRun navigatorRun
	resp.Diagnostics.Append(data.Value(ctx, &navigatorRun, r.opts, runs)...)

	if resp.Diagnostics.HasError() {
		return
	}

	run(ctx, &resp.Diagnostics, timeout, terraformOpCreate, &navigatorRun)
	resp.Diagnostics.Append(data.Set(ctx, navigatorRun)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NavigatorRunResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
}

func (r *NavigatorRunResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state *NavigatorRunResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer func() {
		if !resp.Diagnostics.HasError() {
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		}
	}()

	if !r.ShouldRun(data, state) {
		tflog.Debug(ctx, "skipping run")

		return
	}

	runs := incrementRuns(ctx, &resp.Diagnostics, req.Private.GetKey, resp.Private.SetKey)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.SetField(ctx, "runs", runs)

	timeout, newDiags := terraformOperationResourceTimeout(ctx, terraformOpUpdate, data.Timeouts, defaultNavigatorRunTimeout)
	resp.Diagnostics.Append(newDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var navigatorRun navigatorRun
	resp.Diagnostics.Append(data.Value(ctx, &navigatorRun, r.opts, runs)...)

	if resp.Diagnostics.HasError() {
		return
	}

	run(ctx, &resp.Diagnostics, timeout, terraformOpUpdate, &navigatorRun)
	resp.Diagnostics.Append(data.Set(ctx, navigatorRun)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *NavigatorRunResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *NavigatorRunResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !data.RunOnDestroy.ValueBool() {
		tflog.Debug(ctx, "skipping run, 'run_on_destroy' disabled")

		return
	}

	runs := incrementRuns(ctx, &resp.Diagnostics, req.Private.GetKey, resp.Private.SetKey)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.SetField(ctx, "runs", runs)

	timeout, newDiags := terraformOperationResourceTimeout(ctx, terraformOpDelete, data.Timeouts, defaultNavigatorRunTimeout)
	resp.Diagnostics.Append(newDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var navigatorRun navigatorRun
	resp.Diagnostics.Append(data.Value(ctx, &navigatorRun, r.opts, runs)...)

	if resp.Diagnostics.HasError() {
		return
	}

	run(ctx, &resp.Diagnostics, timeout, terraformOpDelete, &navigatorRun)
}
