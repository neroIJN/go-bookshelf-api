package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestGetBook(t *testing.T) {
	// Setup test data
	books = []Book{
		{
			BookID:          "test-book-id",
			AuthorID:        "test-author-id",
			PublisherID:     "test-publisher-id",
			Title:           "Test Book",
			PublicationDate: "2023-01-01",
			ISBN:            "1234567890",
			Pages:           100,
			Genre:           "Test",
			Description:     "A test book",
			Price:           9.99,
			Quantity:        5,
		},
	}

	// Create a new request
	req, err := http.NewRequest("GET", "/books/test-book-id", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create a router with the test handler
	router := mux.NewRouter()
	router.HandleFunc("/books/{id}", getBook).Methods("GET")

	// Serve the request
	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var response Book
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response body: %v", err)
	}

	if response.BookID != "test-book-id" {
		t.Errorf("handler returned unexpected body: got %v want %v", response.BookID, "test-book-id")
	}
}
