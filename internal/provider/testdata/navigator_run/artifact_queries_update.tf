resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - name: Test
    hosts: localhost
    become: false
    tasks:
    - name: write file
      ansible.builtin.copy:
        dest: /tmp/test
        content: acc_update
    - name: get file
      ansible.builtin.slurp:
        src: /tmp/test
  EOT
  inventory                = "# localhost"
  artifact_queries = {
    stdout = {
      jsonpath = "$.stdout"
    }
    file_contents = {
      jsonpath = "$.plays[?(@.__play_name=='Test')].tasks[?(@.__task=='get file')].res.content"
    }
  }
}

output "file_contents" {
  value = base64decode(ansible_navigator_run.test.artifact_queries.file_contents.result)
}

variable "ansible_navigator_binary" {
  type     = string
  nullable = false
}
