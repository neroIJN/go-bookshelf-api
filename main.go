package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Book struct {
	BookID          string  `json:"bookId"`
	AuthorID        string  `json:"authorId"`
	PublisherID     string  `json:"publisherId"`
	Title           string  `json:"title"`
	PublicationDate string  `json:"publicationDate"`
	ISBN            string  `json:"isbn"`
	Pages           int     `json:"pages"`
	Genre           string  `json:"genre"`
	Description     string  `json:"description"`
	Price           float64 `json:"price"`
	Quantity        int     `json:"quantity"`
}

var books []Book
var fileName = "books.json"

func loadBooks() {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("No existing data found, starting fresh.")
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err == nil {
		json.Unmarshal(data, &books)
	}
}

func saveBooks() {
	data, _ := json.MarshalIndent(books, "", "  ")
	_ = ioutil.WriteFile(fileName, data, 0644)
}

// Get All Books
func getBooks(c *gin.Context) {
	c.JSON(http.StatusOK, books)
}

// Create a Book
func createBook(c *gin.Context) {
	var newBook Book
	if err := c.BindJSON(&newBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newBook.BookID = uuid.New().String()
	books = append(books, newBook)
	saveBooks()
	c.JSON(http.StatusCreated, newBook)
}

// Get a Book by ID
func getBookByID(c *gin.Context) {
	id := c.Param("id")
	for _, book := range books {
		if book.BookID == id {
			c.JSON(http.StatusOK, book)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"message": "Book not found"})
}

// Update a Book
func updateBook(c *gin.Context) {
	id := c.Param("id")
	var updatedBook Book

	if err := c.BindJSON(&updatedBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i, book := range books {
		if book.BookID == id {
			updatedBook.BookID = id
			books[i] = updatedBook
			saveBooks()
			c.JSON(http.StatusOK, updatedBook)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"message": "Book not found"})
}

// Delete a Book
func deleteBook(c *gin.Context) {
	id := c.Param("id")
	for i, book := range books {
		if book.BookID == id {
			books = append(books[:i], books[i+1:]...)
			saveBooks()
			c.JSON(http.StatusOK, gin.H{"message": "Book deleted"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"message": "Book not found"})
}

// Search Books
func searchBooks(c *gin.Context) {
	query := strings.ToLower(c.Query("q"))
	var results []Book
	var wg sync.WaitGroup
	resultChannel := make(chan Book, len(books))

	for _, book := range books {
		wg.Add(1)
		go func(b Book) {
			defer wg.Done()
			if strings.Contains(strings.ToLower(b.Title), query) || strings.Contains(strings.ToLower(b.Description), query) {
				resultChannel <- b
			}
		}(book)
	}

	wg.Wait()
	close(resultChannel)

	for b := range resultChannel {
		results = append(results, b)
	}

	c.JSON(http.StatusOK, results)
}

func main() {
	loadBooks()
	r := gin.Default()

	r.GET("/books", getBooks)
	r.POST("/books", createBook)
	r.GET("/books/:id", getBookByID)
	r.PUT("/books/:id", updateBook)
	r.DELETE("/books/:id", deleteBook)
	r.GET("/books/search", searchBooks)

	log.Fatal(r.Run(":3000"))
}
