# https://www.terraform.io/docs/language/modules/sources.html#github
terraform {
  source = "git@github.com:hashicorp/example.git"
}

inputs = {
  foo = "bar"
}
