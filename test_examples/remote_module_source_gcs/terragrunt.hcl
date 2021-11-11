# https://www.terraform.io/docs/language/modules/sources.html#gcs-bucket
terraform {
  source = "gcs::https://www.googleapis.com/storage/v1/modules/foomodule.zip"
}

inputs = {
  foo = "bar"
}
