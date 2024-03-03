package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

const (
	navigatorRunDestroyEnvVar          = "ANSIBLE_TF_DESTROY"
	navigatorRunDir                    = "tf-provider-ansible-navigator-run"
	defaultNavigatorRunTimeout         = 10 * time.Minute
	defaultNavigatorRunContainerEngine = "auto"
	defaultNavigatorRunImage           = "ghcr.io/ansible/creator-ee:latest"
	defaultNavigatorRunPullPolicy      = "tag"
	defaultNavigatorRunOnDestroy       = false
)

var _ resource.Resource = &NavigatorRunResource{}

func NewNavigatorRunResource() resource.Resource { //nolint:ireturn
	return &NavigatorRunResource{}
}

type NavigatorRunResource struct {
	opts *providerOptions
}

type NavigatorRunResourceModel struct {
	WorkingDirectory       types.String   `tfsdk:"working_directory"`
	Playbook               types.String   `tfsdk:"playbook"`
	Inventory              types.String   `tfsdk:"inventory"`
	ExecutionEnvironment   types.Object   `tfsdk:"execution_environment"`
	AnsibleNavigatorBinary types.String   `tfsdk:"ansible_navigator_binary"`
	AnsibleOptions         types.Object   `tfsdk:"ansible_options"`
	RunOnDestroy           types.Bool     `tfsdk:"run_on_destroy"`
	Triggers               types.Map      `tfsdk:"triggers"`
	ReplacementTriggers    types.Map      `tfsdk:"replacement_triggers"`
	ArtifactQueries        types.Map      `tfsdk:"artifact_queries"`
	ID                     types.String   `tfsdk:"id"`
	Command                types.String   `tfsdk:"command"`
	Timeouts               timeouts.Value `tfsdk:"timeouts"`
}

type ExecutionEnvironmentModel struct {
	ContainerEngine          types.String `tfsdk:"container_engine"`
	EnvironmentVariablesPass types.List   `tfsdk:"environment_variables_pass"`
	EnvironmentVariablesSet  types.Map    `tfsdk:"environment_variables_set"`
	Image                    types.String `tfsdk:"image"`
	PullArguments            types.List   `tfsdk:"pull_arguments"`
	PullPolicy               types.String `tfsdk:"pull_policy"`
}

type AnsibleOptionsModel struct {
	ForceHandlers types.Bool `tfsdk:"force_handlers"`
	Limit         types.List `tfsdk:"limit"`
	Tags          types.List `tfsdk:"tags"`
}

type ArtifactQueryModel struct {
	JSONPath types.String `tfsdk:"jsonpath"`
	Result   types.String `tfsdk:"result"`
}

func (ExecutionEnvironmentModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"container_engine":           types.StringType,
		"environment_variables_pass": types.ListType{ElemType: types.StringType},
		"environment_variables_set":  types.MapType{ElemType: types.StringType},
		"image":                      types.StringType,
		"pull_arguments":             types.ListType{ElemType: types.StringType},
		"pull_policy":                types.StringType,
	}
}

func (m ExecutionEnvironmentModel) Value(ctx context.Context, settings *ansible.NavigatorSettings) diag.Diagnostics {
	var diags diag.Diagnostics

	settings.ContainerEngine = m.ContainerEngine.ValueString()

	var envVarsPass []string
	if !m.EnvironmentVariablesPass.IsNull() {
		diags.Append(m.EnvironmentVariablesPass.ElementsAs(ctx, &envVarsPass, false)...)
	}

	settings.EnvironmentVariablesPass = envVarsPass

	envVarsSet := map[string]string{}
	if !m.EnvironmentVariablesSet.IsNull() {
		diags.Append(m.EnvironmentVariablesSet.ElementsAs(ctx, &envVarsSet, false)...)
	}

	settings.EnvironmentVariablesSet = envVarsSet

	settings.Image = m.Image.ValueString()

	var pullArguments []string
	if !m.PullArguments.IsNull() {
		diags.Append(m.PullArguments.ElementsAs(ctx, &pullArguments, false)...)
	}

	settings.PullArguments = pullArguments

	settings.PullPolicy = m.PullPolicy.ValueString()

	return diags
}

