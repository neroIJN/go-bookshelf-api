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

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Book represents a book entity
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

// Global variables
var books []Book
var dataFile = "books.json"
var mutex = &sync.RWMutex{}

// Helper function to save books to file
func saveBooks() error {
	mutex.Lock()
	defer mutex.Unlock()

	data, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dataFile, data, 0644)
}

// Helper function to load books from file
func loadBooks() error {
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		// If file doesn't exist, initialize with empty slice
		books = []Book{}
		return saveBooks()
	}

	data, err := ioutil.ReadFile(dataFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &books)
}

// GET /books - Return all books
func getBooks(w http.ResponseWriter, r *http.Request) {
	mutex.RLock()
	defer mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

// POST /books - Create a new book
func createBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var book Book
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Generate a new UUID if not provided
	if book.BookID == "" {
		book.BookID = uuid.New().String()
	}

	mutex.Lock()
	books = append(books, book)
	mutex.Unlock()

	err = saveBooks()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to save book"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

// GET /books/{id} - Return a single book by ID
func getBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	bookID := params["id"]

	mutex.RLock()
	defer mutex.RUnlock()

	for _, book := range books {
		if book.BookID == bookID {
			json.NewEncoder(w).Encode(book)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "Book not found"})
}

// PUT /books/{id} - Update a single book by ID
func updateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	bookID := params["id"]

	var updatedBook Book
	err := json.NewDecoder(r.Body).Decode(&updatedBook)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Ensure the book ID in the URL matches the book ID in the request body
	if updatedBook.BookID != bookID && updatedBook.BookID != "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Book ID mismatch"})
		return
	}

	updatedBook.BookID = bookID

	mutex.Lock()
	defer mutex.Unlock()

	found := false
	for i, book := range books {
		if book.BookID == bookID {
			books[i] = updatedBook
			found = true
			break
		}
	}

	if !found {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Book not found"})
		return
	}

	err = saveBooks()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update book"})
		return
	}

	json.NewEncoder(w).Encode(updatedBook)
}

// DELETE /books/{id} - Delete a single book by ID
func deleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	bookID := params["id"]

	mutex.Lock()
	defer mutex.Unlock()

	found := false
	for i, book := range books {
		if book.BookID == bookID {
			books = append(books[:i], books[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Book not found"})
		return
	}

	err := saveBooks()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete book"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Book deleted successfully"})
}

// GET /books/search?q=<keyword> - Search books by keyword in title and description
func searchBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := strings.ToLower(r.URL.Query().Get("q"))
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Search query is required"})
		return
	}

	mutex.RLock()
	booksCopy := make([]Book, len(books))
	copy(booksCopy, books)
	mutex.RUnlock()

	// Split books into chunks for parallel processing
	numWorkers := 4
	chunkSize := (len(booksCopy) + numWorkers - 1) / numWorkers

	// Channel to collect results
	resultsChan := make(chan []Book, numWorkers)

	var wg sync.WaitGroup

	// Create worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()

			var results []Book
			for j := start; j < end && j < len(booksCopy); j++ {
				book := booksCopy[j]
				if strings.Contains(strings.ToLower(book.Title), query) ||
					strings.Contains(strings.ToLower(book.Description), query) {
					results = append(results, book)
				}
			}

			resultsChan <- results
		}(i*chunkSize, (i+1)*chunkSize)
	}

	// Wait for all goroutines to finish and close the channel
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var searchResults []Book
	for results := range resultsChan {
		searchResults = append(searchResults, results...)
	}

	json.NewEncoder(w).Encode(searchResults)
}

func main() {
	// Load books from file
	err := loadBooks()
	if err != nil {
		log.Fatal("Failed to load books from file:", err)
	}

	// Create router
	router := mux.NewRouter()

	// Register routes
	router.HandleFunc("/books", getBooks).Methods("GET")
	router.HandleFunc("/books", createBook).Methods("POST")
	router.HandleFunc("/books/{id}", getBook).Methods("GET")
	router.HandleFunc("/books/{id}", updateBook).Methods("PUT")
	router.HandleFunc("/books/{id}", deleteBook).Methods("DELETE")
	router.HandleFunc("/books/search", searchBooks).Methods("GET")

	// Start server
	fmt.Println("Server is running on port 3000")
	log.Fatal(http.ListenAndServe(":3000", router))
}
