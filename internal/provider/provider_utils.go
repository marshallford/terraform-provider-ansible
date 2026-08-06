package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	actionTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/action/timeouts"
	dataSourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	ephemeralResourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/ephemeral/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/action"
	aschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	eschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	diagDetailPrefix = "Underlying error details"
)

type attrDescription struct {
	Description         string
	MarkdownDescription string
}

var markdownLink = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)

func describe(markdown string, args ...any) attrDescription {
	if len(args) > 0 {
		markdown = fmt.Sprintf(markdown, args...)
	}

	return attrDescription{
		Description:         strings.ReplaceAll(markdownLink.ReplaceAllString(markdown, "$1"), "`", "'"),
		MarkdownDescription: markdown,
	}
}

func (d attrDescription) append(markdown string, args ...any) attrDescription {
	appended := describe(markdown, args...)

	return attrDescription{
		Description:         d.Description + " " + appended.Description,
		MarkdownDescription: d.MarkdownDescription + " " + appended.MarkdownDescription,
	}
}

type providerOptions struct {
	BaseRunDirectory    string
	PersistRunDirectory bool
}

type (
	terraformOp  string
	terraformOps []terraformOp
)

const (
	terraformOpCreate terraformOp = "create"
	terraformOpRead   terraformOp = "read"
	terraformOpUpdate terraformOp = "update"
	terraformOpDelete terraformOp = "delete"
	terraformOpOpen   terraformOp = "open"
	terraformOpInvoke terraformOp = "invoke"
)

func (op terraformOp) String() string {
	return string(op)
}

func (ops terraformOps) Strings() []string {
	output := make([]string, 0, len(ops))
	for _, element := range ops {
		output = append(output, element.String())
	}

	return output
}

type surface int

const (
	surfaceResource surface = iota
	surfaceDataSource
	surfaceEphemeral
	surfaceAction
)

func (s surface) allowsComputed() bool {
	return s != surfaceAction
}

func (s surface) allowsSensitive() bool {
	return s != surfaceAction
}

func (s surface) stringDefault(value string) defaults.String { //nolint:ireturn
	if s == surfaceAction {
		return nil
	}

	return stringdefault.StaticString(value)
}

func (s surface) boolDefault(value bool) defaults.Bool { //nolint:ireturn
	if s == surfaceAction {
		return nil
	}

	return booldefault.StaticBool(value)
}

func (s surface) objectDefault(value types.Object) defaults.Object { //nolint:ireturn
	if s == surfaceAction {
		return nil
	}

	return objectdefault.StaticValue(value)
}

func dataSourceAttributes(attributes map[string]schema.Attribute) map[string]dschema.Attribute {
	converted := make(map[string]dschema.Attribute, len(attributes))
	for name, attribute := range attributes {
		converted[name] = attribute
	}

	return converted
}

func ephemeralAttributes(attributes map[string]schema.Attribute) map[string]eschema.Attribute {
	converted := make(map[string]eschema.Attribute, len(attributes))
	for name, attribute := range attributes {
		converted[name] = attribute
	}

	return converted
}

func actionAttributes(attributes map[string]schema.Attribute) map[string]aschema.Attribute {
	converted := make(map[string]aschema.Attribute, len(attributes))
	for name, attribute := range attributes {
		converted[name] = attribute
	}

	return converted
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
	case terraformOpOpen, terraformOpInvoke:
		return defaultTimeout, nil
	}

	return defaultTimeout, nil
}

func terraformOperationDataSourceTimeout(ctx context.Context, value dataSourceTimeouts.Value, defaultTimeout time.Duration) (time.Duration, diag.Diagnostics) {
	return value.Read(ctx, defaultTimeout)
}

func terraformOperationEphemeralResourceTimeout(ctx context.Context, value ephemeralResourceTimeouts.Value, defaultTimeout time.Duration) (time.Duration, diag.Diagnostics) {
	return value.Open(ctx, defaultTimeout)
}

func terraformOperationActionTimeout(ctx context.Context, value actionTimeouts.Value, defaultTimeout time.Duration) (time.Duration, diag.Diagnostics) {
	return value.Invoke(ctx, defaultTimeout)
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

func configureActionClient(req action.ConfigureRequest, resp *action.ConfigureResponse) (*providerOptions, bool) {
	if req.ProviderData == nil {
		return nil, false
	}

	opts, ok := req.ProviderData.(*providerOptions)

	if !ok {
		summary, detail := unexpectedConfigureType("Action", req.ProviderData)
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

func addWarning(diags *diag.Diagnostics, summary string, err error) bool {
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