// func (AnsibleOptionsModel) attrTypes() map[string]attr.Type {
// 	return map[string]attr.Type{
// 		"force_handlers": types.BoolType,
// 		"limit":          types.ListType{ElemType: types.StringType},
// 		"tags":           types.ListType{ElemType: types.StringType},
// 	}
// }

func (m AnsibleOptionsModel) Value(ctx context.Context, opts *ansible.RunOptions) diag.Diagnostics {
	var diags diag.Diagnostics

	opts.ForceHandlers = m.ForceHandlers.ValueBool()

	var limit []string
	if !m.Limit.IsNull() {
		diags.Append(m.Limit.ElementsAs(ctx, &limit, false)...)
	}

	opts.Limit = limit

	var tags []string
	if !m.Tags.IsNull() {
		diags.Append(m.Tags.ElementsAs(ctx, &tags, false)...)
	}

	opts.Tags = tags

	return diags
}

func (ArtifactQueryModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"jsonpath": types.StringType,
		"result":   types.StringType,
	}
}

func (m ArtifactQueryModel) Value(ctx context.Context, query *ansible.ArtifactQuery) diag.Diagnostics {
	var diags diag.Diagnostics

	query.JSONPath = m.JSONPath.ValueString()
	query.Result = m.Result.ValueString()

	return diags
}

func (m *ArtifactQueryModel) Set(ctx context.Context, query ansible.ArtifactQuery) diag.Diagnostics {
	var diags diag.Diagnostics

	m.JSONPath = types.StringValue(query.JSONPath)
	m.Result = types.StringValue(query.Result)

	return diags
}

