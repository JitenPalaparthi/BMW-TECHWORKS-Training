from flask import Flask, jsonify

app = Flask(__name__)


@app.get("/health")
def health():
    return jsonify({
        "message": "Python service is healthy"
    })


@app.get("/hello")
def hello():
    return jsonify({
        "message": "Hello from Python container on ECS"
    })


if __name__ == "__main__":
    app.run(
        host="0.0.0.0",
        port=8080
    )