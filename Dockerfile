# Use Go base image
FROM golang:1.20

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o book_api

EXPOSE 8080
CMD ["./book_api"]
