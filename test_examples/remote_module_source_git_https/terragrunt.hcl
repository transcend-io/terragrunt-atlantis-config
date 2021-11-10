# https://www.terraform.io/docs/language/modules/sources.html#generic-git-repository
terraform {
  source = "git::https://example.com/vpc.git"
}

inputs = {
  foo = "bar"
}
