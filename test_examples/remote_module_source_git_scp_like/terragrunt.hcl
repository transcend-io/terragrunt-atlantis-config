# https://www.terraform.io/docs/language/modules/sources.html#quot-scp-like-quot-address-syntax
terraform {
  source = "git::git@example.com:storage.git"
}

inputs = {
  foo = "bar"
}
