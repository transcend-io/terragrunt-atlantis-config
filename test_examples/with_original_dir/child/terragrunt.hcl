include "root" {
  path   = find_in_parent_folders()
  expose = true
}

include "common_configs" {
  path   = "${dirname(find_in_parent_folders())}/common/terragrunt.hcl"
  expose = true
}

terraform {
  source = "git::git@github.com:transcend-io/terraform-aws-fargate-container?ref=v0.0.4"
}

inputs = {
  foo = "bar"
}
