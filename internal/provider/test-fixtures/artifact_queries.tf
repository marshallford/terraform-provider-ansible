resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = "%s"
  working_directory        = "%s"
  playbook                 = <<-EOT
  - name: Test
    hosts: localhost
    become: false
    tasks:
    - name: write file
      copy:
        dest: /tmp/test
        content: acc
    - name: get file
      ansible.builtin.slurp:
        src: /tmp/test
  EOT
  inventory                = "# localhost"
  artifact_queries = {
    stdout = {
      jsonpath = "$.stdout"
    }
    file = {
      jsonpath = "$.plays[?(@.__play_name=='Test')].tasks[?(@.__task=='get file')].res.content"
    }
  }
}
