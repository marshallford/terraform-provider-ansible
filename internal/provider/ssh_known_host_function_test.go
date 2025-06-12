package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

func TestAccSSHKnownHostFunction_basic(t *testing.T) {
	t.Parallel()

	address := "example.com"
	publicKey, _ := testSSHKeygen(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				variable "addresses" {
					type = list(string)
				}

				variable "public_key" {
					type = string
				}

				output "test" {
					value = provider::ansible::ssh_known_host(var.public_key, var.addresses...)
				}`,
				ConfigVariables: config.Variables{
					"addresses":  config.ListVariable(config.StringVariable(address)),
					"public_key": config.StringVariable(publicKey),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact(fmt.Sprintf("%s %s", address, publicKey))),
				},
			},
		},
	})
}

func TestAccSSHKnownHostFunction_alternative_port(t *testing.T) {
	t.Parallel()

	address := "10.0.0.1"
	port := "2222"
	publicKey, _ := testSSHKeygen(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				variable "addresses" {
					type = list(string)
				}

				variable "public_key" {
					type = string
				}

				output "test" {
					value = provider::ansible::ssh_known_host(var.public_key, var.addresses...)
				}`,
				ConfigVariables: config.Variables{
					"addresses":  config.ListVariable(config.StringVariable(fmt.Sprintf("%s:%s", address, port))),
					"public_key": config.StringVariable(publicKey),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact(fmt.Sprintf("[%s]:%s %s", address, port, publicKey))),
				},
			},
		},
	})
}

func TestAccSSHKnownHostFunction_multiple_addresses(t *testing.T) {
	t.Parallel()

	address1 := "host1"
	address2 := "host2"
	publicKey, _ := testSSHKeygen(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				variable "addresses" {
					type = list(string)
				}

				variable "public_key" {
					type = string
				}

				output "test" {
					value = provider::ansible::ssh_known_host(var.public_key, var.addresses...)
				}`,
				ConfigVariables: config.Variables{
					"addresses":  config.ListVariable(config.StringVariable(address1), config.StringVariable(address2)),
					"public_key": config.StringVariable(publicKey),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact(fmt.Sprintf("%s,%s %s", address1, address2, publicKey))),
				},
			},
		},
	})
}

func TestAccSSHKnownHostFunction_errors(t *testing.T) {
	t.Parallel()

	publicKey, _ := testSSHKeygen(t)

	tests := []struct {
		name     string
		value    string
		expected *regexp.Regexp
	}{
		{
			name:     "invalid_public_key",
			value:    `provider::ansible::ssh_known_host("invalid key", "example.com")`,
			expected: regexp.MustCompile("failed to parse public key"),
		},
		{
			name:     "no_addresses",
			value:    fmt.Sprintf(`provider::ansible::ssh_known_host("%s")`, publicKey),
			expected: regexp.MustCompile("no addresses provided"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`
						output "test" {
							value = %s
						}`, test.value),
						ExpectError: test.expected,
					},
				},
			})
		})
	}
}
