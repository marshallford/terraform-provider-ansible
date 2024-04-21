resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = "%s"
  playbook                 = <<-EOT
  - hosts: test
    gather_facts: false
    become: false
    tasks:
    - ansible.builtin.raw: test
      register: connect
    - ansible.builtin.assert:
        that: connect.stdout == 'hello world!'
  EOT
  inventory = yamlencode({
    all = {
      hosts = {
        test = {
          ansible_host = "127.0.0.1"
          ansible_port = "%d"
        }
      }
    }
  })
  execution_environment = {
    container_options = [
      "--net=host",
    ]
  }
  ansible_options = {
    private_keys = [
      {
        name = "test"
        data = <<EOT
%s
        EOT
      }
    ]
  }
}
