//nolint:dupl // surface plumbing, not schema, is what still overlaps
package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/ephemeral/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

var (
	_ ephemeral.EphemeralResource              = (*NavigatorRunEphemeralResource)(nil)
	_ ephemeral.EphemeralResourceWithConfigure = (*NavigatorRunEphemeralResource)(nil)
)

type NavigatorRunEphemeralResourceModel struct {
	NavigatorRunCommonModel

	ArtifactQueries types.Map      `tfsdk:"artifact_queries"`
	ID              types.String   `tfsdk:"id"`
	Command         types.String   `tfsdk:"command"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
}

func (m NavigatorRunEphemeralResourceModel) Value(ctx context.Context, opts *providerOptions, runData *navigatorRunData) diag.Diagnostics {
	var diags diag.Diagnostics

	*runData = navigatorRunData{
		hostDir:    navigatorRunDirPath(opts.BaseRunDirectory, m.ID.ValueString(), 0),
		persistDir: opts.PersistRunDirectory,
	}

	diags.Append(runData.Load(ctx, m.NavigatorRunCommonModel)...)

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

func (m *NavigatorRunEphemeralResourceModel) Set(ctx context.Context, run navigatorRunData) diag.Diagnostics {
	return run.Store(ctx, &m.Command, &m.AnsibleOptions, &m.ArtifactQueries)
}

type NavigatorRunEphemeralResource struct {
	opts *providerOptions
}

func NewNavigatorRunEphemeralResource() ephemeral.EphemeralResource { //nolint:ireturn
	return &NavigatorRunEphemeralResource{}
}

func (er *NavigatorRunEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_navigator_run", req.ProviderTypeName)
}

func (er *NavigatorRunEphemeralResource) Schema(ctx context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	description := navigatorRunDescription(surfaceEphemeral)
	attributes := ephemeralAttributes(navigatorRunAttributes(surfaceEphemeral))
	// TODO include defaultNavigatorRunTimeout in description
	attributes["timeouts"] = timeouts.Attributes(ctx)

	resp.Schema = schema.Schema{
		Description:         description.Description,
		MarkdownDescription: description.MarkdownDescription,
		Attributes:          attributes,
	}
}

func (er *NavigatorRunEphemeralResource) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	opts, ok := configureEphemeralResourceClient(req, resp)
	if !ok {
		return
	}

	er.opts = opts
}

func (er *NavigatorRunEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data *NavigatorRunEphemeralResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	resp.Diagnostics.Append(data.SetDefaults(ctx)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, newDiags := terraformOperationEphemeralResourceTimeout(ctx, data.Timeouts, defaultNavigatorRunTimeout)
	resp.Diagnostics.Append(newDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout+navigatorRunTimeoutOverhead)
	defer cancel()

	data.ID = types.StringValue(uuid.New().String())

	var runData navigatorRunData

	resp.Diagnostics.Append(data.Value(ctx, er.opts, &runData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	runData.operation = terraformOpOpen
	runData.config.Settings.Timeout = timeout

	run(ctx, &resp.Diagnostics, &runData)
	resp.Diagnostics.Append(data.Set(ctx, runData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
