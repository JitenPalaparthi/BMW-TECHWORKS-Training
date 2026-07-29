import subprocess
import time
import pytest
import requests

BASE_URL = "http://localhost:8080"


@pytest.fixture(scope="function", autouse=True)
def docker_compose_service():
    """Brings up Docker Compose before each test, waits for the API to be ready,

    and tears down the containers after each test completes.
    """
    # 1. Start Docker Compose in detached mode before the test
    subprocess.run(
        ["docker", "compose", "up", "-d"],
        check=True,
    )

    # 2. Health check: Wait for the API server to be responsive
    timeout = 30  # max wait time in seconds
    start_time = time.time()
    api_ready = False

    while time.time() - start_time < timeout:
        try:
            res = requests.get(f"{BASE_URL}/api/products", timeout=2)
            if res.status_code < 500:
                api_ready = True
                break
        except requests.exceptions.RequestException:
            pass
        time.sleep(1)

    if not api_ready:
        subprocess.run(["docker", "compose", "down", "-v"])
        pytest.fail("Docker service failed to start or respond within timeout.")

    # Yield control to execute the test function
    yield

    # 3. Tear down containers and volumes immediately after the test finishes
    subprocess.run(
        ["docker", "compose", "down", "-v"],
        check=True,
    )


@pytest.fixture
def valid_product_payload():
    """Fixture providing a valid product payload."""
    return {
        "name": "Apple Macbook M2",
        "description": "Apple Bionic Chip M2 with ARM Processor",
        "price": 150000,
        "quantity": 20,
    }


def test_create_product_success(valid_product_payload):
    """Test successful creation of a product (HTTP 201/200)."""
    response = requests.post(
        f"{BASE_URL}/api/products",
        json=valid_product_payload,
        headers={"Content-Type": "application/json"},
    )

    # Assert correct status code
    assert response.status_code in (200, 201), f"Unexpected status code: {response.status_code}"

    # Parse JSON response
    data = response.json()

    # Validate response structure and returned values
    assert "id" in data or "productId" in data, "Response should include a product identifier"
    assert data["name"] == valid_product_payload["name"]
    assert data["price"] == valid_product_payload["price"]
    assert data["quantity"] == valid_product_payload["quantity"]


def test_create_product_missing_required_fields():
    """Test creating a product with missing required fields (HTTP 400 Bad Request)."""
    incomplete_payload = {
        "name": "Incomplete Product"
        # Missing price and quantity
    }

    response = requests.post(
        f"{BASE_URL}/api/products",
        json=incomplete_payload,
    )

    assert response.status_code == 400


def test_create_product_invalid_data_types():
    """Test payload validation for invalid data types (e.g., negative price or non-numeric quantity)."""
    invalid_payload = {
        "name": "Invalid Price Product",
        "description": "Test",
        "price": -500,
        "quantity": "twenty",
    }

    response = requests.post(
        f"{BASE_URL}/api/products",
        json=invalid_payload,
    )

    assert response.status_code in (400, 422, 500)