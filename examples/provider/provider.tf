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
