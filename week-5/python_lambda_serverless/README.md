# Python AWS Lambda - Serverless Demo

This package is ready to upload as a ZIP deployment package to AWS Lambda.

## Handler

Set the Lambda handler to:

lambda_function.lambda_handler

## Runtime

Use a current AWS Lambda Python runtime, for example Python 3.12 or later.

## Environment variables

Optional:

- APP_NAME=python-serverless-demo
- LOG_LEVEL=INFO

## Supported routes

The function works with AWS Lambda Function URLs and API Gateway proxy events.

### GET /health

Returns service health and timestamp.

### GET /hello?name=Jiten

Returns a greeting.

### POST /echo

JSON body example:

{
  "message": "Hello Lambda"
}

Returns the supplied JSON.

## Deploy from AWS Console

1. Open AWS Lambda.
2. Create function.
3. Choose "Author from scratch".
4. Choose Python runtime.
5. Create the function.
6. Open Code.
7. Upload from -> .zip file.
8. Upload `python_lambda_serverless.zip`.
9. Ensure handler is `lambda_function.lambda_handler`.
10. Deploy.
11. Add a Function URL or API Gateway trigger.

## Example Function URL tests

Health:

curl "https://YOUR_FUNCTION_URL/health"

Greeting:

curl "https://YOUR_FUNCTION_URL/hello?name=Jiten"

POST:

curl -X POST \
  "https://YOUR_FUNCTION_URL/echo" \
  -H "Content-Type: application/json" \
  -d '{"message":"Hello from my Mac"}'

## Notes

This demo has no external Python dependencies, so no `pip install` step is required.
