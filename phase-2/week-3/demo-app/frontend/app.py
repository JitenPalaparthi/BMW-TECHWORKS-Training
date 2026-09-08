from flask import Flask, render_template
import requests
import os

app = Flask(__name__)

BACKEND_URL = os.getenv(
    "BACKEND_URL",
    "http://localhost:5001"
)

@app.route("/")
def home():
    try:
        response = requests.get(
            f"{BACKEND_URL}/api/message",
            timeout=3
        )

        data = response.json()
        message = data["message"]

    except Exception as e:
        message = f"Backend unavailable: {e}"

    return render_template(
        "index.html",
        message=message
    )

@app.route("/health")
def health():
    return {"status": "ok"}, 200

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000)