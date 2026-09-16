import json
import os
import uuid
from datetime import datetime, timezone

import boto3
from botocore.exceptions import ClientError
from flask import Flask, jsonify, request


# ---------------------------------------------------------
# Elastic Beanstalk expects a WSGI application
# ---------------------------------------------------------

application = Flask(__name__)

# Optional alias for local development
app = application


# ---------------------------------------------------------
# Configuration
# ---------------------------------------------------------

S3_BUCKET_NAME = os.getenv("S3_BUCKET_NAME", "demo-users-2026")

s3_client = boto3.client("s3")


# ---------------------------------------------------------
# Utility functions
# ---------------------------------------------------------

def generate_id():
    return uuid.uuid4().hex[:16]


def user_key(user_id):
    return f"users/{user_id}.json"


def save_user(user):
    key = user_key(user["id"])

    s3_client.put_object(
        Bucket=S3_BUCKET_NAME,
        Key=key,
        Body=json.dumps(user, indent=2).encode("utf-8"),
        ContentType="application/json"
    )


def read_user(user_id):
    key = user_key(user_id)

    try:
        response = s3_client.get_object(
            Bucket=S3_BUCKET_NAME,
            Key=key
        )

        content = response["Body"].read()

        return json.loads(content)

    except ClientError as e:

        error_code = e.response["Error"]["Code"]

        if error_code in ("NoSuchKey", "404"):
            return None

        raise


# ---------------------------------------------------------
# Root endpoint
# ---------------------------------------------------------

@application.route("/", methods=["GET"])
def root():
    return jsonify({
        "message": "Flask + AWS S3 User API"
    })


# ---------------------------------------------------------
# Health endpoint
# ---------------------------------------------------------

@application.route("/health", methods=["GET"])
def health():
    return jsonify({
        "status": "UP"
    })


# ---------------------------------------------------------
# CREATE USER
#
# POST /api/users
# ---------------------------------------------------------

@application.route("/api/users", methods=["POST"])
def create_user():

    data = request.get_json()

    if not data:
        return jsonify({
            "error": "JSON body is required"
        }), 400

    name = data.get("name")
    email = data.get("email")

    if not name:
        return jsonify({
            "error": "name is required"
        }), 400

    if not email:
        return jsonify({
            "error": "email is required"
        }), 400

    now = datetime.now(timezone.utc).isoformat()

    user = {
        "id": generate_id(),
        "name": name,
        "email": email,
        "created_at": now,
        "updated_at": now
    }

    try:

        save_user(user)

        return jsonify(user), 201

    except ClientError as e:

        application.logger.exception(
            "Failed to save user to S3"
        )

        return jsonify({
            "error": "failed to save user to S3",
            "details": str(e)
        }), 500


# ---------------------------------------------------------
# GET ALL USERS
#
# GET /api/users
# ---------------------------------------------------------

@application.route("/api/users", methods=["GET"])
def get_users():

    users = []

    try:

        paginator = s3_client.get_paginator("list_objects_v2")

        pages = paginator.paginate(
            Bucket=S3_BUCKET_NAME,
            Prefix="users/"
        )

        for page in pages:

            for obj in page.get("Contents", []):

                key = obj["Key"]

                if not key.endswith(".json"):
                    continue

                user_id = key.removeprefix("users/")
                user_id = user_id.removesuffix(".json")

                user = read_user(user_id)

                if user:
                    users.append(user)

        return jsonify(users)

    except ClientError as e:

        application.logger.exception(
            "Failed to list users"
        )

        return jsonify({
            "error": "failed to list users",
            "details": str(e)
        }), 500


# ---------------------------------------------------------
# GET ONE USER
#
# GET /api/users/<id>
# ---------------------------------------------------------

@application.route("/api/users/<user_id>", methods=["GET"])
def get_user(user_id):

    try:

        user = read_user(user_id)

        if user is None:

            return jsonify({
                "error": "user not found"
            }), 404

        return jsonify(user)

    except ClientError as e:

        application.logger.exception(
            "Failed to read user"
        )

        return jsonify({
            "error": "failed to read user from S3",
            "details": str(e)
        }), 500


# ---------------------------------------------------------
# UPDATE USER
#
# PUT /api/users/<id>
# ---------------------------------------------------------

@application.route("/api/users/<user_id>", methods=["PUT"])
def update_user(user_id):

    data = request.get_json()

    if not data:

        return jsonify({
            "error": "JSON body is required"
        }), 400

    try:

        existing_user = read_user(user_id)

        if existing_user is None:

            return jsonify({
                "error": "user not found"
            }), 404

        name = data.get("name")
        email = data.get("email")

        if not name:

            return jsonify({
                "error": "name is required"
            }), 400

        if not email:

            return jsonify({
                "error": "email is required"
            }), 400

        existing_user["name"] = name
        existing_user["email"] = email
        existing_user["updated_at"] = (
            datetime.now(timezone.utc).isoformat()
        )

        save_user(existing_user)

        return jsonify(existing_user)

    except ClientError as e:

        application.logger.exception(
            "Failed to update user"
        )

        return jsonify({
            "error": "failed to update user",
            "details": str(e)
        }), 500


# ---------------------------------------------------------
# DELETE USER
#
# DELETE /api/users/<id>
# ---------------------------------------------------------

@application.route(
    "/api/users/<user_id>",
    methods=["DELETE"]
)
def delete_user(user_id):

    try:

        existing_user = read_user(user_id)

        if existing_user is None:

            return jsonify({
                "error": "user not found"
            }), 404

        key = user_key(user_id)

        s3_client.delete_object(
            Bucket=S3_BUCKET_NAME,
            Key=key
        )

        return jsonify({
            "message": "user deleted successfully",
            "id": user_id
        })

    except ClientError as e:

        application.logger.exception(
            "Failed to delete user"
        )

        return jsonify({
            "error": "failed to delete user",
            "details": str(e)
        }), 500


# ---------------------------------------------------------
# Local development only
# ---------------------------------------------------------

if __name__ == "__main__":

    port = int(os.getenv("PORT", "5001"))

    application.run(
        host="0.0.0.0",
        port=port,
        debug=False
    )