func (r *NavigatorRunResource) Run(ctx context.Context, diags *diag.Diagnostics, data *NavigatorRunResourceModel, destroy bool) { //nolint:cyclop
	var err error

	timeout, newDiags := data.Timeouts.Create(ctx, defaultNavigatorRunTimeout)
	diags.Append(newDiags...)

	if diags.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var eeModel ExecutionEnvironmentModel
	diags.Append(data.ExecutionEnvironment.As(ctx, &eeModel, basetypes.ObjectAsOptions{})...)

	var navigatorSettings ansible.NavigatorSettings
	diags.Append(eeModel.Value(ctx, &navigatorSettings)...)

	var optsModel AnsibleOptionsModel
	diags.Append(data.AnsibleOptions.As(ctx, &optsModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)

	var ansibleOptions ansible.RunOptions
	diags.Append(optsModel.Value(ctx, &ansibleOptions)...)

	var queriesModel map[string]ArtifactQueryModel
	diags.Append(data.ArtifactQueries.ElementsAs(ctx, &queriesModel, false)...)

	artifactQueries := map[string]ansible.ArtifactQuery{}
	for name, model := range queriesModel {
		var query ansible.ArtifactQuery

		diags.Append(model.Value(ctx, &query)...)
		artifactQueries[name] = query
	}

	if diags.HasError() {
		return
	}

	if destroy {
		navigatorSettings.EnvironmentVariablesSet[navigatorRunDestroyEnvVar] = "true"
	}

	navigatorSettingsContents, err := ansible.GenerateNavigatorSettings(&navigatorSettings)
	addError(diags, "Ansible navigator settings not generated", err)

	err = ansible.WorkingDirectoryPreflight(data.WorkingDirectory.ValueString())
	addPathError(diags, path.Root("working_directory"), "Working directory preflight check", err)

	err = ansible.ContainerEnginePreflight(navigatorSettings.ContainerEngine)
	addPathError(diags, path.Root("execution_environment").AtMapKey("container_engine"), "Container engine preflight check", err)

	ansibleNavigatorBinary := data.AnsibleNavigatorBinary.ValueString()
	if data.AnsibleNavigatorBinary.IsNull() {
		ansibleNavigatorBinary, err = ansible.NavigatorPath()
		addError(diags, "Ansible navigator not found", err)
	}

	err = ansible.NavigatorPreflight(ansibleNavigatorBinary)
	addPathError(diags, path.Root("ansible_navigator_binary"), "Ansible navigator preflight check", err)

	tempRunDir, err := ansible.CreateTempRunDir(r.opts.BaseRunDirectory, navigatorRunDir)
	addError(diags, "Temporary run directory not created", err)

	err = ansible.CreatePlaybookFile(tempRunDir, data.Playbook.ValueString())
	addError(diags, "Ansible playbook file not created", err)

	err = ansible.CreateInventoryFile(tempRunDir, data.Inventory.ValueString())
	addError(diags, "Ansible inventory file not created", err)

	err = ansible.CreateNavigatorSettingsFile(tempRunDir, navigatorSettingsContents)
	addError(diags, "Ansible navigator settings file not created", err)

	command := ansible.GenerateNavigatorRunCommand(
		ctx,
		data.WorkingDirectory.ValueString(),
		ansibleNavigatorBinary,
		tempRunDir,
		&ansibleOptions,
	)

	if diags.HasError() {
		if !r.opts.PersistRunDirectory {
			err = ansible.RemoveTempRunDir(tempRunDir)
			addError(diags, "Temporary run directory not removed", err)
		}

		return
	}

	data.Command = types.StringValue(command.String())

	output, err := ansible.ExecNavigatorRunCommand(command)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			addError(diags, "Ansible navigator run timed out", err)
		} else {
			addError(diags, "Ansible navigator run failed", fmt.Errorf("%w\n\nstdout and stderr:\n\n%s", err, output))
		}
	}

	// TODO stream to log file instead
	err = ansible.CreateNavigatorRunLogFile(tempRunDir, output)
	addError(diags, "Ansible navigator run output log not created", err)

	// TODO skip on destroy?
	err = ansible.QueryPlaybookArtifact(tempRunDir, artifactQueries)
	addPathError(diags, path.Root("artifact_queries"), "Playbook artifact queries failed", err)

	for name, model := range queriesModel {
		diags.Append(model.Set(ctx, artifactQueries[name])...)
		queriesModel[name] = model
	}

	newQueriesModel, newDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: ArtifactQueryModel{}.attrTypes()}, queriesModel)
	diags.Append(newDiags...)

	data.ArtifactQueries = newQueriesModel

	if !r.opts.PersistRunDirectory {
		err = ansible.RemoveTempRunDir(tempRunDir)
		addError(diags, "Temporary run directory not removed", err)
	}
}

func (r *NavigatorRunResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_navigator_run", req.ProviderTypeName)
}

