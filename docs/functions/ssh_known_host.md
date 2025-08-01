---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "ssh_known_host function - terraform-provider-ansible"
subcategory: ""
description: |-
  Format a public key and addresses into a known hosts entry.
---

# function: ssh_known_host

Format a public key and addresses into a known hosts entry/line suitable for use in an SSH known hosts file.

## Example Usage

```terraform
resource "tls_private_key" "this" {
  algorithm = "ED25519"
}

output "simple" {
  value = provider::ansible::ssh_known_host(tls_private_key.this.public_key_openssh, "host-a.example.com")
}

output "alternative_port" {
  value = provider::ansible::ssh_known_host(tls_private_key.this.public_key_openssh, "host-b.example.com:2222")
}

output "multiple_addresses" {
  value = provider::ansible::ssh_known_host(tls_private_key.this.public_key_openssh, "host-a.example.com", "10.0.0.1")
}
```

## Signature

<!-- signature generated by tfplugindocs -->
```text
ssh_known_host(public_key string, addresses string...) string
```

## Arguments

<!-- arguments generated by tfplugindocs -->
1. `public_key` (String) Public key data in the authorized keys format.
<!-- variadic argument generated by tfplugindocs -->
1. `addresses` (Variadic, String) Addresses to associate with the public key. Can be one or more hostnames or IP addresses with an optional port.
