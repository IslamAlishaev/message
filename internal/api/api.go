package api

import (
    "example.com/mod/internal/handler"
    "example.com/mod/internal/kafka"
    "example.com/mod/internal/storage"
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRoutes настраивает маршруты для API
func SetupRoutes(r *gin.Engine, db *storage.PostgresDB, producer *kafka.Producer) {
    h := handler.NewHandler(db, producer)

    r.POST("/messages", h.HandleCreateMessage) // запросы на создание сообщения
    r.GET("/messages/:id", h.HandleGetMessage) // запросы на получение сообщения
    r.GET("/messages/stats", h.HandleGetStats) // запросы на получение статистики сообщений
    r.GET("/metrics", gin.WrapH(promhttp.Handler())) // для интеграции приложения с Prometheus (мониторинг)
    // Добавленние маршрутов для проверки состояния системы
    r.GET("/health/db", h.CheckDBHealth)       // проверка состояния БД
    r.GET("/health/kafka", h.CheckKafkaHealth) // проверка состояния Kafka
}

