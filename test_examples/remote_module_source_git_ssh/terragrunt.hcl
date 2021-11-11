# https://www.terraform.io/docs/language/modules/sources.html#generic-git-repository
terraform {
  source = "git::ssh://username@example.com/storage.git"
}

inputs = {
  foo = "bar"
}
