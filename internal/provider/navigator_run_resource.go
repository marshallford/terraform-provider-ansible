package provider

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

const (
	navigatorRunOperationEnvVar        = "ANSIBLE_TF_OPERATION"
	navigatorRunDir                    = "tf-ansible-navigator-run"
	defaultNavigatorRunTimeout         = 10 * time.Minute
	defaultNavigatorRunContainerEngine = "auto"
	defaultNavigatorRunImage           = "ghcr.io/ansible/community-ansible-dev-tools:v24.7.0"
	defaultNavigatorRunPullPolicy      = "tag"
	defaultNavigatorRunTimezone        = "UTC"
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
	Playbook               types.String   `tfsdk:"playbook"`
	Inventory              types.String   `tfsdk:"inventory"`
	WorkingDirectory       types.String   `tfsdk:"working_directory"`
	ExecutionEnvironment   types.Object   `tfsdk:"execution_environment"`
	AnsibleNavigatorBinary types.String   `tfsdk:"ansible_navigator_binary"`
	AnsibleOptions         types.Object   `tfsdk:"ansible_options"`
	Timezone               types.String   `tfsdk:"timezone"`
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
	ContainerOptions         types.List   `tfsdk:"container_options"`
}

type AnsibleOptionsModel struct {
	ForceHandlers types.Bool   `tfsdk:"force_handlers"`
	SkipTags      types.List   `tfsdk:"skip_tags"`
	StartAtTask   types.String `tfsdk:"start_at_task"`
	Limit         types.List   `tfsdk:"limit"`
	Tags          types.List   `tfsdk:"tags"`
	PrivateKeys   types.List   `tfsdk:"private_keys"`
}

type PrivateKeyModel struct {
	Name types.String `tfsdk:"name"`
	Data types.String `tfsdk:"data"`
}

type ArtifactQueryModel struct {
	JSONPath types.String `tfsdk:"jsonpath"`
	Result   types.String `tfsdk:"result"`
}

func (ExecutionEnvironmentModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"container_engine":           types.StringType,
		"environment_variables_pass": types.ListType{ElemType: types.StringType},
		"environment_variables_set":  types.MapType{ElemType: types.StringType},
		"image":                      types.StringType,
		"pull_arguments":             types.ListType{ElemType: types.StringType},
		"pull_policy":                types.StringType,
		"container_options":          types.ListType{ElemType: types.StringType},
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

	var containerOptions []string
	if !m.ContainerOptions.IsNull() {
		diags.Append(m.ContainerOptions.ElementsAs(ctx, &containerOptions, false)...)
	}
	settings.ContainerOptions = containerOptions

	return diags
}

func (m AnsibleOptionsModel) Value(ctx context.Context, opts *ansible.RunOptions) diag.Diagnostics {
	var diags diag.Diagnostics

	opts.ForceHandlers = m.ForceHandlers.ValueBool()

	var skipTags []string
	if !m.SkipTags.IsNull() {
		diags.Append(m.SkipTags.ElementsAs(ctx, &skipTags, false)...)
	}
	opts.SkipTags = skipTags

	opts.StartAtTask = m.StartAtTask.ValueString()

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

	var privateKeysModel []PrivateKeyModel
	if !m.PrivateKeys.IsNull() {
		diags.Append(m.PrivateKeys.ElementsAs(ctx, &privateKeysModel, false)...)
	}

	privateKeys := make([]string, 0, len(privateKeysModel))
	for _, privateKeyModel := range privateKeysModel {
		privateKeys = append(privateKeys, privateKeyModel.Name.ValueString())
	}
	opts.PrivateKeys = privateKeys

	return diags
}

func (m PrivateKeyModel) Value(ctx context.Context, key *ansible.PrivateKey) diag.Diagnostics {
	var diags diag.Diagnostics

	key.Name = m.Name.ValueString()
	key.Data = m.Data.ValueString()

	return diags
}

