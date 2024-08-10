package kafka

import (
    "context"
    "encoding/json"
    "log"
    "time"

    "example.com/mod/internal/storage"
    "github.com/segmentio/kafka-go"
    "github.com/segmentio/kafka-go/snappy"
)

type Producer struct {
    writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
    writer := kafka.NewWriter(kafka.WriterConfig{
        Brokers:           brokers,
        Topic:             topic,
        Balancer:          &kafka.LeastBytes{},
        BatchSize:         1,
        BatchTimeout:      10 * time.Millisecond,
        CompressionCodec:  snappy.NewCompressionCodec(),
        WriteTimeout:      10 * time.Second,
    })
    return &Producer{writer: writer}, nil
}
// Ping проверяет состояние подключения к Kafka
func (p *Producer) Ping() error {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()
    return p.writer.WriteMessages(ctx, kafka.Message{
        Key:   []byte("ping"),
        Value: []byte("ping"),
    })
}

func (p *Producer) SendMessage(ctx context.Context, message *storage.Message) error {
    key := []byte(message.ID.String())
    value, err := json.Marshal(message)
    if err != nil {
        log.Printf("Failed to marshal message: %v", err)
        return err
    }

    msg := kafka.Message{
        Key:   key,
        Value: value,
    }

    if err := p.writer.WriteMessages(ctx, msg); err != nil {
        log.Printf("Failed to send message to Kafka: %v", err)
        return err
    }

    log.Printf("Message successfully sent to Kafka: %s", message.ID)
    return nil
}

func (p *Producer) Close() error {
    if err := p.writer.Close(); err != nil {
        log.Printf("Failed to close Kafka writer: %v", err)
        return err
    }

    log.Println("Kafka writer successfully closed")
    return nil
}

