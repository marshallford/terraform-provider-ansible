package provider

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

var (
	_ resource.Resource               = (*NavigatorRunResource)(nil)
	_ resource.ResourceWithConfigure  = (*NavigatorRunResource)(nil)
	_ resource.ResourceWithModifyPlan = (*NavigatorRunResource)(nil)
)

type NavigatorRunResourceModel struct {
	NavigatorRunCommonModel

	RunOnDestroy    types.Bool     `tfsdk:"run_on_destroy"`
	DestroyPlaybook types.String   `tfsdk:"destroy_playbook"`
	Triggers        types.Object   `tfsdk:"triggers"`
	ArtifactQueries types.Map      `tfsdk:"artifact_queries"`
	ID              types.String   `tfsdk:"id"`
	Command         types.String   `tfsdk:"command"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
}

func (m NavigatorRunResourceModel) Value(ctx context.Context, destroy bool, opts *providerOptions, runs uint32, previousInventory *string, runData *navigatorRunData) diag.Diagnostics {
	var diags diag.Diagnostics

	*runData = navigatorRunData{
		hostDir:    navigatorRunDirPath(opts.BaseRunDirectory, m.ID.ValueString(), runs),
		persistDir: opts.PersistRunDirectory,
	}

	diags.Append(runData.Load(ctx, m.NavigatorRunCommonModel)...)

	if destroy && !m.DestroyPlaybook.IsNull() {
		runData.config.Playbook = m.DestroyPlaybook.ValueString()
	}

	if previousInventory != nil {
		runData.config.Inventories = append(runData.config.Inventories, ansible.Inventory{Name: navigatorRunPrevInventoryName, Contents: *previousInventory, Exclude: true})
	}

	var queriesModel map[string]ArtifactQueryModel
	diags.Append(m.ArtifactQueries.ElementsAs(ctx, &queriesModel, false)...)

	runData.userArtifactQueries = true
	runData.playbookArtifactQueries = map[string]ansible.PlaybookArtifactQuery{}
	for name, model := range queriesModel {
		var query ansible.PlaybookArtifactQuery

		diags.Append(model.Value(ctx, &query)...)
		runData.playbookArtifactQueries[name] = query
	}

	return diags
}

func (m *NavigatorRunResourceModel) Set(ctx context.Context, run navigatorRunData) diag.Diagnostics {
	return run.Store(ctx, &m.Command, &m.AnsibleOptions, &m.ArtifactQueries)
}

func (m *NavigatorRunResourceModel) Trigger(name string) attr.Value { //nolint:ireturn
	if m.Triggers.IsNull() {
		return types.DynamicNull()
	}

	return m.Triggers.Attributes()[name]
}

func (m *NavigatorRunResourceModel) ShouldRun(state *NavigatorRunResourceModel) bool {
	if !m.Trigger("exclusive_run").IsNull() {
		return !m.Trigger("exclusive_run").Equal(state.Trigger("exclusive_run"))
	}

	// skip working_directory, ansible_navigator_binary, run_on_destroy, destroy_playbook, timeouts
	unchanged := []bool{
		m.Playbook.Equal(state.Playbook),
		m.Inventory.Equal(state.Inventory),
		m.ExecutionEnvironment.Equal(state.ExecutionEnvironment),
		m.AnsibleOptions.Equal(state.AnsibleOptions),
		m.Timezone.Equal(state.Timezone),
		m.Trigger("run").Equal(state.Trigger("run")),
		m.ArtifactQueries.Equal(state.ArtifactQueries),
	}

	return slices.Contains(unchanged, false)
}

type NavigatorRunResource struct {
	opts *providerOptions
}

func NewNavigatorRunResource() resource.Resource { //nolint:ireturn
	return &NavigatorRunResource{}
}

func (r *NavigatorRunResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_navigator_run", req.ProviderTypeName)
}

func (r *NavigatorRunResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	description := navigatorRunDescription(surfaceResource)
	attributes := navigatorRunAttributes(surfaceResource)
	// TODO include defaultNavigatorRunTimeout in description
	attributes["timeouts"] = timeouts.Attributes(ctx, timeouts.Opts{
		Create: true,
		Update: true,
		Delete: true,
	})

	resp.Schema = schema.Schema{
		Description:         description.Description,
		MarkdownDescription: description.MarkdownDescription,
		Attributes:          attributes,
	}
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

	if optsPlanModel.KnownHosts.IsUnknown() && data.Trigger("known_hosts").Equal(state.Trigger("known_hosts")) {
		tflog.Trace(ctx, "keeping known hosts from state")

		optsPlanModel.KnownHosts = optsStateModel.KnownHosts
	}

	optsPlanValue, newDiags := types.ObjectValueFrom(ctx, AnsibleOptionsModel{}.AttrTypes(), optsPlanModel)
	resp.Diagnostics.Append(newDiags...)
	data.AnsibleOptions = optsPlanValue

	if !data.ShouldRun(state) {
		tflog.Debug(ctx, "planning no run", map[string]any{"reason": "no changes to run for"})

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

	ctx = tflog.SetField(ctx, "runs", runs)

	timeout, newDiags := terraformOperationResourceTimeout(ctx, terraformOpCreate, data.Timeouts, defaultNavigatorRunTimeout)
	resp.Diagnostics.Append(newDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout+navigatorRunTimeoutOverhead)
	defer cancel()

	data.ID = types.StringValue(uuid.New().String())

	var runData navigatorRunData

	resp.Diagnostics.Append(data.Value(ctx, false, r.opts, runs, nil, &runData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	runData.operation = terraformOpCreate
	runData.config.Settings.Timeout = timeout

	run(ctx, &resp.Diagnostics, &runData)
	resp.Diagnostics.Append(data.Set(ctx, runData)...)

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

	if !data.ShouldRun(state) {
		tflog.Debug(ctx, "skipping run", map[string]any{"reason": "no changes to run for"})

		return
	}

	runs := incrementRuns(ctx, &resp.Diagnostics, req.Private.GetKey, resp.Private.SetKey)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "runs", runs)

	timeout, newDiags := terraformOperationResourceTimeout(ctx, terraformOpUpdate, data.Timeouts, defaultNavigatorRunTimeout)
	resp.Diagnostics.Append(newDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout+navigatorRunTimeoutOverhead)
	defer cancel()

	var runData navigatorRunData

	resp.Diagnostics.Append(data.Value(ctx, false, r.opts, runs, state.Inventory.ValueStringPointer(), &runData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	runData.operation = terraformOpUpdate
	runData.config.Settings.Timeout = timeout

	run(ctx, &resp.Diagnostics, &runData)
	resp.Diagnostics.Append(data.Set(ctx, runData)...)

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
		tflog.Debug(ctx, "skipping run", map[string]any{"reason": "run_on_destroy disabled"})

		return
	}

	runs := incrementRuns(ctx, &resp.Diagnostics, req.Private.GetKey, resp.Private.SetKey)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "runs", runs)

	timeout, newDiags := terraformOperationResourceTimeout(ctx, terraformOpDelete, data.Timeouts, defaultNavigatorRunTimeout)
	resp.Diagnostics.Append(newDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout+navigatorRunTimeoutOverhead)
	defer cancel()

	var runData navigatorRunData

	resp.Diagnostics.Append(data.Value(ctx, true, r.opts, runs, nil, &runData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	runData.operation = terraformOpDelete
	runData.config.Settings.Timeout = timeout

	run(ctx, &resp.Diagnostics, &runData)
}
