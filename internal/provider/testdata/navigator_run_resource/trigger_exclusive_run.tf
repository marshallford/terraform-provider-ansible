resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    gather_facts: false
    become: false
  EOT
  inventory                = var.inventory
  artifact_queries = {
    "test" = {
      jq_filter = "now"
    }
  }
  triggers = {
    exclusive_run = var.trigger
  }
}

variable "inventory" {
  type     = string
  nullable = false

}

variable "trigger" {
  type     = string
  nullable = true
  default  = null
}
