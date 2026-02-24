package provider_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	schemavalidator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-ansible/internal/provider"
)

func TestStringValidators_DescriptionsNotEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name      string
		validator schemavalidator.String
	}{
		{
			name:      "ssh_private_key",
			validator: provider.StringIsSSHPrivateKey(),
		},
		{
			name:      "ssh_private_key_name",
			validator: provider.StringIsSSHPrivateKeyName(),
		},
		{
			name:      "ssh_known_host",
			validator: provider.StringIsSSHKnownHost(),
		},
		{
			name:      "env_var_name",
			validator: provider.StringIsEnvVarName(),
		},
		{
			name:      "yaml",
			validator: provider.StringIsYAML(),
		},
		{
			name:      "iana_timezone",
			validator: provider.StringIsIANATimezone(),
		},
		{
			name:      "jq_filter",
			validator: provider.StringIsJQFilter(),
		},
		{
			name:      "container_image_name",
			validator: provider.StringIsContainerImageName(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			description := test.validator.Description(ctx)
			markdownDescription := test.validator.MarkdownDescription(ctx)

			if strings.TrimSpace(description) == "" {
				t.Fatal("description must not be empty")
			}

			if strings.TrimSpace(markdownDescription) == "" {
				t.Fatal("markdown description must not be empty")
			}
		})
	}
}

func TestStringValidators(t *testing.T) {
	t.Parallel()

	publicKey, privateKey := testSSHKeygen(t)

	tests := []struct {
		name          string
		validator     schemavalidator.String
		validValues   []string
		invalidValues []string
	}{
		{
			name:          "ssh_private_key",
			validator:     provider.StringIsSSHPrivateKey(),
			validValues:   []string{privateKey},
			invalidValues: []string{"not-a-private-key", ""},
		},
		{
			name:          "ssh_private_key_name",
			validator:     provider.StringIsSSHPrivateKeyName(),
			validValues:   []string{"id-ed25519", "my-key-1"},
			invalidValues: []string{"-bad-key-name", "bad-key-name-", "bad_key_name", "key!name"},
		},
		{
			name:          "ssh_known_host",
			validator:     provider.StringIsSSHKnownHost(),
			validValues:   []string{fmt.Sprintf("some-host %s", publicKey)},
			invalidValues: []string{"invalid-known-host-entry", ""},
		},
		{
			name:          "env_var_name",
			validator:     provider.StringIsEnvVarName(),
			validValues:   []string{"VALID_ENV_VAR", "TF_VAR_foo", "_VAR", "a"},
			invalidValues: []string{"NOT=VALID", ""},
		},
		{
			name:          "yaml",
			validator:     provider.StringIsYAML(),
			validValues:   []string{"key: value", "- one\n- two"},
			invalidValues: []string{"key: [", "foo: {{"},
		},
		{
			name:          "iana_timezone",
			validator:     provider.StringIsIANATimezone(),
			validValues:   []string{"UTC", "local", "America/New_York"},
			invalidValues: []string{"Not/A_Real_Timezone", ""},
		},
		{
			name:          "jq_filter",
			validator:     provider.StringIsJQFilter(),
			validValues:   []string{".foo", ".items[] | .name", "."},
			invalidValues: []string{".foo[", "if", ""},
		},
		{
			name:          "container_image_name",
			validator:     provider.StringIsContainerImageName(),
			validValues:   []string{"ghcr.io/ansible/community-ansible-dev-tools:v26.1.0", "docker.io/library/alpine:3.21"},
			invalidValues: []string{"not a valid image", ""},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			testCases := []struct {
				name     string
				value    types.String
				hasError bool
			}{
				{name: "unknown", value: types.StringUnknown(), hasError: false},
				{name: "null", value: types.StringNull(), hasError: false},
			}

			for idx, value := range test.validValues {
				testCases = append(testCases, struct {
					name     string
					value    types.String
					hasError bool
				}{
					name:     fmt.Sprintf("valid_sample_%d", idx),
					value:    types.StringValue(value),
					hasError: false,
				})
			}

			for idx, value := range test.invalidValues {
				testCases = append(testCases, struct {
					name     string
					value    types.String
					hasError bool
				}{
					name:     fmt.Sprintf("invalid_sample_%d", idx),
					value:    types.StringValue(value),
					hasError: true,
				})
			}

			for _, testCase := range testCases {
				t.Run(testCase.name, func(t *testing.T) {
					t.Parallel()

					response := schemavalidator.StringResponse{}
					test.validator.ValidateString(
						context.Background(),
						schemavalidator.StringRequest{
							Path:        path.Root("test_attribute"),
							ConfigValue: testCase.value,
						},
						&response,
					)

					if response.Diagnostics.HasError() != testCase.hasError {
						t.Fatalf("unexpected diagnostics error state: got %t, want %t", response.Diagnostics.HasError(), testCase.hasError)
					}
				})
			}
		})
	}
}
