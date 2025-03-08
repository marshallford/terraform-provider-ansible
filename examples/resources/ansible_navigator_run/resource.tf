# 1. inline playbook and inventory
resource "ansible_navigator_run" "inline" {
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

# 2. existing playbook and inventory
resource "ansible_navigator_run" "existing" {
  playbook  = file("playbook.yaml")
  inventory = file("inventory/baremetal.yaml")
}

# 3. configure ansible with ansible.cfg placed in working directory (see example below)
resource "ansible_navigator_run" "working_directory" {
  playbook          = "# example"
  inventory         = yamlencode({})
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
  inventory = yamlencode({})
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
  playbook  = "# example"
  inventory = yamlencode({})
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
  inventory      = yamlencode({})
  run_on_destroy = true
}

# 7. triggers
resource "ansible_navigator_run" "triggers" {
  playbook  = "# example"
  inventory = yamlencode({})
  triggers = {
    run         = some_resource.example.status # run playbook when status changes
    replace     = some_resource.example.id     # recreate resource when id changes
    known_hosts = some_resource.example.name   # reset known_hosts when name changes
  }
}

# 8. artifact queries -- get playbook stdout
resource "ansible_navigator_run" "artifact_query_stdout" {
  playbook  = "# example"
  inventory = yamlencode({})
  artifact_queries = {
    "stdout" = {
      jq_filter = ".stdout"
    }
  }
}

output "playbook_stdout" {
  value = join("\n", jsondecode(ansible_navigator_run.artifact_query_stdout.artifact_queries.stdout.results[0]))
}

# 9. ssh private keys
resource "tls_private_key" "client" {
  algorithm = "ED25519"
}

resource "ansible_navigator_run" "private_keys" {
  playbook  = "# example"
  inventory = yamlencode({})
  ansible_options = {
    private_keys = [
      {
        name = "example"
        data = tls_private_key.client.private_key_openssh
      }
    ]
  }
}

# 10. ssh known hosts
resource "tls_private_key" "server" {
  algorithm = "ED25519"
}

resource "ansible_navigator_run" "known_hosts" {
  playbook = "# example"
  inventory = yamlencode({
    all = {
      vars = {
        ansible_ssh_common_args = "-o StrictHostKeyChecking=yes -o UserKnownHostsFile={{ ansible_ssh_known_hosts_file }}"
      }
    }
  })
  ansible_options = {
    known_hosts = [
      "10.0.0.1 ${tls_private_key.server.public_key_openssh}"
    ]
  }
}

# 11. compare previous inventory with current inventory
resource "ansible_navigator_run" "compare_inventory" {
  playbook = <<-EOT
  - hosts: all
    tasks:
    - name: Get inventory of previous run
      ansible.builtin.command:
        cmd: "ansible-inventory --list -i {{ lookup('ansible.builtin.env', 'ANSIBLE_TF_PREVIOUS_INVENTORY') }}"
      register: previous_inventory
      delegate_to: localhost
      run_once: true
      when: lookup('ansible.builtin.env', 'ANSIBLE_TF_OPERATION') == 'update'
    - name: Get inventory of current run
      ansible.builtin.command:
        cmd: "ansible-inventory --list -i {{ inventory_file }}"
      register: current_inventory
      delegate_to: localhost
      run_once: true
      when: lookup('ansible.builtin.env', 'ANSIBLE_TF_OPERATION') == 'update'
    - name: Compare
      ansible.builtin.debug:
        msg: "{{ previous_hosts | difference(current_hosts) }}"
      vars:
        previous_hosts: "{{ (previous_inventory.stdout | from_json)._meta.hostvars.keys() | list }}"
        current_hosts: "{{ (current_inventory.stdout | from_json)._meta.hostvars.keys() | list }}"
      when: lookup('ansible.builtin.env', 'ANSIBLE_TF_OPERATION') == 'update'
  EOT
  inventory = yamlencode({
    all = {
      hosts = {
        a = { ansible_host = "host-a.example.com" }
        b = { ansible_host = "host-b.example.com" }
      }
    }
  })
}
