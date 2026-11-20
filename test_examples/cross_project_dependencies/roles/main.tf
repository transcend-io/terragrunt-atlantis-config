resource "aws_iam_role" "test_role" {
  name = "test_role_terragrunt_atlantis_project"

  assume_role_policy = data.aws_iam_policy_document.policy.json
}

data "aws_iam_policy_document" "policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

output "role_name" {
  value = aws_iam_role.test_role.name
}
