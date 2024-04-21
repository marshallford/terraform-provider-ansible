data "aws_region" "this" {}

data "aws_availability_zones" "this" {
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "this" {
  cidr_block           = "172.16.0.0/16"
  instance_tenancy     = "default"
  enable_dns_hostnames = true
  enable_dns_support   = true
  tags                 = { Name = "ansible-provider-example" }
}

resource "aws_default_route_table" "this" {
  default_route_table_id = aws_vpc.this.default_route_table_id
  tags                   = { Name = "ansible-provider-example" }
}

resource "aws_subnet" "this" {
  vpc_id               = aws_vpc.this.id
  cidr_block           = "172.16.0.0/24"
  availability_zone_id = data.aws_availability_zones.this.zone_ids[0]
  tags                 = { Name = "ansible-provider-example" }
}

resource "aws_security_group" "endpoints" {
  name   = "ansible-provider-example-endpoints"
  vpc_id = aws_vpc.this.id

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = [aws_vpc.this.cidr_block]
  }
}

resource "aws_vpc_endpoint" "this" {
  for_each = toset([
    "com.amazonaws.${data.aws_region.this.name}.ssm",
    "com.amazonaws.${data.aws_region.this.name}.ec2messages",
    "com.amazonaws.${data.aws_region.this.name}.ec2",
    "com.amazonaws.${data.aws_region.this.name}.ssmmessages",
  ])
  vpc_id              = aws_vpc.this.id
  subnet_ids          = [aws_subnet.this.id]
  security_group_ids  = [aws_security_group.endpoints.id]
  service_name        = each.key
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = true
}
