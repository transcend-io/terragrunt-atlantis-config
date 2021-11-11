# https://www.terraform.io/docs/language/modules/sources.html#http-urls
terraform {
  source = "http://example.com/vpc-module.zip"
}

inputs = {
  foo = "bar"
}
