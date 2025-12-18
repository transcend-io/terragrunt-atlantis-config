include {
  path = find_in_parent_folders()
}

terraform {
  source = "${find_in_parent_folders()}/../../../modules/subdir/resource"
}

inputs = {
  foo = "bar"
}
