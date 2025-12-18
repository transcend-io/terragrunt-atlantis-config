variable "role_arn" {
  description = "Name of the role we want to bring in"
  type        = string
}

output "role_arn" {
  value = var.role_arn
}
