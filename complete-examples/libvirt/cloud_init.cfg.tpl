#cloud-config
hostname: ${hostname}
users:
- default
- name: root
  lock_passwd: false
  plain_text_passwd: hello
ssh_authorized_keys:
- ${ssh_authorized_key}
