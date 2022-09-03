remote_state {
  backend = "gcs"
  generate = {
    path      = "backend.tf"
    if_exists = "skip"
  }
  config = {
    bucket                      = "magical-iac-configuration"
    prefix                      = "terragrunt/${path_relative_to_include()}"
    project                     = "mars-007"
  }
}
