# Architecture walkthrough

## Request path

1. Client calls the public demo Lambda Function URL.
2. AWS invokes the Go Lambda.
3. Lambda runs with ENIs in the two configured VPC subnets.
4. Lambda resolves the private RDS endpoint through VPC DNS.
5. Lambda-SG opens TCP 5432 to RDS-SG; RDS-SG accepts 5432 only from Lambda-SG.
6. `pgxpool` executes PostgreSQL statements and returns JSON to the client.

## IAM vs database authorization

IAM is used for the Lambda execution environment: CloudWatch logging and VPC network-interface operations. PostgreSQL access in this demo is username/password based, so SQL authorization is performed by PostgreSQL rather than IAM.

## Cold start

On cold start, the process creates one `pgxpool.Pool`, pings RDS, and idempotently creates the `users` table. Warm invocations reuse the same process and connection pool.

## Production variant

A stronger production design places the secret in Secrets Manager and commonly inserts RDS Proxy between Lambda and RDS. The request flow becomes Lambda -> RDS Proxy -> RDS, reducing direct connection pressure during concurrency bursts.
