include "tf_13" {
  path = find_in_parent_folders("use_terraform_13_parent.hcl")
}

include "tf_12" {
  path = find_in_parent_folders("use_terraform_12_parent.hcl")
}

terraform {
  source = "git::git@github.com:transcend-io/terraform-aws-fargate-container?ref=v0.0.4"
}

inputs = {
  foo = "bar"
}