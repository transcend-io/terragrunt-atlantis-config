locals {
  extra_atlantis_dependencies = [
    # A relative file to the child should work
    "some_parent_dep",

    # Functions should run from the child dir, not the parent dir
    find_in_parent_folders("file_in_parent_of_child.json"),
    "${get_parent_terragrunt_dir()}/folder_under_parent/common_tags.hcl",
  ]
}