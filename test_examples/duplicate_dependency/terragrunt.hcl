terraform {
  extra_arguments "common" {
    commands = get_terraform_commands_that_need_vars()
    optional_var_files = [
      "${get_terragrunt_dir()}/../shared_vars.hcl"
    ]
  }
}
