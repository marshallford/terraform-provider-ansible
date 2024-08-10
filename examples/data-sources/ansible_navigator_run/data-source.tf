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
  - name: Get file
    hosts: all
    tasks:
    - name: resolv.conf
      ansible.builtin.slurp:
        src: /etc/resolv.conf
  EOT
  inventory = "..."
  artifact_queries = {
    "resolv_conf" = {
      jsonpath = "$.plays[?(@.__play_name=='Get file')].tasks[?(@.__task=='resolv.conf')].res.content"
    }
  }
}

output "resolv_conf" {
  value = base64decode(data.ansible_navigator_run.artifact_query_file.artifact_queries.resolv_conf.result)
}
