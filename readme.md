# Book API

A simple REST API for performing CRUD operations on Book entities, written in Go.

## Features

- CRUD operations on Book entities
- Keyword search functionality on book titles and descriptions
- Optimized search using Go's concurrency features
- Persistence using JSON files
- Docker containerization

## Requirements

- Go 1.19 or higher
- Docker (optional, for containerization)

## Installation

1. Clone the repository
2. Install dependencies:
   ```
   go mod download
   ```

## Running the Application

### Local Development

```bash
go run main.go
```

The server will start on port 8080.

### Using Docker

Build the Docker image:
```bash
docker build -t book-api .
```

Run the container:
```bash
docker run -p 8080:8080 book-api
```

## API Endpoints

- `GET /books`: Return a list of all books
- `POST /books`: Create a new book
- `GET /books/{id}`: Return a single book by ID
- `PUT /books/{id}`: Update a single book by ID
- `DELETE /books/{id}`: Delete a single book by ID
- `GET /books/search?q=<keyword>`: Search books by keyword in title and description

## Running Tests

```bash
go test -v
```

## Example Usage

### Creating a Book

```bash
curl -X POST http://localhost:8080/books \
  -H "Content-Type: application/json" \
  -d '{
    "authorId": "e0d91f68-a183-477d-8aa4-1f44ccc78a70",
    "publisherId": "2f7b19e9-b268-4440-a15b-bed8177ed607",
    "title": "The Great Gatsby",
    "publicationDate": "1925-04-10",
    "isbn": "9780743273565",
    "pages": 180,
    "genre": "Novel",
    "description": "Set in the 1920s, this classic novel explores themes of wealth, love, and the American Dream.",
    "price": 15.99,
    "quantity": 5
  }'
```

### Searching Books

```bash
curl -X GET "http://localhost:8080/books/search?q=gatsby"
```