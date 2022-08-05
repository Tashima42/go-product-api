package main_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/tashima42/go-product-api"
)

var a main.App

func TestMain(m *testing.M) {
	a.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"),
	)

	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/products", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetNonExistentProduct(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/products/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Product not found" {
		t.Errorf("Expected the 'error' key of the response to be 'Product not found'. Got %s", m["error"])
	}
}

func TestCreateProduct(t *testing.T) {
	clearTable()

	var jsonStr = []byte(`{"name": "test product", "price": 11.72}`)

	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(jsonStr))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["name"] != "test product" {
		t.Errorf("Expected 'error' to be 'test product'. Got %v", m["name"])
	}
	if m["price"] != 11.72 {
		t.Errorf("Expected 'price' to be 11.72. Got %v", m["price"])
	}
	if m["id"] != 1.0 {
		t.Errorf("Expected 'id' to be 1.0. Got %v", m["id"])
	}
}

func TestGetProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdateProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	reqGet, _ := http.NewRequest("GET", "/products/1", nil)
	responseGet := executeRequest(reqGet)

	var originalProduct map[string]interface{}
	json.Unmarshal(responseGet.Body.Bytes(), &originalProduct)

	var jsonStr = []byte(`{"name": "test product - updated name", "price": 11.22}`)
	req, _ := http.NewRequest("PUT", "/product/1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != originalProduct["name"] {
		t.Errorf("Expected 'name' to change from '%v' to 'test product - updated name'. Got %v", originalProduct["name"], m["name"])
	}
	if m["price"] != originalProduct["price"] {
		t.Errorf("Expected 'price' to change from '%v' to 11.22. Got %v", originalProduct["price"], m["price"])
	}
	if m["id"] == originalProduct["id"] {
		t.Errorf("Expected 'id' to remain the same (%v). Got %v", originalProduct["id"], m["id"])
	}
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	reqGetOk, _ := http.NewRequest("GET", "/product/1", nil)
	responseGetOk := executeRequest(reqGetOk)
	checkResponseCode(t, http.StatusOK, responseGetOk.Code)

	reqDelete, _ := http.NewRequest("DELETE", "/product/1", nil)
	responseDelete := executeRequest(reqDelete)
	checkResponseCode(t, http.StatusOK, responseDelete.Code)

	reqGetNotFound, _ := http.NewRequest("GET", "/product/1", nil)
	responseGetNotFound := executeRequest(reqGetNotFound)
	checkResponseCode(t, http.StatusNotFound, responseGetNotFound.Code)
}

func addProducts(count int) {
	if count < 1 {
		count = 1
	}
	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO products (name, price) values ($1, $2)", "Product "+strconv.Itoa(i), (i+1.0)*10)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected code %d. Got %d\n", expected, actual)
	}
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS products
(
    id SERIAL,
    name TEXT NOT NULL,
    price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    CONSTRAINT products_pkey PRIMARY KEY (id)
)`

func clearTable() {
	a.DB.Exec("DELETE FROM products")
	a.DB.Exec("ALTER SEQUENCE products_id_seq RESTART WITH 1")
}
