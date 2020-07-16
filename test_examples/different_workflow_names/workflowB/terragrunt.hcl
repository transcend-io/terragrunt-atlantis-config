terraform {
  source = "git::git@github.com:transcend-io/terraform-aws-fargate-container?ref=v0.0.4"
}

locals {
  atlantis_workflow = "workflowB"
}

inputs = {
  foo = "bar"
}