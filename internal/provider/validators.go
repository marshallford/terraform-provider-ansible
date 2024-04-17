package provider

import (
	"context"
	"errors"
	"path/filepath"
	"time"
	_ "time/tzdata"
	"unicode"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
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
			"Must be an absolute path",
		)
	}
}

func stringIsAbsolutePath() stringIsAbsolutePathValidator {
	return stringIsAbsolutePathValidator{}
}

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
			"Not an environment variable name",
			"Environment variable name must not be empty",
		)

		return
	}

	for _, r := range req.ConfigValue.ValueString() {
		if r > unicode.MaxASCII || !unicode.IsPrint(r) || r == '=' {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Not an environment variable name",
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
	return "string must be valid YAML"
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
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Not valid YAML",
			"Must be valid YAML",
		)

		return
	}

	_, err = yaml.Marshal(output)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Not valid YAML",
			"Must be valid YAML",
		)
	}
}

func stringIsYAML() stringIsYAMLValidator {
	return stringIsYAMLValidator{}
}

type stringIsIANATimezoneValidator struct{}

func (v stringIsIANATimezoneValidator) Description(ctx context.Context) string {
	return "string must be a valid IANA time zone"
}

func (v stringIsIANATimezoneValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsIANATimezoneValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	if req.ConfigValue.ValueString() == "local" {
		return
	}

	if _, err := time.LoadLocation(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Not an IANA time zone",
			"Must be an IANA time zone, use 'local' for the system time zone",
		)
	}
}

func stringIsIANATimezone() stringIsIANATimezoneValidator {
	return stringIsIANATimezoneValidator{}
}
