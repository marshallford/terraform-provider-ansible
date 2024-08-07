package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	dataSourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const (
	terraformOperationCreate = iota
	terraformOperationRead   = iota
	terraformOperationUpdate = iota
	terraformOperationDelete = iota
	diagDetailPrefix         = "Underlying error details"
)

type providerOptions struct {
	BaseRunDirectory    string
	PersistRunDirectory bool
}

type terraformOperation int

var terraformOperations = []string{"create", "read", "update", "delete"} //nolint:gochecknoglobals

func (op terraformOperation) String() string {
	return terraformOperations[op]
}

func terraformOperationResourceTimeout(ctx context.Context, operation terraformOperation, value resourceTimeouts.Value, defaultTimeout time.Duration) (time.Duration, diag.Diagnostics) {
	switch operation {
	case terraformOperationCreate:
		return value.Create(ctx, defaultTimeout)
	case terraformOperationRead:
		return value.Read(ctx, defaultTimeout)
	case terraformOperationUpdate:
		return value.Update(ctx, defaultTimeout)
	case terraformOperationDelete:
		return value.Delete(ctx, defaultTimeout)
	default:
		return defaultTimeout, nil
	}
}

func terraformOperationDataSourceTimeout(ctx context.Context, value dataSourceTimeouts.Value, defaultTimeout time.Duration) (time.Duration, diag.Diagnostics) {
	return value.Read(ctx, defaultTimeout)
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

func addError(diags *diag.Diagnostics, summary string, err error) bool {
	if err != nil {
		diags.AddError(summary, fmt.Sprintf("%s: %s", diagDetailPrefix, err))

		return true
	}

	return false
}

func addPathError(diags *diag.Diagnostics, path path.Path, summary string, err error) bool {
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

func remove[T comparable](l []T, item T) []T {
	out := make([]T, 0)
	for _, element := range l {
		if element != item {
			out = append(out, element)
		}
	}

	return out
}
