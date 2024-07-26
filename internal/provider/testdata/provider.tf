variable "base_run_directory" {
  type     = string
  nullable = false
}

provider "ansible" {
  base_run_directory    = var.base_run_directory
  persist_run_directory = true
}
