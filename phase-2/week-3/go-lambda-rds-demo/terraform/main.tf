terraform {
  required_version = ">= 1.6.0"
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 6.0" }
  }
}
provider "aws" { region = var.region }

data "aws_availability_zones" "available" { state = "available" }
data "aws_caller_identity" "current" {}

resource "aws_vpc" "main" {
  cidr_block = "10.40.0.0/16"
  enable_dns_support = true
  enable_dns_hostnames = true
  tags = { Name = var.project }
}

resource "aws_subnet" "private_a" {
  vpc_id = aws_vpc.main.id
  cidr_block = "10.40.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = { Name = "${var.project}-private-a" }
}
resource "aws_subnet" "private_b" {
  vpc_id = aws_vpc.main.id
  cidr_block = "10.40.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  tags = { Name = "${var.project}-private-b" }
}

resource "aws_security_group" "lambda" {
  name = "${var.project}-lambda-sg"
  vpc_id = aws_vpc.main.id
  egress { from_port = 0 to_port = 0 protocol = "-1" cidr_blocks = ["0.0.0.0/0"] }
}
resource "aws_security_group" "rds" {
  name = "${var.project}-rds-sg"
  vpc_id = aws_vpc.main.id
  ingress {
    from_port = 5432
    to_port = 5432
    protocol = "tcp"
    security_groups = [aws_security_group.lambda.id]
  }
  egress { from_port = 0 to_port = 0 protocol = "-1" cidr_blocks = ["0.0.0.0/0"] }
}

resource "aws_db_subnet_group" "db" {
  name = "${var.project}-db-subnets"
  subnet_ids = [aws_subnet.private_a.id, aws_subnet.private_b.id]
}
resource "aws_db_instance" "postgres" {
  identifier = var.project
  engine = "postgres"
  instance_class = "db.t4g.micro"
  allocated_storage = 20
  storage_type = "gp3"
  db_name = var.db_name
  username = var.db_username
  password = var.db_password
  db_subnet_group_name = aws_db_subnet_group.db.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible = false
  skip_final_snapshot = true
  deletion_protection = false
  backup_retention_period = 0
}

resource "aws_iam_role" "lambda" {
  name = "${var.project}-lambda-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{Effect="Allow", Principal={Service="lambda.amazonaws.com"}, Action="sts:AssumeRole"}]
  })
}
resource "aws_iam_role_policy_attachment" "basic" {
  role = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}
resource "aws_iam_role_policy_attachment" "vpc" {
  role = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

resource "aws_lambda_function" "app" {
  function_name = var.project
  filename = "${path.module}/../function.zip"
  source_code_hash = filebase64sha256("${path.module}/../function.zip")
  role = aws_iam_role.lambda.arn
  runtime = "provided.al2023"
  handler = "bootstrap"
  architectures = ["x86_64"]
  timeout = 15
  memory_size = 256
  vpc_config {
    subnet_ids = [aws_subnet.private_a.id, aws_subnet.private_b.id]
    security_group_ids = [aws_security_group.lambda.id]
  }
  environment {
    variables = {
      DB_HOST = aws_db_instance.postgres.address
      DB_PORT = tostring(aws_db_instance.postgres.port)
      DB_NAME = var.db_name
      DB_USER = var.db_username
      DB_PASSWORD = var.db_password
      DB_SSLMODE = "require"
    }
  }
  depends_on = [aws_iam_role_policy_attachment.basic, aws_iam_role_policy_attachment.vpc]
}

resource "aws_lambda_function_url" "app" {
  function_name = aws_lambda_function.app.function_name
  authorization_type = "NONE"
}

resource "aws_lambda_permission" "function_url_public" {
  statement_id = "AllowPublicFunctionURL"
  action = "lambda:InvokeFunctionUrl"
  function_name = aws_lambda_function.app.function_name
  principal = "*"
  function_url_auth_type = "NONE"
}

resource "aws_lambda_permission" "function_url_invoke_public" {
  statement_id = "AllowPublicInvokeViaFunctionURL"
  action = "lambda:InvokeFunction"
  function_name = aws_lambda_function.app.function_name
  principal = "*"
  invoked_via_function_url = true
}

output "function_url" { value = aws_lambda_function_url.app.function_url }
output "rds_endpoint" { value = aws_db_instance.postgres.address }
output "lambda_security_group" { value = aws_security_group.lambda.id }
output "rds_security_group" { value = aws_security_group.rds.id }
