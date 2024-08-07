resource "terraform_data" "this" {
  input = false
}

provider "ansible" {
  persist_run_directory = terraform_data.this.output
}

data "ansible_navigator_run" "test" {
  playbook  = <<-EOT
  - hosts: localhost
    become: false
  EOT
  inventory = "# localhost"
}
