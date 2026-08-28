# Terraform Docker Module Example

This project demonstrates Terraform variables, modules, resources, and outputs by creating an Nginx Docker container.

## Prerequisites

- Terraform installed
- Docker installed and running

Verify:

```bash
terraform version
docker version
```

## Structure

```text
terraform-docker-modules/
├── main.tf
├── variables.tf
├── outputs.tf
├── terraform.tfvars
└── modules/
    └── docker-container/
        ├── main.tf
        ├── variables.tf
        └── outputs.tf
```

## Run

From this directory:

```bash
terraform init
terraform fmt -recursive
terraform validate
terraform plan
terraform apply
```

Type `yes` when prompted.

Verify:

```bash
docker ps
terraform output
```

Open:

http://localhost:8080

## Change configuration

Edit `terraform.tfvars`, for example change the external port, then run:

```bash
terraform plan
terraform apply
```

## Destroy

```bash
terraform destroy
```

Type `yes` when prompted.

## Concepts

- `variable`: input supplied to Terraform.
- `local`: internal reusable/calculated value (not required by this small module).
- `resource`: infrastructure Terraform manages.
- `module`: reusable Terraform configuration.
- `output`: a value exposed by Terraform/module after evaluation.
