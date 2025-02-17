package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	dataSourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	ephemeralResourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/ephemeral/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const (
	terraformOpCreate = iota
	terraformOpRead   = iota
	terraformOpUpdate = iota
	terraformOpDelete = iota
	terraformOpOpen   = iota
	diagDetailPrefix  = "Underlying error details"
)

type attrDescription struct {
	Description         string
	MarkdownDescription string
}

type providerOptions struct {
	BaseRunDirectory    string
	PersistRunDirectory bool
}

type (
	terraformOp  int
	terraformOps []terraformOp
)

var terraformOpNames = []string{"create", "read", "update", "delete", "open"} //nolint:gochecknoglobals

func (op terraformOp) String() string {
	return terraformOpNames[op]
}

func (ops terraformOps) Strings() []string {
	output := make([]string, 0, len(ops))
	for _, element := range ops {
		output = append(output, element.String())
	}

	return output
}

func terraformOperationResourceTimeout(ctx context.Context, op terraformOp, value resourceTimeouts.Value, defaultTimeout time.Duration) (time.Duration, diag.Diagnostics) {
	switch op {
	case terraformOpCreate:
		return value.Create(ctx, defaultTimeout)
	case terraformOpRead:
		return value.Read(ctx, defaultTimeout)
	case terraformOpUpdate:
		return value.Update(ctx, defaultTimeout)
	case terraformOpDelete:
		return value.Delete(ctx, defaultTimeout)
	default:
		return defaultTimeout, nil
	}
}

func terraformOperationDataSourceTimeout(ctx context.Context, value dataSourceTimeouts.Value, defaultTimeout time.Duration) (time.Duration, diag.Diagnostics) {
	return value.Read(ctx, defaultTimeout)
}

func terraformOperationEphemeralResourceTimeout(ctx context.Context, value ephemeralResourceTimeouts.Value, defaultTimeout time.Duration) (time.Duration, diag.Diagnostics) {
	return value.Open(ctx, defaultTimeout)
}

func unknownProviderValue(value path.Path) (string, string) {
	return fmt.Sprintf("Unknown configuration value '%s'", value),
		fmt.Sprintf("The provider cannot be configured as there is an unknown configuration value for '%s'. ", value) +
			"Either target apply the source of the value first or set the value statically in the configuration."
}

func unexpectedConfigureType(value string, providerData any) (string, string) {
	return fmt.Sprintf("Unexpected %s Configure Type", value),
		fmt.Sprintf("Expected *providerOptions, got: %T. Please report this issue to the provider developers.", providerData)
}

func configureResourceClient(req resource.ConfigureRequest, resp *resource.ConfigureResponse) (*providerOptions, bool) {
	if req.ProviderData == nil {
		return nil, false
	}

	opts, ok := req.ProviderData.(*providerOptions)

	if !ok {
		summary, detail := unexpectedConfigureType("Resource", req.ProviderData)
		resp.Diagnostics.AddError(summary, detail)
	}

	return opts, ok
}

func configureDataSourceClient(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) (*providerOptions, bool) {
	if req.ProviderData == nil {
		return nil, false
	}

	opts, ok := req.ProviderData.(*providerOptions)

	if !ok {
		summary, detail := unexpectedConfigureType("Data Source", req.ProviderData)
		resp.Diagnostics.AddError(summary, detail)
	}

	return opts, ok
}

func configureEphemeralResourceClient(req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) (*providerOptions, bool) {
	if req.ProviderData == nil {
		return nil, false
	}

	opts, ok := req.ProviderData.(*providerOptions)

	if !ok {
		summary, detail := unexpectedConfigureType("Ephemeral Resource", req.ProviderData)
		resp.Diagnostics.AddError(summary, detail)
	}

	return opts, ok
}

func addError(diags *diag.Diagnostics, summary string, err error) bool {
	if err != nil {
		diags.AddError(summary, fmt.Sprintf("%s: %s", diagDetailPrefix, err))

		return true
	}

	return false
}

func addPathError(diags *diag.Diagnostics, path path.Path, summary string, err error) bool { //nolint:unparam
	if err != nil {
		diags.AddAttributeError(path, summary, fmt.Sprintf("%s: %s", diagDetailPrefix, err))

		return true
	}

	return false
}

func addWarning(diags *diag.Diagnostics, summary string, err error) bool { //nolint:unparam
	if err != nil {
		diags.AddWarning(summary, fmt.Sprintf("%s: %s", diagDetailPrefix, err))

		return true
	}

	return false
}

func wrapElements(input []string, wrap string) []string {
	output := make([]string, 0, len(input))
	for _, element := range input {
		output = append(output, fmt.Sprintf("%s%s%s", wrap, element, wrap))
	}

	return output
}

func wrapElementsJoin(input []string, wrap string) string {
	return strings.Join(wrapElements(input, wrap), ", ")
}
