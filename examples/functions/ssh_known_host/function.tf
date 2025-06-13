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
