include {
  path = find_in_parent_folders("terragrunt.hcl")
}

terraform {
  source = "../terraform-module"
}