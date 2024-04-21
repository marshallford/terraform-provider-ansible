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

# 3. use custom modules, module utils, filter plugins, and roles

# ansible.cfg file in some-directory
# [defaults]
# library=library
# module_utils=module_utils
# filter_plugins=filter_plugins
# roles_path=roles
# ...

resource "ansible_navigator_run" "working_directory" {
  playbook          = "..."
  inventory         = "..."
  working_directory = "some-directory"
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
        destroy: "{{ lookup('ansible.builtin.env', 'ANSIBLE_TF_OPERATION') == 'destroy' }}"
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
resource "ansible_navigator_run" "artifact_query_stdout_example" {
  playbook  = "..."
  inventory = "..."
  artifact_queries = {
    "stdout" = {
      jsonpath = "$.stdout"
    }
  }
}

output "playbook_stdout" {
  value = join("\n", jsondecode(ansible_navigator_run.artifact_query_stdout_example.artifact_queries.stdout.result))
}

# 9. artifact queries -- get file contents
resource "ansible_navigator_run" "artifact_query_file_example" {
  playbook  = <<-EOT
  - name: Get file
    hosts: all
    become: false
    tasks:
    - name: resolv.conf
      ansible.builtin.slurp:
        src: /etc/resolv.conf
  EOT
  inventory = "..."
  artifact_queries = {
    "resolv_conf" = {
      jsonpath = "$.plays[?(@.__play_name=='Get file')].tasks[?(@.__task=='resolv.conf')].res.content"
    }
  }
}

output "resolv_conf" {
  value = base64decode(ansible_navigator_run.artifact_query_file_example.artifact_queries.resolv_conf.result)
}
