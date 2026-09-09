import json


def lambda_handler(event, context):
    print("Lambda invoked")

    response = {
        "message": "Hello from Python AWS Lambda"
    }

    return {
        "statusCode": 200,
        "headers": {
            "Content-Type": "application/json"
        },
        "body": json.dumps(response)
    }