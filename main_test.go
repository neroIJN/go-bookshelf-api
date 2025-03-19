package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/books", createBook)
	return r
}

func TestCreateBook(t *testing.T) {
	router := setupRouter()

	book := Book{
		AuthorID:    "123",
		PublisherID: "456",
		Title:       "Test Book",
		Description: "A sample test book",
		Pages:       100,
		Price:       9.99,
		Quantity:    5,
	}
	body, _ := json.Marshal(book)

	req, _ := http.NewRequest("POST", "/books", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)
}
