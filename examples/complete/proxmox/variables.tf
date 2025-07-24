variable "proxmox_endpoint" {
  type = string
}

variable "proxmox_username" {
  type    = string
  default = "root@pam"
}

variable "proxmox_password" {
  type = string
}

variable "proxmox_node" {
  type = string
}

variable "proxmox_datastore_directory" {
  type    = string
  default = "local"
}

variable "proxmox_datastore_disk" {
  type    = string
  default = "local-lvm"
}
