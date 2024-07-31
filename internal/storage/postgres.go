package storage

import (
    "context"
    "github.com/jackc/pgx/v4/pgxpool"
    "github.com/google/uuid"
)

type PostgresDB struct {
    pool *pgxpool.Pool
}

type Message struct {
    ID    uuid.UUID `json:"id"`
    Key   string    `json:"key"`
    Value string    `json:"value"`
}

func NewPostgresDB(databaseURL string) (*PostgresDB, error) {
    config, err := pgxpool.ParseConfig(databaseURL)
    if err != nil {
        return nil, err
    }
    config.MaxConns = 50
    dbpool, err := pgxpool.ConnectConfig(context.Background(), config)
    if err != nil {
        return nil, err
    }
    return &PostgresDB{pool: dbpool}, nil
}

func (db *PostgresDB) SaveMessage(msg Message) error {
    _, err := db.pool.Exec(context.Background(), "INSERT INTO messages (id, key, value) VALUES ($1, $2, $3)", msg.ID, msg.Key, msg.Value)
    return err
}

func (db *PostgresDB) GetMessageStats() (map[string]int, error) {
    var count int
    err := db.pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM messages").Scan(&count)
    if err != nil {
        return nil, err
    }
    stats := map[string]int{"total_messages": count}
    return stats, nil
}

func (db *PostgresDB) Close() {
    db.pool.Close()
}

