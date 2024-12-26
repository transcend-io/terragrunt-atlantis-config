terraform {
  source = "git::git@github.com:transcend-io/terraform-aws-fargate-container?ref=v0.0.4"
  extra_arguments "extra" {
    commands = get_terraform_commands_that_need_vars()
    optional_var_files = [
      "${get_terragrunt_dir()}/extra.tfvars",
    ]
  }
}

inputs = {
  foo = "bar"
}