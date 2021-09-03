module "some_module" {
  source = "../terraform-module"
}

module "another_module" {
  source = "git::git@github.com:transcend-io/terraform-aws-fargate-container?ref=v0.0.4"
}
