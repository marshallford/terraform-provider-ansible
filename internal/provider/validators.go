package provider

import (
	"context"
	"errors"
	"time"
	_ "time/tzdata"
	"unicode"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/marshallford/terraform-provider-ansible/pkg/ansible"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
)

type stringIsSSHPrivateKeyValidator struct{}

func (v stringIsSSHPrivateKeyValidator) Description(ctx context.Context) string {
	return "string must be an unencrypted SSH private key"
}

func (v stringIsSSHPrivateKeyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsSSHPrivateKeyValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	_, err := ssh.ParseRawPrivateKey([]byte(req.ConfigValue.ValueString()))

	var passphraseErr *ssh.PassphraseMissingError
	if errors.As(err, &passphraseErr) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Not an unencrypted SSH private key",
			"Must be an unencrypted (meaning no passphrase) private key",
		)

		return
	}

	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Not a SSH private key",
			"Must be a RSA, DSA, ECDSA, or Ed25519 private key formatted as PKCS#1, PKCS#8, OpenSSL, or OpenSSH",
		)
	}
}

func stringIsSSHPrivateKey() stringIsSSHPrivateKeyValidator {
	return stringIsSSHPrivateKeyValidator{}
}

type stringIsSSHKnownHostValidator struct{}

func (v stringIsSSHKnownHostValidator) Description(ctx context.Context) string {
	return "string must be a SSH known host entry"
}

func (v stringIsSSHKnownHostValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsSSHKnownHostValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	if req.ConfigValue.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Not a SSH known host entry",
			"Known host entry must not be empty",
		)

		return
	}

	_, _, _, _, rest, err := ssh.ParseKnownHosts([]byte(req.ConfigValue.ValueString())) //nolint:dogsled

	addPathError(&resp.Diagnostics, req.Path, "Not a SSH known host entry", err)

	if len(rest) > 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Not a single SSH known host entry",
			"Must not include multiple known host entries or additional data",
		)
	}
}

func stringIsSSHKnownHost() stringIsSSHKnownHostValidator {
	return stringIsSSHKnownHostValidator{}
}

type stringIsEnvVarNameValidator struct{}

func (v stringIsEnvVarNameValidator) Description(ctx context.Context) string {
	return "string must be an environment variable name"
}

func (v stringIsEnvVarNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsEnvVarNameValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	if req.ConfigValue.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Not a valid environment variable name",
			"Environment variable name must not be empty",
		)

		return
	}

	for _, r := range req.ConfigValue.ValueString() {
		if r > unicode.MaxASCII || !unicode.IsPrint(r) || r == '=' {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Not a valid environment variable name",
				"Environment variable name must consist only of printable ASCII characters other than '='",
			)

			return
		}
	}
}

func stringIsEnvVarName() stringIsEnvVarNameValidator {
	return stringIsEnvVarNameValidator{}
}

type stringIsYAMLValidator struct{}

func (v stringIsYAMLValidator) Description(ctx context.Context) string {
	return "string must be YAML"
}

func (v stringIsYAMLValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsYAMLValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	var output interface{}
	err := yaml.Unmarshal([]byte(req.ConfigValue.ValueString()), &output)
	if addPathError(&resp.Diagnostics, req.Path, "Not valid YAML", err) {
		return
	}

	_, err = yaml.Marshal(output)
	addPathError(&resp.Diagnostics, req.Path, "Not valid YAML", err)
}

func stringIsYAML() stringIsYAMLValidator {
	return stringIsYAMLValidator{}
}

type stringIsIANATimezoneValidator struct{}

func (v stringIsIANATimezoneValidator) Description(ctx context.Context) string {
	return "string must be an IANA time zone"
}

func (v stringIsIANATimezoneValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsIANATimezoneValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	if req.ConfigValue.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Not a valid IANA time zone",
			"IANA time zone must not be empty",
		)

		return
	}

	if req.ConfigValue.ValueString() == "local" {
		return
	}

	_, err := time.LoadLocation(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not a valid IANA time zone, use 'local' for the system time zone", err)
}

func stringIsIANATimezone() stringIsIANATimezoneValidator {
	return stringIsIANATimezoneValidator{}
}

type stringIsIsJSONPathExpressionValidator struct{}

func (v stringIsIsJSONPathExpressionValidator) Description(ctx context.Context) string {
	return "string must be a JSONPath expression"
}

func (v stringIsIsJSONPathExpressionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsIsJSONPathExpressionValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := ansible.ValidateJSONPathExpression(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not a valid JSONPath expression", err)
}

func stringIsIsJSONPathExpression() stringIsIsJSONPathExpressionValidator {
	return stringIsIsJSONPathExpressionValidator{}
}
