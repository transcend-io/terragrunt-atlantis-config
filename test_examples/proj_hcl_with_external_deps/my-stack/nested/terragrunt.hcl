
include "stack" {
  path = find_in_parent_folders("stack.hcl")
}
include "env" {
  path   = "${get_terragrunt_dir()}/../../_env/cluster/fargate.hcl"
  expose = true
}

terraform {
  source = include.env.locals.source_base_url
}



