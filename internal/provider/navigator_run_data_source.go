//nolint:dupl // surface plumbing, not schema, is what still overlaps
package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

var (
	_ datasource.DataSource              = (*NavigatorRunDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*NavigatorRunDataSource)(nil)
)

type NavigatorRunDataSourceModel struct {
	NavigatorRunCommonModel

	ArtifactQueries types.Map      `tfsdk:"artifact_queries"`
	ID              types.String   `tfsdk:"id"`
	Command         types.String   `tfsdk:"command"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
}

func (m NavigatorRunDataSourceModel) Value(ctx context.Context, opts *providerOptions, runData *navigatorRunData) diag.Diagnostics {
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

func (m *NavigatorRunDataSourceModel) Set(ctx context.Context, run navigatorRunData) diag.Diagnostics {
	return run.Store(ctx, &m.Command, &m.AnsibleOptions, &m.ArtifactQueries)
}

type NavigatorRunDataSource struct {
	opts *providerOptions
}

func NewNavigatorRunDataSource() datasource.DataSource { //nolint:ireturn
	return &NavigatorRunDataSource{}
}

func (d *NavigatorRunDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_navigator_run", req.ProviderTypeName)
}

func (d *NavigatorRunDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	description := navigatorRunDescription(surfaceDataSource)
	attributes := dataSourceAttributes(navigatorRunAttributes(surfaceDataSource))
	// TODO include defaultNavigatorRunTimeout in description
	attributes["timeouts"] = timeouts.Attributes(ctx)

	resp.Schema = schema.Schema{
		Description:         description.Description,
		MarkdownDescription: description.MarkdownDescription,
		Attributes:          attributes,
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

	ctx, cancel := context.WithTimeout(ctx, timeout+navigatorRunTimeoutOverhead)
	defer cancel()

	data.ID = types.StringValue(uuid.New().String())

	var runData navigatorRunData

	resp.Diagnostics.Append(data.Value(ctx, d.opts, &runData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	runData.operation = terraformOpRead
	runData.config.Settings.Timeout = timeout

	run(ctx, &resp.Diagnostics, &runData)
	resp.Diagnostics.Append(data.Set(ctx, runData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
