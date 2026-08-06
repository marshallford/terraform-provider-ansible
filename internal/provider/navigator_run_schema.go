package provider

import (
	"maps"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/dynamicplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible/navigator"
)

func navigatorRunDescription(target surface) attrDescription {
	preamble := "Run an Ansible playbook."

	switch target {
	case surfaceResource, surfaceAction:
	case surfaceDataSource:
		preamble = "Run an Ansible playbook to gather information. It is recommended to only run playbooks without observable side effects."
	case surfaceEphemeral:
		preamble = "Run an Ansible playbook to gather temporary and likely sensitive information. It is recommended to only run playbooks without observable side effects."
	}

	return describe("%s Requires `%s` and a container engine to run within an execution environment (EE).", preamble, navigator.Program)
}

func playbookDescription() attrDescription {
	return describe("Ansible [playbook](https://docs.ansible.com/ansible/latest/playbook_guide/playbooks_intro.html) contents (YAML).")
}

func inventoryDescription(target surface) attrDescription {
	description := describe("Ansible [inventory](https://docs.ansible.com/ansible/latest/getting_started/get_started_inventory.html) contents. The environment variable `%s` is set to the path of the inventory in cases where `{{ inventory_file }}` cannot be referenced.", navigatorRunInventoryEnvVar)

	if target != surfaceResource {
		return description
	}

	return description.append("In addition, the environment variable `%s` is set to the path of the last applied inventory when the resource is updated.", navigatorRunPrevInventoryEnvVar)
}

func environmentVariablesSetDescription(target surface) attrDescription {
	description := describe("Environment variables to be [set](https://ansible.readthedocs.io/projects/navigator/settings/#set-environment-variable) within the execution environment.")

	operation := terraformOpRead

	switch target {
	case surfaceResource:
		return description.append("`%s` is automatically set to the current CRUD operation (%s).", navigatorRunOperationEnvVar, wrapElementsJoin(terraformOps{terraformOpCreate, terraformOpUpdate, terraformOpDelete}.Strings(), "`"))
	case surfaceDataSource:
	case surfaceEphemeral:
		operation = terraformOpOpen
	case surfaceAction:
		operation = terraformOpInvoke
	}

	return description.append("`%s` is automatically set to `%s`.", navigatorRunOperationEnvVar, operation)
}

