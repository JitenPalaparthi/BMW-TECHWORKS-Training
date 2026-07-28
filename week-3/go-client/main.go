package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func main() {

	url := "http://localhost:8080/api/products"
	method := "POST"

	payload := strings.NewReader(`{
    "name": "Apple Macbook M5",
    "description": "Apple Bionioc Chip M5 with ARM Processor",
    "price": 251000.00,
    "quantity": 50
}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
