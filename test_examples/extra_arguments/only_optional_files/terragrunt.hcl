include {
  path = find_in_parent_folders()
}

terraform {
  source = "git::git@github.com:transcend-io/terraform-aws-fargate-container?ref=v0.0.4"
  extra_arguments "conditional_vars" {
    commands = [
      "apply",
      "plan",
      "import",
      "push",
      "refresh"
    ]

    # Small note: get_env can't really be supported very well here.
    # In a case like this, I'd use a for loop to construct
    # `extra_atlantis_dependencies` in locals for all possible regions
    optional_var_files = [
      "${get_parent_terragrunt_dir()}/${get_env("TF_VAR_env", "dev")}.tfvars",
      "${get_parent_terragrunt_dir()}/${get_env("TF_VAR_region", "us-east-1")}.tfvars",
      "${get_terragrunt_dir()}/${get_env("TF_VAR_env", "dev")}.tfvars",
      "${get_terragrunt_dir()}/${get_env("TF_VAR_region", "us-east-1")}.tfvars"
    ]
  }
}

inputs = {
  foo = "bar"
}