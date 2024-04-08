terraform {
  required_version = ">= 1.5.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.44.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "4.0.5"
    }
    ansible = {
      source  = "marshallford/ansible"
      version = "0.7.0"
    }
  }
}
provider "aws" {
  region = "us-east-2"
}

provider "tls" {}

provider "ansible" {}

data "aws_ami" "this" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-2023.*-x86_64"]
  }
}

data "aws_iam_policy_document" "this" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "this" {
  name                = "ansible-provider-example"
  assume_role_policy  = data.aws_iam_policy_document.this.json
  managed_policy_arns = ["arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"]
}

resource "aws_iam_instance_profile" "this" {
  name = "ansible-provider-example"
  role = aws_iam_role.this.name
}

resource "aws_security_group" "instance" {
  name   = "ansible-provider-example-instance"
  vpc_id = aws_vpc.this.id

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}

resource "tls_private_key" "this" {
  algorithm = "ED25519"
}

resource "aws_key_pair" "this" {
  key_name   = "ansible-provider-example"
  public_key = tls_private_key.this.public_key_openssh
}

resource "aws_instance" "this" {
  ami                    = data.aws_ami.this.id
  instance_type          = "t3a.nano"
  subnet_id              = aws_subnet.this.id
  vpc_security_group_ids = [aws_security_group.instance.id]
  iam_instance_profile   = aws_iam_instance_profile.this.name
  key_name               = aws_key_pair.this.key_name

  tags = {
    Name = "ansible-provider-example"
  }

  lifecycle {
    ignore_changes = [ami]
  }
}

resource "aws_iam_user" "ssh_ssm" {
  name = "ansible-ssm-ssh"
}

resource "aws_iam_access_key" "ssh_ssm" {
  user = aws_iam_user.ssh_ssm.name
}

data "aws_iam_policy_document" "shh_ssm" {
  statement {
    actions = ["ssm:StartSession"]
    resources = [
      aws_instance.this.arn,
      "arn:aws:ssm:*:*:document/AWS-StartSSHSession",
    ]
    condition {
      test     = "BoolIfExists"
      variable = "ssm:SessionDocumentAccessCheck"
      values   = ["true"]
    }
  }
  statement {
    actions   = ["ssm:TerminateSession", "ssm:ResumeSession"]
    resources = ["arn:aws:ssm:*:*:session/$${aws:username}-*"]
  }
}

resource "aws_iam_user_policy" "ssh_ssm" {
  name   = "ssh-ssm"
  user   = aws_iam_user.ssh_ssm.name
  policy = data.aws_iam_policy_document.shh_ssm.json
}

locals {
  inventory = yamlencode({
    all = {
      vars = {
        ansible_ssh_extra_args = "-o ProxyCommand=\"sh -c \\\"aws ssm start-session --target %h --document-name AWS-StartSSHSession --parameters 'portNumber=%p'\\\"\""
      }
      children = {
        example_group = {
          hosts = {
            example = {
              ansible_host = aws_instance.this.id
              ansible_user = "ec2-user"
            }
          }
          vars = {
            hello_msg = "hello world"
          }
        }
      }
    }
  })
}

resource "ansible_navigator_run" "this" {
  working_directory        = abspath("${path.root}/working-directory")
  playbook                 = <<-EOT
  - hosts: all
    gather_facts: false
    tasks:
    - name: wait for hosts
      ansible.builtin.wait_for_connection:
        timeout: 600
    - name: gather facts
      ansible.builtin.setup:
    - name: hello
      ansible.builtin.debug:
        msg: "{{ hello_msg }}! Distribution: {{ ansible_facts.distribution }}, System Vendor: {{ ansible_facts.system_vendor }}"
  EOT
  inventory                = local.inventory
  ansible_navigator_binary = abspath("${path.root}/.venv/bin/ansible-navigator")
  execution_environment = {
    image = "ansible-execution-env-aws-example:v1"
    environment_variables_set = {
      AWS_ACCESS_KEY_ID     = aws_iam_access_key.ssh_ssm.id
      AWS_SECRET_ACCESS_KEY = aws_iam_access_key.ssh_ssm.secret
      AWS_DEFAULT_REGION    = data.aws_region.this.name
    }
  }
  ansible_options = {
    private_keys = [
      { name = "terraform", data = tls_private_key.this.private_key_openssh },
    ]
  }
  artifact_queries = {
    "stdout" = {
      jsonpath = "$.stdout"
    }
  }
  depends_on = [aws_iam_user_policy.ssh_ssm]
}

output "playbook_stdout" {
  value = join("\n", jsondecode(ansible_navigator_run.this.artifact_queries.stdout.result))
}