func (ArtifactQueryModel) AttrTypes() map[string]attr.Type {
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

func (r *NavigatorRunResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_navigator_run", req.ProviderTypeName)
}

func (r *NavigatorRunResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         fmt.Sprintf("Run an Ansible playbook within an execution environment (EE). Requires '%s' and a container engine to run the EEI.", ansible.NavigatorProgram),
		MarkdownDescription: fmt.Sprintf("Run an Ansible playbook within an execution environment (EE). Requires `%s` and a container engine to run the EEI.", ansible.NavigatorProgram),
		Attributes: map[string]schema.Attribute{
			// required
			"playbook": schema.StringAttribute{
				Description:         "Ansible playbook contents.",
				MarkdownDescription: "Ansible [playbook](https://docs.ansible.com/ansible/latest/playbook_guide/playbooks_intro.html) contents.",
				Required:            true,
				Validators: []validator.String{
					stringIsYAML(),
				},
			},
			"inventory": schema.StringAttribute{
				Description:         "Ansible inventory contents.",
				MarkdownDescription: "Ansible [inventory](https://docs.ansible.com/ansible/latest/getting_started/get_started_inventory.html) contents.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			// optional
			"working_directory": schema.StringAttribute{
				Description:         fmt.Sprintf("Directory which '%s' is run from. Recommended to be the root Ansible content directory (sometimes called the project directory), which is likely to contain 'ansible.cfg', 'roles/', etc.", ansible.NavigatorProgram),
				MarkdownDescription: fmt.Sprintf("Directory which `%s` is run from. Recommended to be the root Ansible [content directory](https://docs.ansible.com/ansible/latest/tips_tricks/sample_setup.html#sample-directory-layout) (sometimes called the project directory), which is likely to contain `ansible.cfg`, `roles/`, etc.", ansible.NavigatorProgram),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("."),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"execution_environment": schema.SingleNestedAttribute{
				Description:         "Execution environment related configuration.",
				MarkdownDescription: "[Execution environment](https://ansible.readthedocs.io/en/latest/getting_started_ee/index.html) related configuration.",
				Optional:            true,
				Computed:            true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					ExecutionEnvironmentModel{}.AttrTypes(),
					map[string]attr.Value{
						"container_engine":           types.StringValue(defaultNavigatorRunContainerEngine),
						"environment_variables_pass": types.ListNull(types.StringType),
						"environment_variables_set":  types.MapNull(types.StringType),
						"image":                      types.StringValue(defaultNavigatorRunImage),
						"pull_arguments":             types.ListNull(types.StringType),
						"pull_policy":                types.StringValue(defaultNavigatorRunPullPolicy),
						"container_options":          types.ListNull(types.StringType),
					},
				)),
				Attributes: map[string]schema.Attribute{
					"container_engine": schema.StringAttribute{
						Description:         fmt.Sprintf("Container engine responsible for running the execution environment container image. Options: %s. Defaults to '%s'.", wrapElementsJoin(ansible.ContainerEngineOptions(), "'"), defaultNavigatorRunContainerEngine),
						MarkdownDescription: fmt.Sprintf("[Container engine](https://ansible.readthedocs.io/projects/navigator/settings/#container-engine) responsible for running the execution environment container image. Options: %s. Defaults to `%s`.", wrapElementsJoin(ansible.ContainerEngineOptions(), "`"), defaultNavigatorRunContainerEngine),
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
						Validators: []validator.List{
							listvalidator.ValueStringsAre(stringIsEnvVarName()),
						},
					},
					"environment_variables_set": schema.MapAttribute{
						Description:         fmt.Sprintf("Environment variables to be set within the execution environment. By default '%s' is set to the current CRUD operation (%s).", navigatorRunOperationEnvVar, wrapElementsJoin(terraformOperations, "'")),
						MarkdownDescription: fmt.Sprintf("Environment variables to be [set](https://ansible.readthedocs.io/projects/navigator/settings/#set-environment-variable) within the execution environment. By default `%s` is set to the current CRUD operation (%s).", navigatorRunOperationEnvVar, wrapElementsJoin(terraformOperations, "`")),
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.Map{
							mapvalidator.KeysAre(stringIsEnvVarName()),
						},
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
					"container_options": schema.ListAttribute{
						Description:         "Extra parameters passed to the container engine command.",
						MarkdownDescription: "[Extra parameters](https://ansible.readthedocs.io/projects/navigator/settings/#container-options) passed to the container engine command.",
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
			},
			"ansible_navigator_binary": schema.StringAttribute{
				Description:         fmt.Sprintf("Path to the '%s' binary. By default '$PATH' is searched.", ansible.NavigatorProgram),
				MarkdownDescription: fmt.Sprintf("Path to the `%s` binary. By default `$PATH` is searched.", ansible.NavigatorProgram),
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
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
					"skip_tags": schema.ListAttribute{
						Description: "Only run plays and tasks whose tags do not match these values.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"start_at_task": schema.StringAttribute{
						Description: "Start the playbook at the task matching this name.",
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
					"private_keys": schema.ListNestedAttribute{
						Description:         "SSH private keys used for authentication in addition to the automatically mounted default named keys and SSH agent socket path.",
						MarkdownDescription: "SSH private keys used for authentication in addition to the [automatically mounted](https://ansible.readthedocs.io/projects/navigator/faq/#how-do-i-use-my-ssh-keys-with-an-execution-environment) default named keys and SSH agent socket path.",
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "Key name.",
									Required:    true,
									Validators: []validator.String{
										stringvalidator.RegexMatches(
											regexp.MustCompile(`^[a-zA-Z0-9]*$`),
											"Must only contain only alphanumeric characters",
										),
									},
								},
								"data": schema.StringAttribute{
									Description: "Key data.",
									Required:    true,
									Sensitive:   true,
									Validators: []validator.String{
										stringIsSSHPrivateKey(),
									},
								},
							},
						},
					},
				},
			},
			"timezone": schema.StringAttribute{
				Description:         fmt.Sprintf("IANA time zone, use 'local' for the system time zone. Defaults to '%s'.", defaultNavigatorRunTimezone),
				MarkdownDescription: fmt.Sprintf("IANA time zone, use `local` for the system time zone. Defaults to `%s`.", defaultNavigatorRunTimezone),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(defaultNavigatorRunTimezone),
				Validators: []validator.String{
					stringIsIANATimezone(),
				},
			},
			"run_on_destroy": schema.BoolAttribute{
				Description:         fmt.Sprintf("Run playbook on destroy. The environment variable '%s' is set to '%s' during the run to allow for conditional plays, tasks, etc. Defaults to '%t'.", navigatorRunOperationEnvVar, TerraformOperation(terraformOperationDelete).String(), defaultNavigatorRunOnDestroy),
				MarkdownDescription: fmt.Sprintf("Run playbook on destroy. The environment variable `%s` is set to `%s` during the run to allow for conditional plays, tasks, etc. Defaults to `%t`.", navigatorRunOperationEnvVar, TerraformOperation(terraformOperationDelete).String(), defaultNavigatorRunOnDestroy),
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
				MarkdownDescription: "Query the playbook artifact with [JSONPath](https://goessner.net/articles/JsonPath/). The [playbook artifact](https://access.redhat.com/documentation/en-us/red_hat_ansible_automation_platform/2.0-ea/html/ansible_navigator_creator_guide/assembly-troubleshooting-navigator_ansible-navigator#proc-review-artifact_troubleshooting-navigator) contains detailed information about every play and task, as well as the stdout from the playbook run.",
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

	runs := uint32(1)
	SetRuns(ctx, &resp.Diagnostics, resp.Private.SetKey, runs)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(uuid.New().String())
	r.Run(ctx, &resp.Diagnostics, data, runs, terraformOperationCreate)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NavigatorRunResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *NavigatorRunResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state *NavigatorRunResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if r.ShouldRun(data, state) {
		runs := IncrementRuns(ctx, &resp.Diagnostics, req.Private.GetKey, resp.Private.SetKey)

		if resp.Diagnostics.HasError() {
			return
		}

		r.Run(ctx, &resp.Diagnostics, data, runs, terraformOperationUpdate)

		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		data.Command = state.Command
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NavigatorRunResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *NavigatorRunResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	runs := IncrementRuns(ctx, &resp.Diagnostics, req.Private.GetKey, resp.Private.SetKey)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.RunOnDestroy.ValueBool() {
		r.Run(ctx, &resp.Diagnostics, data, runs, terraformOperationDelete)
	}
}
