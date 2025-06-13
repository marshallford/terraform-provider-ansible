resource "ansible_navigator_run" "strict" {
  playbook = "# example"
  inventory = yamlencode({
    all = {
      vars = {
        ansible_ssh_common_args = provider::ansible::ssh_args(false)
      }
    }
  })
  ansible_options = {
    known_hosts = [
      provider::ansible::ssh_known_host("ssh-ed25519 AAAA...", "host.example.com"),
    ]
  }
}

resource "ansible_navigator_run" "accept_new" {
  playbook = "# example"
  inventory = yamlencode({
    all = {
      vars = {
        ansible_ssh_common_args = provider::ansible::ssh_args(true)
      }
    }
  })
}
