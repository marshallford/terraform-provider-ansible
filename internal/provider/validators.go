package provider

import (
	"context"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type stringIsAbsolutePathValidator struct{}

func (v stringIsAbsolutePathValidator) Description(ctx context.Context) string {
	return "string must be an absolute path"
}

func (v stringIsAbsolutePathValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsAbsolutePathValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	if !filepath.IsAbs(req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Not an absolute path",
			"String must be an absolute path",
		)

		return
	}
}

func stringIsAbsolutePath() stringIsAbsolutePathValidator {
	return stringIsAbsolutePathValidator{}
}
