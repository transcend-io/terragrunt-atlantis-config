terraform {
  source = "git::git@github.com:transcend-io/terraform-aws-fargate-container?ref=v0.0.4"
}

include "some_include" {
  path = find_in_parent_folders("include.hcl")
}

inputs = {
  included_foo = include.some_include.inputs.foo
}