func navigatorRunAttributes(target surface) map[string]schema.Attribute {
	descriptions := map[string]attrDescription{
		"playbook":                 playbookDescription(),
		"working_directory":        describe("Directory in which `%s` runs. Recommended to be the root Ansible [content directory](https://docs.ansible.com/ansible/latest/tips_tricks/sample_setup.html#sample-directory-layout) (sometimes called the project directory), which is likely to contain `ansible.cfg`, `roles/`, etc. Defaults to `%s`.", navigator.Program, defaultNavigatorRunWorkingDir),
		"execution_environment":    describe("[Execution environment](https://ansible.readthedocs.io/en/latest/getting_started_ee/index.html) (EE) related configuration."),
		"ansible_navigator_binary": describe("Path to the `%s` binary. By default `$PATH` is searched.", navigator.Program),
		"ansible_options":          describe("Ansible [playbook](https://docs.ansible.com/ansible/latest/cli/ansible-playbook.html) run related configuration."),
		"timezone":                 describe("IANA time zone, use `local` for the system time zone. Defaults to `%s`.", defaultNavigatorRunTimezone),
		"artifact_queries":         describe("Query the Ansible playbook artifact with [`jq`](https://jqlang.github.io/jq/) syntax. The [playbook artifact](https://access.redhat.com/documentation/en-us/red_hat_ansible_automation_platform/2.0-ea/html/ansible_navigator_creator_guide/assembly-troubleshooting-navigator_ansible-navigator#proc-review-artifact_troubleshooting-navigator) contains detailed information about every play and task, as well as the stdout from the playbook run."),
		"id":                       describe("UUID."),
		"command":                  describe("Generated `%s` run command. Useful for troubleshooting.", navigator.Program),
	}

	attributes := map[string]schema.Attribute{
		"playbook": schema.StringAttribute{
			Description:         descriptions["playbook"].Description,
			MarkdownDescription: descriptions["playbook"].MarkdownDescription,
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringIsYAML(),
			},
		},
		"inventory": schema.StringAttribute{
			Description:         inventoryDescription(target).Description,
			MarkdownDescription: inventoryDescription(target).MarkdownDescription,
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"working_directory": schema.StringAttribute{
			Description:         descriptions["working_directory"].Description,
			MarkdownDescription: descriptions["working_directory"].MarkdownDescription,
			Optional:            true,
			Computed:            target.allowsComputed(),
			Default:             target.stringDefault(defaultNavigatorRunWorkingDir),
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"execution_environment": schema.SingleNestedAttribute{
			Description:         descriptions["execution_environment"].Description,
			MarkdownDescription: descriptions["execution_environment"].MarkdownDescription,
			Optional:            true,
			Computed:            target.allowsComputed(),
			Default:             target.objectDefault(ExecutionEnvironmentModel{}.Defaults()),
			Attributes:          executionEnvironmentAttributes(target),
		},
		"ansible_navigator_binary": schema.StringAttribute{
			Description:         descriptions["ansible_navigator_binary"].Description,
			MarkdownDescription: descriptions["ansible_navigator_binary"].MarkdownDescription,
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"ansible_options": schema.SingleNestedAttribute{
			Description:         descriptions["ansible_options"].Description,
			MarkdownDescription: descriptions["ansible_options"].MarkdownDescription,
			Optional:            true,
			Computed:            target.allowsComputed(),
			Default:             target.objectDefault(AnsibleOptionsModel{}.Defaults()),
			Attributes:          ansibleOptionsAttributes(target),
		},
		"timezone": schema.StringAttribute{
			Description:         descriptions["timezone"].Description,
			MarkdownDescription: descriptions["timezone"].MarkdownDescription,
			Optional:            true,
			Computed:            target.allowsComputed(),
			Default:             target.stringDefault(defaultNavigatorRunTimezone),
			Validators: []validator.String{
				stringIsIANATimezone(),
			},
		},
	}

	if target == surfaceResource {
		maps.Copy(attributes, navigatorRunResourceAttributes())
	}

	// The action reports playbook output as progress events rather than state.
	if target != surfaceAction {
		maps.Copy(attributes, map[string]schema.Attribute{
			"artifact_queries": schema.MapNestedAttribute{
				Description:         descriptions["artifact_queries"].Description,
				MarkdownDescription: descriptions["artifact_queries"].MarkdownDescription,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: artifactQueryAttributes(),
				},
			},
			"id": schema.StringAttribute{
				Description:         descriptions["id"].Description,
				MarkdownDescription: descriptions["id"].MarkdownDescription,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"command": schema.StringAttribute{
				Description:         descriptions["command"].Description,
				MarkdownDescription: descriptions["command"].MarkdownDescription,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		})
	}

	return attributes
}

func executionEnvironmentAttributes(target surface) map[string]schema.Attribute {
	descriptions := map[string]attrDescription{
		"container_engine":           describe("[Container engine](https://ansible.readthedocs.io/projects/navigator/settings/#container-engine) responsible for running the execution environment container image. Options: %s. Defaults to `%s`.", wrapElementsJoin(navigator.AllContainerEngines().Strings(), "`"), defaultNavigatorRunContainerEngine),
		"enabled":                    describe("Enable or disable the use of an execution environment. Disabling requires `%s` and is only recommended when without a container engine. Defaults to `%t`.", ansible.PlaybookProgram, defaultNavigatorRunEEEnabled),
		"environment_variables_pass": describe("Existing environment variables to be [passed](https://ansible.readthedocs.io/projects/navigator/settings/#pass-environment-variable) through to and set within the execution environment."),
		"image":                      describe("Name of the execution environment container [image](https://ansible.readthedocs.io/projects/navigator/settings/#execution-environment-image). Defaults to `%s`.", defaultNavigatorRunImage),
		"pull_arguments":             describe("Additional [parameters](https://ansible.readthedocs.io/projects/navigator/settings/#pull-arguments) that should be added to the pull command when pulling an execution environment container image from a container registry."),
		"pull_policy":                describe("Container image [pull policy](https://ansible.readthedocs.io/projects/navigator/settings/#pull-policy). Defaults to `%s`.", defaultNavigatorRunPullPolicy),
		"container_options":          describe("[Extra parameters](https://ansible.readthedocs.io/projects/navigator/settings/#container-options) passed to the container engine command."),
	}

	return map[string]schema.Attribute{
		"container_engine": schema.StringAttribute{
			Description:         descriptions["container_engine"].Description,
			MarkdownDescription: descriptions["container_engine"].MarkdownDescription,
			Optional:            true,
			Computed:            target.allowsComputed(),
			Default:             target.stringDefault(defaultNavigatorRunContainerEngine),
			Validators: []validator.String{
				stringvalidator.OneOf(navigator.AllContainerEngines().Strings()...),
			},
		},
		"enabled": schema.BoolAttribute{
			Description:         descriptions["enabled"].Description,
			MarkdownDescription: descriptions["enabled"].MarkdownDescription,
			Optional:            true,
			Computed:            target.allowsComputed(),
			Default:             target.boolDefault(defaultNavigatorRunEEEnabled),
		},
		"environment_variables_pass": schema.ListAttribute{
			Description:         descriptions["environment_variables_pass"].Description,
			MarkdownDescription: descriptions["environment_variables_pass"].MarkdownDescription,
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.List{
				listvalidator.ValueStringsAre(stringIsEnvVarName()),
			},
		},
		"environment_variables_set": schema.MapAttribute{
			Description:         environmentVariablesSetDescription(target).Description,
			MarkdownDescription: environmentVariablesSetDescription(target).MarkdownDescription,
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.Map{
				mapvalidator.KeysAre(stringIsEnvVarName()),
			},
		},
		"image": schema.StringAttribute{
			Description:         descriptions["image"].Description,
			MarkdownDescription: descriptions["image"].MarkdownDescription,
			Optional:            true,
			Computed:            target.allowsComputed(),
			Default:             target.stringDefault(defaultNavigatorRunImage),
			Validators: []validator.String{
				stringIsContainerImageName(),
			},
		},
		"pull_arguments": schema.ListAttribute{
			Description:         descriptions["pull_arguments"].Description,
			MarkdownDescription: descriptions["pull_arguments"].MarkdownDescription,
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.List{
				listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
		"pull_policy": schema.StringAttribute{
			Description:         descriptions["pull_policy"].Description,
			MarkdownDescription: descriptions["pull_policy"].MarkdownDescription,
			Optional:            true,
			Computed:            target.allowsComputed(),
			Default:             target.stringDefault(defaultNavigatorRunPullPolicy),
			Validators: []validator.String{
				stringvalidator.OneOf(navigator.AllPullPolicies().Strings()...),
			},
		},
		"container_options": schema.ListAttribute{
			Description:         descriptions["container_options"].Description,
			MarkdownDescription: descriptions["container_options"].MarkdownDescription,
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.List{
				listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
	}
}

func ansibleOptionsAttributes(target surface) map[string]schema.Attribute {
	descriptions := map[string]attrDescription{
		"extra_vars":        describe("Set additional [variables](https://docs.ansible.com/projects/ansible/latest/playbook_guide/playbooks_variables.html#defining-variables-at-runtime) (YAML)."),
		"force_handlers":    describe("Run handlers even if a task fails."),
		"skip_tags":         describe("Only run plays and tasks whose tags do not match these values."),
		"start_at_task":     describe("Start the playbook at the task matching this name."),
		"limit":             describe("Further limit selected hosts to an additional pattern."),
		"tags":              describe("Only run plays and tasks tagged with these values."),
		"private_keys":      describe("SSH private keys used for authentication in addition to the [automatically mounted](https://ansible.readthedocs.io/projects/navigator/faq/#how-do-i-use-my-ssh-keys-with-an-execution-environment) default named keys and SSH agent socket path."),
		"known_hosts":       describe("SSH known host entries. Ansible variable `%s` set to path of `known_hosts` file and SSH option `UserKnownHostsFile` must be configured to that path. Defaults to all of the `known_hosts` entries recorded.", ansible.SSHKnownHostsFileVar),
		"host_key_checking": describe("SSH host key checking. Can help protect against man-in-the-middle attacks by verifying the identity of hosts. Ansible runner (library used by `%s`) defaults this option to `%t` explicitly.", navigator.Program, ansible.RunnerDefaultHostKeyChecking),
	}

	return map[string]schema.Attribute{
		"extra_vars": schema.StringAttribute{
			Description:         descriptions["extra_vars"].Description,
			MarkdownDescription: descriptions["extra_vars"].MarkdownDescription,
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringIsYAML(),
			},
		},
		"force_handlers": schema.BoolAttribute{
			Description:         descriptions["force_handlers"].Description,
			MarkdownDescription: descriptions["force_handlers"].MarkdownDescription,
			Optional:            true,
		},
		"skip_tags": schema.ListAttribute{
			Description:         descriptions["skip_tags"].Description,
			MarkdownDescription: descriptions["skip_tags"].MarkdownDescription,
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.List{
				listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
		"start_at_task": schema.StringAttribute{
			Description:         descriptions["start_at_task"].Description,
			MarkdownDescription: descriptions["start_at_task"].MarkdownDescription,
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"limit": schema.ListAttribute{
			Description:         descriptions["limit"].Description,
			MarkdownDescription: descriptions["limit"].MarkdownDescription,
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.List{
				listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
		"tags": schema.ListAttribute{
			Description:         descriptions["tags"].Description,
			MarkdownDescription: descriptions["tags"].MarkdownDescription,
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.List{
				listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
		"private_keys": schema.ListNestedAttribute{
			Description:         descriptions["private_keys"].Description,
			MarkdownDescription: descriptions["private_keys"].MarkdownDescription,
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: privateKeyAttributes(target),
			},
		},
		"known_hosts": schema.ListAttribute{
			Description:         descriptions["known_hosts"].Description,
			MarkdownDescription: descriptions["known_hosts"].MarkdownDescription,
			Optional:            true,
			Computed:            target.allowsComputed(),
			ElementType:         types.StringType,
			Validators: []validator.List{
				listvalidator.ValueStringsAre(stringIsSSHKnownHost()),
			},
		},
		"host_key_checking": schema.BoolAttribute{
			Description:         descriptions["host_key_checking"].Description,
			MarkdownDescription: descriptions["host_key_checking"].MarkdownDescription,
			Optional:            true,
		},
	}
}

func privateKeyAttributes(target surface) map[string]schema.Attribute {
	descriptions := map[string]attrDescription{
		"name": describe("Key name."),
		"data": describe("Key data."),
	}

	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description:         descriptions["name"].Description,
			MarkdownDescription: descriptions["name"].MarkdownDescription,
			Required:            true,
			Validators: []validator.String{
				stringIsSSHPrivateKeyName(),
			},
		},
		"data": schema.StringAttribute{
			Description:         descriptions["data"].Description,
			MarkdownDescription: descriptions["data"].MarkdownDescription,
			Required:            true,
			Sensitive:           target.allowsSensitive(),
			Validators: []validator.String{
				stringIsSSHPrivateKey(),
			},
		},
	}
}

func artifactQueryAttributes() map[string]schema.Attribute {
	descriptions := map[string]attrDescription{
		"jq_filter": describe("`jq` filter. Example: `.status, .stdout`."),
		"results":   describe("Results of the `jq` filter in JSON format."),
	}

	return map[string]schema.Attribute{
		"jq_filter": schema.StringAttribute{
			Description:         descriptions["jq_filter"].Description,
			MarkdownDescription: descriptions["jq_filter"].MarkdownDescription,
			Required:            true,
			Validators: []validator.String{
				stringIsJQFilter(),
			},
		},
		"results": schema.ListAttribute{ // TODO switch to a dynamic attribute when supported as an element in a collection
			Description:         descriptions["results"].Description,
			MarkdownDescription: descriptions["results"].MarkdownDescription,
			Computed:            true,
			ElementType:         jsontypes.NormalizedType{},
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

func navigatorRunResourceAttributes() map[string]schema.Attribute {
	descriptions := map[string]attrDescription{
		"run_on_destroy":   describe("Run playbook (or alternatively `destroy_playbook` if configured) on destroy. The environment variable `%s` is set to `%s` during the run to allow for conditional plays, tasks, etc. Defaults to `%t`.", navigatorRunOperationEnvVar, terraformOpDelete, defaultNavigatorRunOnDestroy),
		"destroy_playbook": playbookDescription().append("Only run on destroy (`run_on_destroy` must be `true`)."),
		"triggers":         describe("Trigger various behaviors via arbitrary values."),
	}

	triggers := map[string]attrDescription{
		"run":           describe("A value that, when changed, will run the playbook again. Provides a way to initiate a run without changing other attributes such as the inventory or playbook."),
		"exclusive_run": describe("When non-null, only changes to this value will run the playbook again. All other changes are ignored, the exception being resource destruction or replacement. Provides fine-grained control for advanced use cases."),
		"replace":       describe("A value that, when changed, will recreate the resource. Serves as an alternative to the native [`replace_triggered_by`](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#replace_triggered_by) lifecycle argument. Will cause `id` to change. May be useful when combined with `run_on_destroy`."),
		"known_hosts":   describe("A value that, when changed, will reset the computed list of SSH known host entries. Useful when inventory hosts are recreated with the same hostnames/IP addresses, but different SSH keypairs."),
	}

	return map[string]schema.Attribute{
		"run_on_destroy": schema.BoolAttribute{
			Description:         descriptions["run_on_destroy"].Description,
			MarkdownDescription: descriptions["run_on_destroy"].MarkdownDescription,
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(defaultNavigatorRunOnDestroy),
		},
		"destroy_playbook": schema.StringAttribute{
			Description:         descriptions["destroy_playbook"].Description,
			MarkdownDescription: descriptions["destroy_playbook"].MarkdownDescription,
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringIsYAML(),
			},
		},
		"triggers": schema.SingleNestedAttribute{
			Description:         descriptions["triggers"].Description,
			MarkdownDescription: descriptions["triggers"].MarkdownDescription,
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"run": schema.DynamicAttribute{
					Description:         triggers["run"].Description,
					MarkdownDescription: triggers["run"].MarkdownDescription,
					Optional:            true,
				},
				"exclusive_run": schema.DynamicAttribute{
					Description:         triggers["exclusive_run"].Description,
					MarkdownDescription: triggers["exclusive_run"].MarkdownDescription,
					Optional:            true,
				},
				"replace": schema.DynamicAttribute{
					Description:         triggers["replace"].Description,
					MarkdownDescription: triggers["replace"].MarkdownDescription,
					Optional:            true,
					PlanModifiers: []planmodifier.Dynamic{
						dynamicplanmodifier.RequiresReplace(),
					},
				},
				"known_hosts": schema.DynamicAttribute{
					Description:         triggers["known_hosts"].Description,
					MarkdownDescription: triggers["known_hosts"].MarkdownDescription,
					Optional:            true,
				},
			},
		},
	}
}
