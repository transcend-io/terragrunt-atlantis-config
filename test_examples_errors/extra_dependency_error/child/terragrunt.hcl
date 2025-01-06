terraform {
  source = "git::git@github.com:transcend-io/terraform-aws-fargate-container?ref=v0.0.4"
}

locals {
  tg_config = read_terragrunt_config(find_in_parent_folders("config.hcl"))

  extra_atlantis_dependencies = [
    "some_extra_dep0",
    "some_extra_dep1",
    "some_extra_dep2",
    "some_extra_dep3",
    local.tg_config
  ]
}