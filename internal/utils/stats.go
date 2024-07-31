package utils

import (
    "context"
    "encoding/json"
    "time"
    "github.com/go-redis/redis/v8"
    "example.com/mod/internal/storage"
)

func GetCachedMessageStats(redisClient *redis.Client, db *storage.PostgresDB) (map[string]int, error) {
    ctx := context.Background()
    val, err := redisClient.Get(ctx, "message_stats").Result()
    if err == redis.Nil {
        stats, err := db.GetMessageStats()
        if err != nil {
            return nil, err
        }
        statsJSON, err := json.Marshal(stats)
        if err != nil {
            return nil, err
        }
        if err := redisClient.Set(ctx, "message_stats", statsJSON, time.Minute*5).Err(); err != nil {
            return nil, err
        }
        return stats, nil
    } else if err != nil {
        return nil, err
    }

    var stats map[string]int
    if err := json.Unmarshal([]byte(val), &stats); err != nil {
        return nil, err
    }
    return stats, nil
}

