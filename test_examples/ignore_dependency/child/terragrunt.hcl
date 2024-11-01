terraform {
  source = "git::git@github.com:transcend-io/terraform-aws-fargate-container?ref=v0.0.4"
}

dependency "ignore_me" {
  config_path = find_in_parent_folders("ignore_me")
}

dependency "other_ignore_me" {
  config_path = find_in_parent_folders("ignore_me")
}

locals {
  ignore_atlantis_dependencies = ["${find_in_parent_folders("ignore_me")}/terragrunt.hcl", "other_ignore_me"]
}
