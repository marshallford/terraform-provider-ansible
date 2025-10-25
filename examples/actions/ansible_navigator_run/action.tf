action "ansible_navigator_run" "this" {
  config {
    playbook = <<-EOT
    - hosts: webservers
      tasks:
      - ansible.builtin.package:
          name: nginx
    EOT
    inventory = yamlencode({
      webservers = {
        hosts = {
          a = { ansible_host = "webserver-a.example.com" }
        }
      }
    })
  }
}

resource "terraform_data" "this" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.ansible_navigator_run.this]
    }
  }
}
