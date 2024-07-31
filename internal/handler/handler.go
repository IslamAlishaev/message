package handler

import (
    "context" // Добавляем импорт для context
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid" // Импорт для uuid (если нужен для newMessage.ID)

    "example.com/mod/internal/storage"
    "example.com/mod/internal/kafka" // Добавляем импорт для kafka
)

// Message структура для получения сообщений
type Message struct {
    Content string `json:"content"`
}

// ReceiveMessage обрабатывает получение сообщения
func ReceiveMessage(db *storage.PostgresDB, producer *kafka.Producer) gin.HandlerFunc {
    return func(c *gin.Context) {
        var msg Message
        if err := c.ShouldBindJSON(&msg); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        newMessage := storage.Message{
            ID:    uuid.New(), // Установка уникального идентификатора
            Key:   "default", // Установите здесь нужное значение для key
            Value: msg.Content,
        }

        // Сохранение сообщения в PostgreSQL
        if err := db.SaveMessage(newMessage); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        // Отправка сообщения в Kafka
        if err := producer.SendMessage(context.Background(), newMessage); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"status": "Message received", "id": newMessage.ID})
    }
}

// GetStats получает статистику сообщений
func GetStats(db *storage.PostgresDB) gin.HandlerFunc {
    return func(c *gin.Context) {
        stats, err := db.GetMessageStats()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, stats)
    }
}

