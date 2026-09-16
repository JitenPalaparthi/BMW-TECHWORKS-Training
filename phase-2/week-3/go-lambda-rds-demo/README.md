# Go Lambda + RDS PostgreSQL + VPC + IAM + Environment Variables

A complete teaching/demo project that deploys:

```text
Internet client
     |
     v
Lambda Function URL
     |
     v
Go Lambda (private VPC subnets, Lambda SG)
     |
     | TCP 5432 allowed only Lambda-SG -> RDS-SG
     v
RDS PostgreSQL (private, not publicly accessible)
```

## What is included

- Go AWS Lambda using `provided.al2023`
- Lambda Function URL routes: `GET /health`, `POST /users`, `GET /users`
- PostgreSQL RDS instance
- VPC with two private subnets in different Availability Zones
- Separate Lambda and RDS security groups
- Lambda IAM execution role with CloudWatch Logs and VPC ENI permissions
- Lambda environment variables for DB connection settings
- SQL schema
- Unit tests with an in-memory fake store
- `curl` API smoke-test script
- Terraform deployment and destroy scripts

> Demo security note: this project intentionally puts `DB_PASSWORD` in a Lambda environment variable because the exercise explicitly demonstrates environment variables. For production, prefer AWS Secrets Manager, password rotation, and often RDS Proxy for Lambda connection pooling. Also protect the Function URL with IAM/API Gateway authentication instead of `NONE`.

## Prerequisites

- AWS account and credentials configured (`aws sts get-caller-identity` should succeed)
- Go 1.24+
- Terraform 1.6+
- `zip`
- `curl`
- Optional: `psql` if initializing the DB manually

## 1. Configure database password

```bash
cd terraform
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars` and set a strong password.

## 2. Run unit tests

```bash
./scripts/unit-test.sh
```

## 3. Build Lambda ZIP

```bash
./scripts/build.sh
```

The binary is built for Linux/x86_64 and packaged as `function.zip`, with `bootstrap` at the ZIP root.

## 4. Deploy AWS infrastructure

```bash
./scripts/deploy.sh
```

Terraform creates the VPC, subnets, SGs, RDS, IAM role, Lambda, and public demo Function URL.

Get outputs:

```bash
cd terraform
terraform output
FUNCTION_URL=$(terraform output -raw function_url)
echo "$FUNCTION_URL"
```

## 5. Database schema initialization

The Lambda performs an idempotent `CREATE TABLE IF NOT EXISTS users (...)` during cold start, so the deployed demo is self-initializing. The same DDL is also provided in `sql/schema.sql` for teaching or manual administration.

Because RDS is deliberately private, a laptop cannot directly connect unless it has network access to the VPC. From an EC2/SSM/bastion host that can reach RDS:

```bash
export DB_HOST=$(terraform output -raw rds_endpoint)
export DB_PORT=5432
export DB_NAME=appdb
export DB_USER=appuser
export DB_PASSWORD='your-password'
export DB_SSLMODE=require
../scripts/init-db.sh
```

Or directly:

```bash
PGPASSWORD="$DB_PASSWORD" psql \
  "host=$DB_HOST port=5432 dbname=appdb user=appuser sslmode=require" \
  -f ../sql/schema.sql
```

## 6. Test the Lambda URL

```bash
FUNCTION_URL=$(terraform output -raw function_url)
../scripts/api-test.sh "$FUNCTION_URL"
```

Individual calls:

```bash
curl "$FUNCTION_URL/health"
```

```bash
curl -X POST "$FUNCTION_URL/users" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Jiten","email":"jiten@example.com"}'
```

```bash
curl "$FUNCTION_URL/users"
```

## 7. Test through AWS CLI

The repository also includes `scripts/lambda-test-event.json`.

```bash
aws lambda invoke \
  --function-name go-lambda-rds-demo \
  --cli-binary-format raw-in-base64-out \
  --payload fileb://scripts/lambda-test-event.json \
  response.json

cat response.json
```

## Environment variables

Lambda receives:

```text
DB_HOST      RDS private DNS endpoint
DB_PORT      5432
DB_NAME      appdb
DB_USER      appuser
DB_PASSWORD  demo password (use Secrets Manager in production)
DB_SSLMODE   require
```

The Go application converts these into a PostgreSQL URL and creates a `pgxpool.Pool` outside the request handler, allowing warm Lambda invocations to reuse the pool.

## IAM

Two AWS-managed policies are attached to the Lambda execution role:

```text
AWSLambdaBasicExecutionRole
AWSLambdaVPCAccessExecutionRole
```

The first allows Lambda logging to CloudWatch Logs. The second provides the EC2 network-interface permissions required for VPC-attached Lambda functions.

The application does **not** need IAM permission to query PostgreSQL when using username/password database authentication. Network reachability is controlled through VPC routing and security groups; PostgreSQL authenticates the DB credentials.

## VPC/security-group flow

```text
Lambda
  subnet: 10.40.1.0/24 or 10.40.2.0/24
  SG: lambda-sg
       |
       | TCP 5432
       v
RDS
  private DB subnet group
  SG: rds-sg
  inbound source = lambda-sg only
```

RDS is configured with `publicly_accessible = false`.

## Why the Lambda has no NAT Gateway here

The Lambda only needs to talk to RDS inside the VPC and CloudWatch logging is handled by the Lambda service integration/execution environment. This demo does not make outbound internet/API calls from application code. If your VPC-attached function needs arbitrary internet egress, add a NAT Gateway (or appropriate VPC endpoints for AWS services).

## Production improvements

- Store credentials in AWS Secrets Manager instead of plain Lambda environment variables.
- Use RDS Proxy for high-concurrency Lambda workloads to reduce connection pressure.
- Protect the HTTP endpoint using IAM authorization or API Gateway/Cognito/JWT.
- Enable RDS backups, deletion protection, Multi-AZ where appropriate.
- Use migrations (Flyway, Goose, Atlas, etc.) in CI/CD instead of manual schema initialization.
- Add structured logging, tracing, metrics and alarms.
- Tighten IAM to least privilege and pin Terraform/provider/tool versions according to your organization.

## Cleanup

RDS incurs cost. Destroy the demo when finished:

```bash
./scripts/destroy.sh
```
