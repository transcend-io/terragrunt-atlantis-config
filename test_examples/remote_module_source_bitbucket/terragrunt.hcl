# https://www.terraform.io/docs/language/modules/sources.html#bitbucket
terraform {
  source = "bitbucket.org/hashicorp/tf-test-git"
}

inputs = {
  foo = "bar"
}
