# https://www.terraform.io/docs/language/modules/sources.html#github
terraform {
  source = "github.com/hashicorp/example"
}

inputs = {
  foo = "bar"
}
