package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
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

func (p *AnsibleProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ansible"
	resp.Version = p.version
}

func (p *AnsibleProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Run Ansible playbooks.",
		MarkdownDescription: "Run [Ansible](https://github.com/ansible/ansible) playbooks.",
		Attributes: map[string]schema.Attribute{
			"base_run_directory": schema.StringAttribute{
				Description:         "Base directory in which to create run directories. On Unix systems this defaults to '$TMPDIR' if non-empty, else '/tmp'.",
				MarkdownDescription: "Base directory in which to create run directories. On Unix systems this defaults to `$TMPDIR` if non-empty, else `/tmp`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"persist_run_directory": schema.BoolAttribute{
				Description:         fmt.Sprintf("Remove run directory after the run completes. Useful when troubleshooting. Defaults to '%t'.", defaultProviderPersistRunDir),
				MarkdownDescription: fmt.Sprintf("Remove run directory after the run completes. Useful when troubleshooting. Defaults to `%t`.", defaultProviderPersistRunDir),
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

	opts := providerOptions{
		BaseRunDirectory:    os.TempDir(),
		PersistRunDirectory: defaultProviderPersistRunDir,
	}

	if !data.BaseRunDirectory.IsNull() {
		opts.BaseRunDirectory = data.BaseRunDirectory.ValueString()
	}

	err := ansible.DirectoryPreflight(opts.BaseRunDirectory)
	addPathError(&resp.Diagnostics, path.Root("base_run_directory"), "Base run directory preflight check", err)

	if !data.PersistRunDirectory.IsNull() {
		opts.PersistRunDirectory = data.PersistRunDirectory.ValueBool()
	}

	resp.ResourceData = &opts
	resp.DataSourceData = &opts
	resp.EphemeralResourceData = &opts
}

func (p *AnsibleProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewNavigatorRunResource,
	}
}

func (p *AnsibleProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewNavigatorRunDataSource,
	}
}

func (p *AnsibleProvider) EphemeralResources(_ context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		NewNavigatorRunEphemeralResource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AnsibleProvider{
			version: version,
		}
	}
}
