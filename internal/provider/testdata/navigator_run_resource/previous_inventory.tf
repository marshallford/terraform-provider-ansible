resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - name: Test
    hosts: all
    gather_facts: false
    become: false
    tasks:
    - name: Get inventory of previous run
      ansible.builtin.command:
        cmd: "ansible-inventory --list -i {{ lookup('ansible.builtin.env', 'ANSIBLE_TF_PREVIOUS_INVENTORY') }}"
      register: previous_inventory
      delegate_to: localhost
      run_once: true
      when: lookup('ansible.builtin.env', 'ANSIBLE_TF_OPERATION') == 'update'
    - name: Set test facts
      ansible.builtin.set_fact:
        previous_hosts: "{{ (previous_inventory.stdout | from_json)._meta.hostvars.keys() | list }}"
      when: lookup('ansible.builtin.env', 'ANSIBLE_TF_OPERATION') == 'update'
  EOT
  inventory                = file(var.inventory_file)
  execution_environment = {
    enabled = var.ee_enabled
  }
  artifact_queries = {
    "previous_hosts" = {
      jq_filter = <<-EOT
      .plays[] | select(.name=="Test") |
      .tasks[] | select(.task=="Set test facts") |
      .res.ansible_facts.previous_hosts
      EOT
    }
  }
}

output "previous_hosts" {
  value = jsondecode(ansible_navigator_run.test.artifact_queries.previous_hosts.results[0])
}

variable "inventory_file" {
  type     = string
  nullable = false
}

variable "ee_enabled" {
  type     = bool
  nullable = false
}
