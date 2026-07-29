import http.client
import json

conn = http.client.HTTPConnection("localhost", 8080)
payload = json.dumps({
  "name": "Apple Macbook M2",
  "description": "Apple Bionioc Chip M2 with ARM Processor",
  "price": 150000,
  "quantity": 20
})
headers = {
  'Content-Type': 'application/json'
}
conn.request("POST", "/api/products", payload, headers)
res = conn.getresponse()
data = res.read()

print(data.decode("utf-8"))