# 1. inline playbook and inventory
resource "ansible_navigator_run" "inline" {
  playbook = <<-EOT
  - hosts: some_group
    become: false
    tasks:
    - ansible.builtin.debug:
        msg: "{{ some_var }}"
  EOT
  inventory = yamlencode({
    all = {
      children = {
        some_group = {
          hosts = {
            local_container = {
              ansible_connection = "local"
              some_var           = "hello world!"
            }
          }
        }
      }
    }
  })
}

# 2. existing playbook and inventory
resource "ansible_navigator_run" "existing" {
  playbook  = file("playbook.yaml")
  inventory = file("inventory/baremetal.yaml")
}

# 3. configure ansible behavior with ansible.cfg placed in working directory (see example below)
resource "ansible_navigator_run" "working_directory" {
  playbook          = "..."
  inventory         = "..."
  working_directory = "some-directory-with-ansible-cfg-file"
}

# 4. set and pass environment variables
resource "ansible_navigator_run" "environment_variables" {
  playbook  = <<-EOT
  - hosts: all
    tasks:
    - ansible.builtin.debug:
        msg: "{{ item }}"
      loop:
      - "{{ lookup('ansible.builtin.env', 'SOME_VAR') }}"
      - "{{ lookup('ansible.builtin.env', 'EDITOR') }}"
  EOT
  inventory = "..."
  execution_environment = {
    environment_variables_set = {
      "SOME_VAR" = "some-value"
    }
    environment_variables_pass = [
      "EDITOR",
    ]
  }
}

# 5. ansible playbook options
resource "ansible_navigator_run" "ansible_options" {
  playbook  = "..."
  inventory = "..."
  ansible_options = {
    force_handlers = true               # --force-handlers
    skip_tags      = ["tag1", "tag2"]   # --skip-tags tag1,tag2
    start_at_task  = "task name"        # --start-at-task task name
    limit          = ["host1", "host2"] # --limit host1,host2
    tags           = ["tag3", "tag4"]   # --tags tag3,tag4
  }
}

# 6. run on destroy
resource "ansible_navigator_run" "destroy" {
  playbook       = <<-EOT
  - hosts: all
    tasks:
    - ansible.builtin.set_fact:
        destroy: "{{ lookup('ansible.builtin.env', 'ANSIBLE_TF_OPERATION') == 'delete' }}"
    - ansible.builtin.debug:
        msg: "resource is being destroyed!"
      when: destroy
  EOT
  inventory      = "..."
  run_on_destroy = true
}

# 7. triggers and replacement triggers
resource "ansible_navigator_run" "triggers" {
  playbook  = "..."
  inventory = "..."
  triggers = {
    somekey = some_resource.example.id # re-run playbook when id changes
  }
  replacement_triggers = {
    somekey = some_resource.example.id # recreate resource when id changes
  }
}

# 8. artifact queries -- get playbook stdout
resource "ansible_navigator_run" "artifact_query_stdout" {
  playbook  = "..."
  inventory = "..."
  artifact_queries = {
    "stdout" = {
      jsonpath = "$.stdout"
    }
  }
}

output "playbook_stdout" {
  value = join("\n", jsondecode(ansible_navigator_run.artifact_query_stdout.artifact_queries.stdout.result))
}

# 9. specify ssh private keys
resource "tls_private_key" "this" {
  algorithm = "ED25519"
}

resource "ansible_navigator_run" "private_keys" {
  playbook  = "..."
  inventory = "..."
  ansible_options = {
    private_keys = [
      {
        name = "example"
        data = tls_private_key.this.private_key_openssh
      }
    ]
  }
}
