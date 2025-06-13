package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

func TestAccSSHArgsFunction_yes(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = provider::ansible::ssh_args(false)
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact("-o StrictHostKeyChecking=yes -o UserKnownHostsFile={{ ansible_ssh_known_hosts_file }}")),
				},
			},
		},
	})
}

func TestAccSSHArgsFunction_accept_new(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = provider::ansible::ssh_args(true)
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact("-o StrictHostKeyChecking=accept-new -o UserKnownHostsFile={{ ansible_ssh_known_hosts_file }}")),
				},
			},
		},
	})
}
