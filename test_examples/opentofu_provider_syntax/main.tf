variable "regions" {
  type    = set(string)
  default = ["us-east-1", "us-east-2", "us-west-1", "us-west-2"]
}

provider "aws" {
  alias    = "by_region"
  region   = each.key
  for_each = var.regions
}

locals {
  all_regions = var.regions
}

resource "aws_default_vpc" "default" {
  for_each = toset(local.all_regions)

  provider = aws.by_region[each.key]
}

