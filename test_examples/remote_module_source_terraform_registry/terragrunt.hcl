# https://www.terraform.io/docs/language/modules/sources.html#terraform-registry
terraform {
  source = "tfr:///terraform-aws-modules/vpc/aws?version=3.7.0"
}

inputs = {
  foo = "bar"
}
