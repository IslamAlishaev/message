# Stage 1: Build the Go application
FROM golang:1.22.5 AS builder 

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy
RUN go build -o /app/message-service ./cmd/main.go

# Stage 2: Setup Nginx and copy the built application
FROM golang:1.22.5 AS final 

WORKDIR /app

COPY --from=builder /app/message-service /app/message-service

EXPOSE 8080
EXPOSE 6060
CMD ["/app/message-service"]
