import pytest
import requests

BASE_URL = "http://localhost:8080"


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

    # Assert correct status code (201 Created or 200 OK depending on your API spec)
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

    assert response.status_code in (500,400, 422)