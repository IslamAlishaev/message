package main

import (
    "log"
    "net/http"
    "os"

    "example.com/mod/internal/api"
    "example.com/mod/internal/kafka"
    "example.com/mod/internal/storage"
    "github.com/gin-gonic/gin"
    _ "net/http/pprof" // для pprof
)

func main() {
    // Получение переменных окружения
    databaseURL := os.Getenv("DATABASE_URL")
    kafkaBroker := os.Getenv("KAFKA_BROKER")

    if databaseURL == "" || kafkaBroker == "" {
        log.Fatalf("DATABASE_URL и KAFKA_BROKER должны быть установлены")
    }

    // Инициализация подключения к базе данных PostgreSQL
    db, err := storage.NewPostgresDB(databaseURL)
    if err != nil {
        log.Fatalf("Не удалось подключиться к базе данных: %v", err)
    }
    defer db.Close()

    // Инициализация Kafka Producer
    producer, err := kafka.NewProducer([]string{"kafka:9092"}, "messages")
    if err != nil {
        log.Fatalf("Не удалось создать продюсера Kafka: %v", err)
    }
    defer producer.Close()

    // Создание маршрутов и запуск сервера
    r := gin.Default()
    api.SetupRoutes(r, db, producer)

    // Запуск pprof сервера для профилирования
    go func() {
        log.Println(http.ListenAndServe("0.0.0.0:6060", nil)) // для pprof
    }()

    // Запуск основного HTTP сервера
    if err := r.Run(":8080"); err != nil {
        log.Fatalf("Не удалось запустить сервер: %v", err)
    }
}
