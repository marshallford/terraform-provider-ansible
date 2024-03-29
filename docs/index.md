---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "ansible Provider"
subcategory: ""
description: |-
  Interact with Ansible https://github.com/ansible/ansible.
---

# ansible Provider

Interact with [Ansible](https://github.com/ansible/ansible).

## Example Usage

```terraform
# defaults
provider "ansible" {}

# configure base run directory
provider "ansible" {
  base_run_directory = "/some-directory"
}

# persist run directory for troubleshooting
provider "ansible" {
  persist_run_directory = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `base_run_directory` (String) Base directory in which to create run directories. On Unix systems this defaults to `$TMPDIR` if non-empty, else `/tmp`.
- `persist_run_directory` (Boolean) Remove run directory after the run completes. Useful when troubleshooting. Defaults to `false`.
