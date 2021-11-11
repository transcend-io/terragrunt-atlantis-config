# https://www.terraform.io/docs/language/modules/sources.html#s3-bucket
terraform {
  source = "s3::https://s3-eu-west-1.amazonaws.com/examplecorp-terraform-modules/vpc.zip"
}

inputs = {
  foo = "bar"
}
