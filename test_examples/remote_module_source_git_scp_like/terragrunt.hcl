# https://www.terraform.io/docs/language/modules/sources.html#quot-scp-like-quot-address-syntax
terraform {
  source = "git::username@example.com:storage.git"
}

inputs = {
  foo = "bar"
}
