package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
)

type stringIsSSHPrivateKeyValidator struct{}

var _ validator.String = (*stringIsSSHPrivateKeyValidator)(nil)

func (v stringIsSSHPrivateKeyValidator) Description(_ context.Context) string {
	return "string must be an unencrypted SSH private key"
}

func (v stringIsSSHPrivateKeyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsSSHPrivateKeyValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := ansible.ValidateSSHPrivateKey(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not an unencrypted SSH private key", err)
}

func stringIsSSHPrivateKey() stringIsSSHPrivateKeyValidator {
	return stringIsSSHPrivateKeyValidator{}
}

func StringIsSSHPrivateKey() validator.String { //nolint:ireturn
	return stringIsSSHPrivateKey()
}

type stringIsSSHPrivateKeyNameValidator struct{}

var _ validator.String = (*stringIsSSHPrivateKeyNameValidator)(nil)

func (v stringIsSSHPrivateKeyNameValidator) Description(_ context.Context) string {
	return "string must be a valid SSH private key name"
}

func (v stringIsSSHPrivateKeyNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsSSHPrivateKeyNameValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := ansible.ValidateSSHPrivateKeyName(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not a valid SSH private key name", err)
}

func stringIsSSHPrivateKeyName() stringIsSSHPrivateKeyNameValidator {
	return stringIsSSHPrivateKeyNameValidator{}
}

func StringIsSSHPrivateKeyName() validator.String { //nolint:ireturn
	return stringIsSSHPrivateKeyName()
}

type stringIsSSHKnownHostValidator struct{}

var _ validator.String = (*stringIsSSHKnownHostValidator)(nil)

func (v stringIsSSHKnownHostValidator) Description(_ context.Context) string {
	return "string must be a SSH known host entry"
}

func (v stringIsSSHKnownHostValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsSSHKnownHostValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := ansible.ValidateSSHKnownHost(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not a single SSH known host entry", err)
}

func stringIsSSHKnownHost() stringIsSSHKnownHostValidator {
	return stringIsSSHKnownHostValidator{}
}

func StringIsSSHKnownHost() validator.String { //nolint:ireturn
	return stringIsSSHKnownHost()
}

type stringIsEnvVarNameValidator struct{}

var _ validator.String = (*stringIsEnvVarNameValidator)(nil)

func (v stringIsEnvVarNameValidator) Description(_ context.Context) string {
	return "string must be an environment variable name"
}

func (v stringIsEnvVarNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsEnvVarNameValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := ansible.ValidateEnvVarName(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not a valid environment variable name", err)
}

func stringIsEnvVarName() stringIsEnvVarNameValidator {
	return stringIsEnvVarNameValidator{}
}

func StringIsEnvVarName() validator.String { //nolint:ireturn
	return stringIsEnvVarName()
}

type stringIsYAMLValidator struct{}

var _ validator.String = (*stringIsYAMLValidator)(nil)

func (v stringIsYAMLValidator) Description(_ context.Context) string {
	return "string must be YAML"
}

func (v stringIsYAMLValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsYAMLValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := ansible.ValidateYAML(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not valid YAML", err)
}

func stringIsYAML() stringIsYAMLValidator {
	return stringIsYAMLValidator{}
}

func StringIsYAML() validator.String { //nolint:ireturn
	return stringIsYAML()
}

type stringIsIANATimezoneValidator struct{}

var _ validator.String = (*stringIsIANATimezoneValidator)(nil)

func (v stringIsIANATimezoneValidator) Description(_ context.Context) string {
	return "string must be an IANA time zone"
}

func (v stringIsIANATimezoneValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsIANATimezoneValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := ansible.ValidateIANATimezone(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not a valid IANA time zone, use 'local' for the system time zone", err)
}

func stringIsIANATimezone() stringIsIANATimezoneValidator {
	return stringIsIANATimezoneValidator{}
}

func StringIsIANATimezone() validator.String { //nolint:ireturn
	return stringIsIANATimezone()
}

type stringIsJQFilterValidator struct{}

var _ validator.String = (*stringIsJQFilterValidator)(nil)

func (v stringIsJQFilterValidator) Description(_ context.Context) string {
	return "string must be a JQ filter"
}

func (v stringIsJQFilterValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsJQFilterValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := ansible.ValidateJQFilter(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not a valid JQ filter", err)
}

func stringIsJQFilter() stringIsJQFilterValidator {
	return stringIsJQFilterValidator{}
}

func StringIsJQFilter() validator.String { //nolint:ireturn
	return stringIsJQFilter()
}

type stringIsContainerImageNameValidator struct{}

var _ validator.String = (*stringIsContainerImageNameValidator)(nil)

func (v stringIsContainerImageNameValidator) Description(_ context.Context) string {
	return "string must be a container image name"
}

func (v stringIsContainerImageNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsContainerImageNameValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := ansible.ValidateContainerImageName(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not a valid container image name", err)
}

func stringIsContainerImageName() stringIsContainerImageNameValidator {
	return stringIsContainerImageNameValidator{}
}

func StringIsContainerImageName() validator.String { //nolint:ireturn
	return stringIsContainerImageName()
}
