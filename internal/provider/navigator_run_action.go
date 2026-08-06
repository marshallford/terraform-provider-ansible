package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/action/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

var (
	_ action.Action              = (*NavigatorRunAction)(nil)
	_ action.ActionWithConfigure = (*NavigatorRunAction)(nil)
)

type NavigatorRunActionModel struct {
	NavigatorRunCommonModel

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (m NavigatorRunActionModel) Value(ctx context.Context, opts *providerOptions, runData *navigatorRunData) diag.Diagnostics {
	var diags diag.Diagnostics

	*runData = navigatorRunData{
		hostDir:    navigatorRunDirPath(opts.BaseRunDirectory, uuid.New().String(), 0),
		persistDir: opts.PersistRunDirectory,
	}

	diags.Append(runData.Load(ctx, m.NavigatorRunCommonModel)...)

	return diags
}

type NavigatorRunAction struct {
	opts *providerOptions
}

func NewNavigatorRunAction() action.Action { //nolint:ireturn
	return &NavigatorRunAction{}
}

func (a *NavigatorRunAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_navigator_run", req.ProviderTypeName)
}

func (a *NavigatorRunAction) Schema(ctx context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	description := navigatorRunDescription(surfaceAction)
	attributes := actionAttributes(navigatorRunAttributes(surfaceAction))
	// TODO include defaultNavigatorRunTimeout in description
	attributes["timeouts"] = timeouts.Attributes(ctx)

	resp.Schema = schema.Schema{
		Description:         description.Description,
		MarkdownDescription: description.MarkdownDescription,
		Attributes:          attributes,
	}
}

func (a *NavigatorRunAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	opts, ok := configureActionClient(req, resp)
	if !ok {
		return
	}

	a.opts = opts
}

func (a *NavigatorRunAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data *NavigatorRunActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	resp.Diagnostics.Append(data.SetDefaults(ctx)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, newDiags := terraformOperationActionTimeout(ctx, data.Timeouts, defaultNavigatorRunTimeout)
	resp.Diagnostics.Append(newDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout+navigatorRunTimeoutOverhead)
	defer cancel()

	var runData navigatorRunData

	resp.Diagnostics.Append(data.Value(ctx, a.opts, &runData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	runData.playbookArtifactQueries = map[string]ansible.PlaybookArtifactQuery{
		"stdout": {JQFilter: ".stdout[]", Raw: true},
	}

	runData.operation = terraformOpInvoke
	runData.config.Settings.Timeout = timeout

	run(ctx, &resp.Diagnostics, &runData)

	if resp.Diagnostics.HasError() {
		return
	}

	stdout := strings.Join(runData.playbookArtifactQueries["stdout"].Results, "\n")
	resp.SendProgress(action.InvokeProgressEvent{Message: stdout})
}
