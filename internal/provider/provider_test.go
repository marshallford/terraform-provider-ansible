package provider_test

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/marshallford/terraform-provider-ansible/internal/provider"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){ //nolint:gochecknoglobals
	"ansible": providerserver.NewProtocol6WithError(provider.New("test")()),
	"echo":    echoprovider.NewProviderServer(),
}

func TestAccProvider_errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expected *regexp.Regexp
	}{
		{
			name:     "unknown_base_run_directory",
			expected: regexp.MustCompile("Unknown configuration value 'base_run_directory'"),
		},
		{
			name:     "unknown_persist_run_directory",
			expected: regexp.MustCompile("Unknown configuration value 'persist_run_directory'"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      testTerraformFiles(t, filepath.Join("provider", "errors", test.name)),
						ExpectError: test.expected,
					},
				},
			})
		})
	}
}
