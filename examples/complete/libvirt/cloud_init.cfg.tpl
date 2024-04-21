#cloud-config
hostname: ${hostname}
users:
- default
ssh_authorized_keys:
- ${ssh_authorized_key}
