dependency "roles" {
  config_path = "../roles"
  mock_outputs = {
    role_arn = "placeholder"
  }
}

inputs = {
  role_arn = dependency.roles.outputs.role_arn
}
