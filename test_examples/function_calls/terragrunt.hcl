terraform {
  # source = "git::git@github.com:transcend-io/terraform-aws-fargate-container?ref=v0.0.4"
  source = local.config.inputs.stack_name
}

locals {
  config = read_terragrunt_config("test.hcl")
}

inputs = {
  foo = local.config.inputs.stack_name
}