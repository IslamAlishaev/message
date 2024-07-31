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
    r.POST("/messages", handler.ReceiveMessage(db, producer))
    
    r.GET("/stats", handler.GetStats(db))
    
    r.GET("/metrics", gin.WrapH(promhttp.Handler())) // Для Prometheus
}

