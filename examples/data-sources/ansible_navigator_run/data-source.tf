# 1. inline playbook and inventory
data "ansible_navigator_run" "inline" {
  playbook = <<-EOT
  - hosts: some_group
    become: false
    tasks:
    - ansible.builtin.debug:
        msg: "{{ some_var }}"
  EOT
  inventory = yamlencode({
    some_group = {
      hosts = {
        local_container = {
          ansible_connection = "local"
          some_var           = "hello world!"
        }
      }
    }
  })
}

# 2. artifact queries -- get file contents
data "ansible_navigator_run" "artifact_query_file" {
  playbook  = <<-EOT
  - name: Example
    tasks:
    - name: Get file
      ansible.builtin.slurp:
        src: /etc/resolv.conf
  EOT
  inventory = yamlencode({})
  artifact_queries = {
    "resolv_conf" = {
      jq_filter = ".plays[] | select(.name==\"Example\") | .tasks[] | select(.task==\"Get file\") | .res.content"
    }
  }
}

output "resolv_conf" {
  value = base64decode(jsondecode(data.ansible_navigator_run.artifact_query_file.artifact_queries.resolv_conf.results[0]))
}
