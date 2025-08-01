terraform {
  required_version = ">= 1.12.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.99.1"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "4.1.0"
    }
    ansible = {
      source  = "marshallford/ansible"
      version = "0.31.0"
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
  name               = "ansible-provider-example"
  assume_role_policy = data.aws_iam_policy_document.this.json
}

resource "aws_iam_role_policy_attachment" "this" {
  role       = aws_iam_role.this.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
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
  for_each               = toset(["a", "b"])
  ami                    = data.aws_ami.this.id
  instance_type          = "t3a.nano"
  subnet_id              = aws_subnet.this.id
  vpc_security_group_ids = [aws_security_group.instance.id]
  iam_instance_profile   = aws_iam_instance_profile.this.name
  key_name               = aws_key_pair.this.key_name

  tags = {
    Name = "ansible-provider-example-${each.key}"
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
    resources = concat(
      [for instance in aws_instance.this : instance.arn],
      ["arn:aws:ssm:*:*:document/AWS-StartSSHSession"],
    )
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
        ansible_ssh_common_args    = "-o ProxyCommand=\"sh -c \\\"aws ssm start-session --target %h --document-name AWS-StartSSHSession --parameters 'portNumber=%p'\\\"\""
        ansible_python_interpreter = "/usr/bin/python3"
        ansible_user               = "ec2-user"
        hello_msg                  = "hello world (AWS)"
      }
      children = {
        example_group = {
          hosts = { for instance in aws_instance.this : instance.tags.Name => {
            ansible_host = instance.id
          } }
        }
      }
    }
  })
}

resource "ansible_navigator_run" "this" {
  playbook                 = file("${path.root}/playbook.yaml")
  inventory                = local.inventory
  working_directory        = "${path.root}/working-directory"
  ansible_navigator_binary = "${path.root}/.venv/bin/ansible-navigator"
  execution_environment = {
    container_engine = "docker" # same engine used to build the EEI
    image            = "terraform-provider-ansible-example-aws:v1"
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
      jq_filter = ".stdout"
    }
  }
  depends_on = [aws_iam_user_policy.ssh_ssm]
}

output "playbook_stdout" {
  value = join("\n", jsondecode(ansible_navigator_run.this.artifact_queries.stdout.results[0]))
}
