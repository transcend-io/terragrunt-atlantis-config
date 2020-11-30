include {
  path = find_in_parent_folders()
}

terraform {
  source = "git::git@github.com:transcend-io/terraform-aws-fargate-container?ref=v0.0.4"
}

locals {
  atlantis_terraform_version = "0.13.9001"
}

inputs = {
  foo = "bar"
}