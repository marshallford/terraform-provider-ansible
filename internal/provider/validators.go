package provider

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"golang.org/x/crypto/ssh"
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

		return
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
