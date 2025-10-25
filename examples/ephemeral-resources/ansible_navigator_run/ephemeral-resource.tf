ephemeral "ansible_navigator_run" "this" {
  playbook  = <<-EOT
  - name: Example
    hosts: all
    tasks:
    - name: Get secret
      ansible.builtin.slurp:
        src: /etc/secret
  EOT
  inventory = yamlencode({})
  artifact_queries = {
    "secret" = {
      jq_filter = <<-EOT
      .plays[] | select(.name=="Example") |
      .tasks[] | select(.task=="Get secret") |
      .res.content
      EOT
    }
  }
}

resource "vault_kv_secret_v2" "example" {
  mount = "example"
  name  = "example"
  data_json_wo = jsonencode(
    {
      some_secret = base64decode(jsondecode(
        ephemeral.ansible_navigator_run.this.artifact_queries.secret.results[0]
      ))
    }
  )
  data_json_wo_version = 1
}
