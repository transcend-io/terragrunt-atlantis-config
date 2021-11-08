# Set common variables for the environment. This is automatically pulled in in the root terragrunt.hcl configuration to
# feed forward to the child modules.
locals {
  environment = "qa"
  extra_atlantis_dependencies = [
    find_in_parent_folders("arbitrary.hcl"),
    "${dirname(find_in_parent_folders("account.hcl"))}/us-east-1/stage/**/*.hcl",
  ]
  atlantis_workflow = "anotherWorkflowSpecifiedInParent"
    atlantis_project = true
}
