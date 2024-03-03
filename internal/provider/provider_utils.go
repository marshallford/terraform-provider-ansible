package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type providerOptions struct {
	BaseRunDirectory    string
	PersistRunDirectory bool
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

// func configureDataSourceClient(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) (*providerOptions, bool) {
// 	if req.ProviderData == nil {
// 		return nil, false
// 	}

// 	opts, ok := req.ProviderData.(*providerOptions)

// 	if !ok {
// 		summary, detail := unexpectedConfigureType("Data Source", req.ProviderData)
// 		resp.Diagnostics.AddError(summary, detail)
// 	}

// 	return opts, ok
// }

func configureResourceClient(req resource.ConfigureRequest, resp *resource.ConfigureResponse) (*providerOptions, bool) {
	if req.ProviderData == nil {
		return nil, false
	}

	opts, ok := req.ProviderData.(*providerOptions) //nolint:varnamelen

	if !ok {
		summary, detail := unexpectedConfigureType("Resource", req.ProviderData)
		resp.Diagnostics.AddError(summary, detail)
	}

	return opts, ok
}

func addError(diags *diag.Diagnostics, summary string, err error) bool { //nolint:unparam
	if err != nil {
		diags.AddError(summary, fmt.Sprintf("unexpected error: %v", err))

		return true
	}

	return false
}

func addPathError(diags *diag.Diagnostics, path path.Path, summary string, err error) bool { //nolint:unparam
	if err != nil {
		diags.AddAttributeError(path, summary, err.Error())

		return true
	}

	return false
}

func wrapElements(input []string, wrap string) []string {
	output := []string{}
	for _, element := range input {
		output = append(output, fmt.Sprintf("%s%s%s", wrap, element, wrap))
	}

	return output
}