func (r *NavigatorRunResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         fmt.Sprintf("Run an Ansible playbook within an Ansible execution environment (EE). Requires '%s' and a container engine to run the EE.", ansible.NavigatorProgram),
		MarkdownDescription: fmt.Sprintf("Run an Ansible playbook within an Ansible execution environment (EE). Requires `%s` and a container engine to run the EE.", ansible.NavigatorProgram),
		Attributes: map[string]schema.Attribute{
			// required
			"working_directory": schema.StringAttribute{
				Description:         fmt.Sprintf("Absolute directory which '%s' is run from. Recommended to be the root Ansible content directory (sometimes called the project directory), which is likely to contain 'ansible.cfg', 'roles/', etc.", ansible.NavigatorProgram),
				MarkdownDescription: fmt.Sprintf("Absolute directory which `%s` is run from. Recommended to be the root Ansible [content directory](https://docs.ansible.com/ansible/latest/tips_tricks/sample_setup.html#sample-directory-layout) (sometimes called the project directory), which is likely to contain `ansible.cfg`, `roles/`, etc.", ansible.NavigatorProgram),
				Required:            true,
				Validators: []validator.String{
					stringIsAbsolutePath(),
				},
			},
			"playbook": schema.StringAttribute{
				Description:         "Ansible playbook contents.",
				MarkdownDescription: "Ansible [playbook](https://docs.ansible.com/ansible/latest/playbook_guide/playbooks_intro.html) contents.",
				Required:            true,
			},
			"inventory": schema.StringAttribute{
				Description:         "Ansible inventory contents.",
				MarkdownDescription: "Ansible [inventory](https://docs.ansible.com/ansible/latest/getting_started/get_started_inventory.html) contents.",
				Required:            true,
			},
			// optional
			"execution_environment": schema.SingleNestedAttribute{
				Description:         "Execution environment related configuration.",
				MarkdownDescription: "[Execution environment](https://ansible.readthedocs.io/en/latest/getting_started_ee/index.html) related configuration.",
				Optional:            true,
				Computed:            true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					ExecutionEnvironmentModel{}.attrTypes(),
					map[string]attr.Value{
						"container_engine":           types.StringValue(defaultNavigatorRunContainerEngine),
						"environment_variables_pass": types.ListNull(types.StringType),
						"environment_variables_set":  types.MapNull(types.StringType),
						"image":                      types.StringValue(defaultNavigatorRunImage),
						"pull_arguments":             types.ListNull(types.StringType),
						"pull_policy":                types.StringValue(defaultNavigatorRunPullPolicy),
					},
				)),
				Attributes: map[string]schema.Attribute{
					"container_engine": schema.StringAttribute{
						Description:         fmt.Sprintf("Container engine responsible for running the execution environment container image. Options: %s. Defaults to '%s'.", strings.Join(wrapElements(ansible.ContainerEngineOptions(), "'"), ", "), defaultNavigatorRunContainerEngine),
						MarkdownDescription: fmt.Sprintf("[Container engine](https://ansible.readthedocs.io/projects/navigator/settings/#container-engine) responsible for running the execution environment container image. Options: %s. Defaults to `%s`.", strings.Join(wrapElements(ansible.ContainerEngineOptions(), "`"), ", "), defaultNavigatorRunContainerEngine),
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(defaultNavigatorRunContainerEngine),
						Validators: []validator.String{
							stringvalidator.OneOf(ansible.ContainerEngineOptions()...),
						},
					},
					"environment_variables_pass": schema.ListAttribute{
						Description:         "Existing environment variables to be passed through to and set within the execution environment.",
						MarkdownDescription: "Existing environment variables to be [passed](https://ansible.readthedocs.io/projects/navigator/settings/#pass-environment-variable) through to and set within the execution environment.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"environment_variables_set": schema.MapAttribute{
						Description:         "Environment variables to be set within the execution environment.",
						MarkdownDescription: "Environment variables to be [set](https://ansible.readthedocs.io/projects/navigator/settings/#set-environment-variable) within the execution environment.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"image": schema.StringAttribute{
						Description:         fmt.Sprintf("Name of the execution environment container image. Defaults to '%s'.", defaultNavigatorRunImage),
						MarkdownDescription: fmt.Sprintf("Name of the execution environment container [image](https://ansible.readthedocs.io/projects/navigator/settings/#execution-environment-image). Defaults to `%s`.", defaultNavigatorRunImage),
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(defaultNavigatorRunImage),
					},
					"pull_arguments": schema.ListAttribute{
						Description:         "Additional parameters that should be added to the pull command when pulling an execution environment container image from a container registry.",
						MarkdownDescription: "Additional [parameters](https://ansible.readthedocs.io/projects/navigator/settings/#pull-arguments) that should be added to the pull command when pulling an execution environment container image from a container registry.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"pull_policy": schema.StringAttribute{
						Description:         fmt.Sprintf("Container image pull policy. Defaults to '%s'.", defaultNavigatorRunPullPolicy),
						MarkdownDescription: fmt.Sprintf("Container image [pull policy](https://ansible.readthedocs.io/projects/navigator/settings/#pull-policy). Defaults to `%s`.", defaultNavigatorRunPullPolicy),
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(defaultNavigatorRunPullPolicy),
						Validators: []validator.String{
							stringvalidator.OneOf(ansible.PullPolicyOptions()...),
						},
					},
					// TODO volume-mounts, container-options
				},
			},
			"ansible_navigator_binary": schema.StringAttribute{
				Description:         fmt.Sprintf("Absolute path to '%s' binary. By default '$PATH' is searched for the binary.", ansible.NavigatorProgram),
				MarkdownDescription: fmt.Sprintf("Absolute path to `%s` binary. By default `$PATH` is searched for the binary.", ansible.NavigatorProgram),
				Optional:            true,
				Validators: []validator.String{
					stringIsAbsolutePath(),
				},
			},
			"ansible_options": schema.SingleNestedAttribute{
				Description:         "Ansible playbook run related configuration.",
				MarkdownDescription: "Ansible [playbook](https://docs.ansible.com/ansible/latest/cli/ansible-playbook.html) run related configuration.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"force_handlers": schema.BoolAttribute{
						Description: "Run handlers even if a task fails.",
						Optional:    true,
					},
					"limit": schema.ListAttribute{
						Description: "Further limit selected hosts to an additional pattern.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"tags": schema.ListAttribute{
						Description: "Only run plays and tasks tagged with these values.",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"run_on_destroy": schema.BoolAttribute{
				Description:         fmt.Sprintf("Run playbook on destroy. The environment variable '%s' is set during the destroy run to allow for conditional plays, tasks, etc. Defaults to '%t'.", navigatorRunDestroyEnvVar, defaultNavigatorRunOnDestroy),
				MarkdownDescription: fmt.Sprintf("Run playbook on destroy. The environment variable `%s` is set during the destroy run to allow for conditional plays, tasks, etc. Defaults to `%t`.", navigatorRunDestroyEnvVar, defaultNavigatorRunOnDestroy),
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(defaultNavigatorRunOnDestroy),
			},
			"triggers": schema.MapAttribute{
				Description: "Arbitrary map of values that, when changed, will run the playbook again. Serves as alternative way to trigger a run without changing the inventory or playbook.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"replacement_triggers": schema.MapAttribute{
				Description:         "Arbitrary map of values that, when changed, will recreate the resource. Similar to 'triggers', but will cause 'id' to change. Useful when combined with 'run_on_destroy'.",
				MarkdownDescription: "Arbitrary map of values that, when changed, will recreate the resource. Similar to `triggers`, but will cause `id` to change. Useful when combined with `run_on_destroy`.",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"artifact_queries": schema.MapNestedAttribute{
				Description:         "Query the playbook artifact with JSONPath. The playbook artifact contains detailed information about every play and task, as well as the stdout from the playbook run.",
				MarkdownDescription: "Query the playbook artifact with [JSONPath](https://goessner.net/articles/JsonPath/). The playbook artifact contains detailed information about every play and task, as well as the stdout from the playbook run.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"jsonpath": schema.StringAttribute{
							Description: "JSONPath expression.",
							Required:    true,
						},
						"result": schema.StringAttribute{
							Description: "Result of the query. Result may be empty if a field or map key cannot be located.",
							Computed:    true,
						},
					},
				},
			},
			// computed
			"id": schema.StringAttribute{
				Description: "UUID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"command": schema.StringAttribute{
				Description:         fmt.Sprintf("Generated '%s' run command. Useful for troubleshooting.", ansible.NavigatorProgram),
				MarkdownDescription: fmt.Sprintf("Generated `%s` run command. Useful for troubleshooting.", ansible.NavigatorProgram),
				Computed:            true,
			},
			// timeouts
			// TODO include defaultNavigatorRunTimeout in description
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *NavigatorRunResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	opts, ok := configureResourceClient(req, resp)
	if !ok {
		return
	}

	r.opts = opts
}

func (r *NavigatorRunResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *NavigatorRunResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(uuid.New().String())
	r.Run(ctx, &resp.Diagnostics, data, false)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NavigatorRunResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *NavigatorRunResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *NavigatorRunResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.Run(ctx, &resp.Diagnostics, data, false)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NavigatorRunResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *NavigatorRunResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.RunOnDestroy.ValueBool() {
		r.Run(ctx, &resp.Diagnostics, data, true)
	}
}
