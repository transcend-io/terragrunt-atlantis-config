# https://www.terraform.io/docs/language/modules/sources.html#http-urls
terraform {
  source = "hg::http://example.com/vpc.hg"
}

inputs = {
  foo = "bar"
}
