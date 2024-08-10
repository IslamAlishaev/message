package storage

import (
    "context"
    "errors"
    "time"
    "github.com/jackc/pgx/v4"
    "github.com/jackc/pgx/v4/pgxpool"
    "github.com/google/uuid"
)

type PostgresDB struct {
    pool *pgxpool.Pool
}

type Message struct {
    ID        uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
    Key       string    `json:"key" example:"example_key"`
    Value     string    `json:"value" example:"example_value"`
    Processed bool      `json:"processed" example:"false"`
    CreatedAt time.Time `json:"created_at" example:"2024-08-09T12:34:56Z"`
}

// Экспортируемый тип MessageStats
type MessageStats struct {
    TotalMessages     int `json:"total_messages" example:"100"`
    ProcessedMessages int `json:"processed_messages" example:"80"`
}

// NewPostgresDB создает новое подключение к базе данных
func NewPostgresDB(databaseURL string) (*PostgresDB, error) {
    config, err := pgxpool.ParseConfig(databaseURL)
    if (err != nil) {
        return nil, err
    }
    config.MaxConns = 50
    dbpool, err := pgxpool.ConnectConfig(context.Background(), config)
    if err != nil {
        return nil, err
    }
    return &PostgresDB{pool: dbpool}, nil
}

// Ping проверяет состояние подключения к базе данных
func (db *PostgresDB) Ping() error {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()
    return db.pool.Ping(ctx)
}

// SaveMessage сохраняет сообщение в базе данных
func (db *PostgresDB) SaveMessage(msg Message) error {
    _, err := db.pool.Exec(context.Background(), "INSERT INTO messages (id, key, value, processed, created_at) VALUES ($1, $2, $3, $4, NOW())", msg.ID, msg.Key, msg.Value, msg.Processed)
    return err
}

// GetMessageStats возвращает статистику сообщений
func (db *PostgresDB) GetMessageStats() (MessageStats, error) {
    var totalCount int
    var processedCount int

    err := db.pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM messages").Scan(&totalCount)
    if err != nil {
        return MessageStats{}, err
    }

    err = db.pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM messages WHERE processed = TRUE").Scan(&processedCount)
    if err != nil {
        return MessageStats{}, err
    }

    stats := MessageStats{
        TotalMessages:     totalCount,
        ProcessedMessages: processedCount,
    }
    return stats, nil
}

// MarkMessageAsProcessed помечает сообщение как обработанное
func (db *PostgresDB) MarkMessageAsProcessed(id uuid.UUID) error {
    _, err := db.pool.Exec(context.Background(), "UPDATE messages SET processed = TRUE WHERE id = $1", id)
    return err
}

// DeleteOldMessages удаляет старые сообщения из базы данных
func (db *PostgresDB) DeleteOldMessages() error {
    _, err := db.pool.Exec(context.Background(), "DELETE FROM messages WHERE created_at < NOW() - INTERVAL '7 days'")
    return err
}

// Close закрывает соединение с базой данных
func (db *PostgresDB) Close() {
    db.pool.Close()
}

// GetMessageByID возвращает сообщение по его ID
func (db *PostgresDB) GetMessageByID(id uuid.UUID) (*Message, error) {
    var msg Message
    err := db.pool.QueryRow(context.Background(), "SELECT id, key, value, processed, created_at FROM messages WHERE id = $1", id).Scan(&msg.ID, &msg.Key, &msg.Value, &msg.Processed, &msg.CreatedAt)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, nil // Сообщение не найдено
        }
        return nil, err // Другая ошибка
    }
    return &msg, nil
}

