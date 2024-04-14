# inline playbook and inventory
resource "ansible_navigator_run" "inline" {
  working_directory = "/some/dir"
  playbook          = <<-EOT
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

# existing playbook and inventory
locals {
  dir = "/some/dir"
}

resource "ansible_navigator_run" "existing" {
  working_directory = local.dir
  playbook          = file("${local.dir}/playbook.yaml")
  inventory         = file("${local.dir}/inventory/baremetal.yaml")
}

# set and pass environment variables
resource "ansible_navigator_run" "environment_variables" {
  working_directory = "/home/username/ansible-project"
  playbook          = <<-EOT
  - hosts: all
    tasks:
    - ansible.builtin.debug:
        msg: "{{ item }}"
      loop:
      - "{{ lookup('ansible.builtin.env', 'SOME_VAR') }}"
      - "{{ lookup('ansible.builtin.env', 'EDITOR') }}"
  EOT
  inventory         = "..."
  execution_environment = {
    environment_variables_set = {
      "SOME_VAR" = "foobar"
    }
    environment_variables_pass = [
      "EDITOR",
    ]
  }
}

# ansible playbook options
resource "ansible_navigator_run" "ansible_options" {
  working_directory = "/some/dir"
  playbook          = "..."
  inventory         = "..."
  ansible_options = {
    force_handlers = true               # --force-handlers
    skip_tags      = ["tag1", "tag2"]   # --skip-tags tag1,tag2
    start_at_task  = "task name"        # --start-at-task task name
    limit          = ["host1", "host2"] # --limit host1,host2
    tags           = ["tag3", "tag4"]   # --tags tag3,tag4
  }
}

# run on destroy
resource "ansible_navigator_run" "destroy" {
  working_directory = "/some/dir"
  playbook          = <<-EOT
  - hosts: all
    tasks:
    - ansible.builtin.set_fact:
        destroy: "{{ lookup('ansible.builtin.env', 'ANSIBLE_TF_OPERATION') == 'destroy' }}"
    - ansible.builtin.debug:
        msg: "resource is being destroyed!"
      when: destroy
  EOT
  inventory         = "..."
  run_on_destroy    = true
}

# triggers and replacement triggers
resource "ansible_navigator_run" "triggers" {
  working_directory = "/some/dir"
  playbook          = "..."
  inventory         = "..."
  triggers = {
    somekey = some_resource.example.id # re-run playbook when id changes
  }
  replacement_triggers = {
    somekey = some_resource.example.id # recreate resource when id changes
  }
}

# artifact queries -- get playbook stdout
resource "ansible_navigator_run" "artifact_query_stdout_example" {
  working_directory = "/some/dir"
  playbook          = "..."
  inventory         = "..."
  artifact_queries = {
    "stdout" = {
      jsonpath = "$.stdout"
    }
  }
}

output "playbook_stdout" {
  value = join("\n", jsondecode(ansible_navigator_run.artifact_query_stdout_example.artifact_queries.stdout.result))
}

# artifact queries -- get file contents
resource "ansible_navigator_run" "artifact_query_file_example" {
  working_directory = "/some/dir"
  playbook          = <<-EOT
  - name: Get file
    hosts: all
    become: false
    tasks:
    - name: resolv.conf
      ansible.builtin.slurp:
        src: /etc/resolv.conf
  EOT
  inventory         = "..."
  artifact_queries = {
    "resolv_conf" = {
      jsonpath = "$.plays[?(@.__play_name=='Get file')].tasks[?(@.__task=='resolv.conf')].res.content"
    }
  }
}

output "resolv_conf" {
  value = base64decode(ansible_navigator_run.artifact_query_file_example.artifact_queries.resolv_conf.result)
}
