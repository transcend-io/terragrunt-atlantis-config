include {
  path = find_in_parent_folders()
}

terraform {
  source = "git::git@github.com:transcend-io/terraform-aws-fargate-container?ref=v0.0.4"
}

locals {
  atlantis_silence_pr_comments = ["plan"]
}

inputs = {
  foo = "bar"
}
