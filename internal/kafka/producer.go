package kafka

import (
    "context"
    "example.com/mod/internal/storage"
    "github.com/segmentio/kafka-go"
)

type Producer struct {
    writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
    writer := kafka.NewWriter(kafka.WriterConfig{
        Brokers:  brokers,
        Topic:    topic,
        Balancer: &kafka.LeastBytes{},
    })
    return &Producer{writer: writer}, nil
}

func (p *Producer) SendMessage(ctx context.Context, message storage.Message) error {
    msg := kafka.Message{
        Key:   []byte(message.ID.String()),
        Value: []byte(message.Value),
    }
    return p.writer.WriteMessages(ctx, msg)
}

func (p *Producer) Close() error {
    return p.writer.Close()
}

