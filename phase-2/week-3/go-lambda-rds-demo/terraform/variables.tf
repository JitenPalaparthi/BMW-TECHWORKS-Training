variable "region" { type = string default = "ap-south-1" }
variable "project" { type = string default = "go-lambda-rds-demo" }
variable "db_name" { type = string default = "appdb" }
variable "db_username" { type = string default = "appuser" }
variable "db_password" { type = string sensitive = true }
