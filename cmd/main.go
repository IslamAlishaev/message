package main

import (
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    _ "net/http/pprof" // для pprof

    "example.com/mod/internal/api"
    "example.com/mod/internal/kafka"
    "example.com/mod/internal/storage"
    ginSwagger "github.com/swaggo/gin-swagger" // Swagger
    swaggerFiles "github.com/swaggo/files"     // Swagger files
    _ "example.com/mod/docs"                   // импортируйте документацию
)

// MaxRequestBodySizeMiddleware ограничивает размер тела запроса
func MaxRequestBodySizeMiddleware(limit int64) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
        c.Next()
    }
}

// @title Message Service API
// @version 1.0
// @description API для работы с сообщениями.
// @host localhost:8080
// @BasePath /
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
    producer, err := kafka.NewProducer([]string{kafkaBroker}, "messages")
    if err != nil {
        log.Fatalf("Не удалось создать продюсера Kafka: %v", err)
    }
    defer producer.Close()

    // Создание маршрутов и запуск сервера
    r := gin.Default()
    r.Use(MaxRequestBodySizeMiddleware(10 << 20)) // Установите лимит на 10 МБ

    api.SetupRoutes(r, db, producer)

    // Добавление маршрутов для Swagger
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // Запуск pprof сервера для профилирования
    go func() {
        log.Println("Starting pprof server on :6060")
        if err := http.ListenAndServe("0.0.0.0:6060", nil); err != nil {
            log.Fatalf("pprof server failed: %v", err)
        }
    }()

    // Запуск основного HTTP сервера
    go func() {
        if err := r.Run(":8080"); err != nil {
            log.Fatalf("Не удалось запустить сервер: %v", err)
        }
    }()

    // Запуск планировщика удаления старых сообщений
    go func() {
        ticker := time.NewTicker(24 * time.Hour) // Удаление каждую ночь
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                if err := db.DeleteOldMessages(); err != nil {
                    log.Printf("Не удалось удалить старые сообщения: %v", err)
                }
            }
        }
    }()

    // Обработка сигнала завершения
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Завершение работы сервера...")
}

