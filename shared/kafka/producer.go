package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog"
)

// Producer represents a Kafka producer
type Producer struct {
	producer sarama.SyncProducer
	logger   *zerolog.Logger
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string, logger *zerolog.Logger) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Compression = sarama.CompressionSnappy

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &Producer{
		producer: producer,
		logger:   logger,
	}, nil
}

// Publish publishes a message to a Kafka topic
func (p *Producer) Publish(ctx context.Context, topic string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		p.logger.Error().
			Err(err).
			Str("topic", topic).
			Msg("failed to publish message to Kafka")
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.logger.Debug().
		Str("topic", topic).
		Int32("partition", partition).
		Int64("offset", offset).
		Msg("message published to Kafka")

	return nil
}

// PublishWithKey publishes a message with a key to a Kafka topic
func (p *Producer) PublishWithKey(ctx context.Context, topic, key string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		p.logger.Error().
			Err(err).
			Str("topic", topic).
			Str("key", key).
			Msg("failed to publish message to Kafka")
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.logger.Debug().
		Str("topic", topic).
		Str("key", key).
		Int32("partition", partition).
		Int64("offset", offset).
		Msg("message published to Kafka")

	return nil
}

// Close closes the Kafka producer
func (p *Producer) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}
