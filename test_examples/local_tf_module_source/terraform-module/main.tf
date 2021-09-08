resource "some_resource" "some_name" {
  foo = "bar"
}

module "nested_module" {
  source = "./nested-module"
}

module "another_module" {
  source = "../terraform-another-module"
}
