terraform {
  source = "git::git@github.com:transcend-io/terraform-aws-fargate-container?ref=v0.0.4"
}

locals {
  extra_atlantis_dependencies = [
    "some_extra_dep",
    find_in_parent_folders("test_file.json")
  ]
}