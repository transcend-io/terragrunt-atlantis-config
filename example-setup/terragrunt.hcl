# Parent file without any source/dependencies

terraform_version_constraint = ">= 0.12, < 0.13"

terraform {
  extra_arguments "retry_lock" {
    commands  = get_terraform_commands_that_need_locking()
    arguments = ["-lock-timeout=5m"]
  }
}

inputs = {
  foo = "bar"
}
