FROM golang:1.22.5

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy
RUN go build -o /message-service ./cmd/main.go

EXPOSE 8080
CMD ["/message-service"]
