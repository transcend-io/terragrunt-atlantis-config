terraform {
  source = "git::git@github.com:transcend-io/terraform-aws-fargate-container?ref=v0.0.4"
}

dependency "some_dep" {
  config_path = "../depender"
}

dependency "nested" {
  config_path = "./nested"
}

inputs = {
  foo = dependency.some_dep.outputs.some_output
}