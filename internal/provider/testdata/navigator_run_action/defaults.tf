action "ansible_navigator_run" "test" {
  config {
    ansible_navigator_binary = var.ansible_navigator_binary
    playbook                 = <<-EOT
    - hosts: localhost
      gather_facts: false
      become: false
    EOT
    inventory                = "# localhost"
    execution_environment    = {}
    ansible_options          = {}
  }
}

resource "terraform_data" "test" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.ansible_navigator_run.test]
    }
  }
}
