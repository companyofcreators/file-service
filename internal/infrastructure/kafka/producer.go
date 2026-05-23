package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"

	appfile "github.com/companyofcreators/file-service/internal/application/file"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	return &Producer{writer: writer}
}

func (p *Producer) PublishFileUploaded(ctx context.Context, event appfile.FileUploadedEvent) error {
	return p.publish(ctx, "file.uploaded", event)
}

func (p *Producer) PublishFileDeleted(ctx context.Context, event appfile.FileDeletedEvent) error {
	return p.publish(ctx, "file.deleted", event)
}

func (p *Producer) publish(ctx context.Context, topic string, event interface{}) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event for topic %s: %w", topic, err)
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   nil,
		Value: payload,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to publish message to topic %s: %w", topic, err)
	}

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
