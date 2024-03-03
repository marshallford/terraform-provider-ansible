// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const defaultProviderPersistRunDir = false

var _ provider.Provider = &AnsibleProvider{}

type AnsibleProvider struct {
	version string
}

type AnsibleProviderModel struct {
	BaseRunDirectory    types.String `tfsdk:"base_run_directory"`
	PersistRunDirectory types.Bool   `tfsdk:"persist_run_directory"`
}

func (p *AnsibleProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ansible"
	resp.Version = p.version
}

func (p *AnsibleProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Interact with Ansible.",
		MarkdownDescription: "Interact with [Ansible](https://github.com/ansible/ansible).",
		Attributes: map[string]schema.Attribute{
			"base_run_directory": schema.StringAttribute{
				Description:         "Base directory in which to create temporary run directories. On Unix systems this defaults to '$TMPDIR' if non-empty, else '/tmp'.",
				MarkdownDescription: "Base directory in which to create temporary run directories. On Unix systems this defaults to `$TMPDIR` if non-empty, else `/tmp`.",
				Optional:            true,
			},
			"persist_run_directory": schema.BoolAttribute{
				Description:         fmt.Sprintf("Remove temporary run directory after the run completes. Useful when troubleshooting. Defaults to '%t'.", defaultProviderPersistRunDir),
				MarkdownDescription: fmt.Sprintf("Remove temporary run directory after the run completes. Useful when troubleshooting. Defaults to `%t`.", defaultProviderPersistRunDir),
				Optional:            true,
			},
		},
	}
}

func (p *AnsibleProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AnsibleProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.BaseRunDirectory.IsUnknown() {
		path := path.Root("base_run_directory")
		summary, detail := unknownProviderValue(path)
		resp.Diagnostics.AddAttributeError(path, summary, detail)
	}

	if data.PersistRunDirectory.IsUnknown() {
		path := path.Root("persist_run_directory")
		summary, detail := unknownProviderValue(path)
		resp.Diagnostics.AddAttributeError(path, summary, detail)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	var opts providerOptions

	if data.BaseRunDirectory.IsNull() {
		opts.BaseRunDirectory = os.TempDir()
	} else {
		baseRunDirectory := data.BaseRunDirectory.ValueString()
		if !filepath.IsAbs(baseRunDirectory) {
			resp.Diagnostics.AddAttributeError(
				path.Root("base_run_directory"),
				"Base run directory must be an absolute path",
				fmt.Sprintf("%s is not an absolute path", baseRunDirectory),
			)

			return
		}
		opts.BaseRunDirectory = baseRunDirectory
	}

	if data.PersistRunDirectory.IsNull() {
		opts.PersistRunDirectory = defaultProviderPersistRunDir
	} else {
		opts.PersistRunDirectory = data.PersistRunDirectory.ValueBool()
	}

	resp.DataSourceData = &opts
	resp.ResourceData = &opts
}

func (p *AnsibleProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewNavigatorRunResource,
	}
}

func (p *AnsibleProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AnsibleProvider{
			version: version,
		}
	}
}
