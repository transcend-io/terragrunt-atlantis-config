terraform {
  source = "git::git@github.com:transcend-io/terraform-aws-fargate-container?ref=v0.0.4"
}

locals {
  atlantis_project_name = "project-a"
}

inputs = {
  foo = "bar"
}
