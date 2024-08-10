package handler

import (
    "context"
    "log"
    "net/http"

    "example.com/mod/internal/kafka"
    "example.com/mod/internal/storage"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

// Handler представляет обработчик API
type Handler struct {
    db       *storage.PostgresDB
    producer *kafka.Producer
}

// NewHandler создает новый обработчик API
func NewHandler(db *storage.PostgresDB, producer *kafka.Producer) *Handler {
    return &Handler{
        db:       db,
        producer: producer,
    }
}

// CheckDBHealth проверяет состояние подключения к базе данных
// @Summary Проверяет состояние подключения к базе данных
// @Description Проверяет состояние подключения к базе данных
// @Tags health
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /health/db [get]
func (h *Handler) CheckDBHealth(c *gin.Context) {
    if err := h.db.Ping(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "Database connection failed", "error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "Database connection is healthy"})
}

// CheckKafkaHealth проверяет состояние подключения к Kafka
// @Summary Проверяет состояние подключения к Kafka
// @Description Проверяет состояние подключения к Kafka
// @Tags health
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /health/kafka [get]
func (h *Handler) CheckKafkaHealth(c *gin.Context) {
    if err := h.producer.Ping(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "Kafka connection failed", "error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "Kafka connection is healthy"})
}

// HandleCreateMessage обрабатывает запросы на создание сообщения
// @Summary Создает новое сообщение
// @Description Создает новое сообщение и отправляет его в Kafka
// @Tags messages
// @Accept json
// @Produce json
// @Param message body storage.Message true "Message"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /messages [post]
func (h *Handler) HandleCreateMessage(c *gin.Context) {
    var msg storage.Message
    if err := c.ShouldBindJSON(&msg); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }
    msg.ID = uuid.New()
    msg.Processed = false // Изначально сообщение не обработано

    if err := h.db.SaveMessage(msg); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
        return
    }

    go func() {
        if err := h.producer.SendMessage(context.Background(), &msg); err != nil {
            log.Printf("Failed to send message to Kafka: %v", err)
            return
        }

        if err := h.db.MarkMessageAsProcessed(msg.ID); err != nil {
            log.Printf("Failed to mark message as processed: %v", err)
            return
        }
    }()

    c.JSON(http.StatusOK, gin.H{"id": msg.ID.String(), "status": "Message received"})
}

// HandleGetMessage обрабатывает запросы на получение сообщения
// @Summary Получает сообщение по ID
// @Description Возвращает сообщение по ID
// @Tags messages
// @Produce json
// @Param id path string true "Message ID"
// @Success 200 {object} Message
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /messages/{id} [get]
func (h *Handler) HandleGetMessage(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
        return
    }

    msg, err := h.db.GetMessageByID(id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get message"})
        return
    }

    c.JSON(http.StatusOK, msg)
}

// HandleGetStats обрабатывает запросы на получение статистики сообщений
// @Summary Получает статистику сообщений
// @Description Возвращает статистику сообщений
// @Tags messages
// @Produce json
// @Success 200 {object} storage.MessageStats
// @Failure 500 {object} map[string]interface{}
// @Router /messages/stats [get]
func (h *Handler) HandleGetStats(c *gin.Context) {
    stats, err := h.db.GetMessageStats()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
        return
    }
    c.JSON(http.StatusOK, stats)
}

