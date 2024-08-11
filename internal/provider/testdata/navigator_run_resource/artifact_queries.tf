resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - name: Test
    hosts: localhost
    become: false
    tasks:
    - name: Write file
      ansible.builtin.copy:
        dest: /tmp/test
        content: ${var.file_contents}
    - name: Get file
      ansible.builtin.slurp:
        src: /tmp/test
  EOT
  inventory                = "# localhost"
  artifact_queries = {
    "stdout" = {
      jq_filter = ".stdout"
    }
    "file_contents" = {
      jq_filter = ".plays[] | select(.name==\"Test\") | .tasks[] | select(.task==\"Get file\") | .res.content"
    }
  }
}

output "file_contents" {
  value = base64decode(jsondecode(ansible_navigator_run.test.artifact_queries.file_contents.results[0]))
}

variable "ansible_navigator_binary" {
  type     = string
  nullable = false
}

variable "file_contents" {
  type     = string
  nullable = false
}
