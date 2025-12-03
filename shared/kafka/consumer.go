package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog"
)

// MessageHandler is a function that handles a Kafka message
type MessageHandler func(ctx context.Context, message []byte) error

// Consumer represents a Kafka consumer
type Consumer struct {
	consumer sarama.ConsumerGroup
	logger   *zerolog.Logger
	handlers map[string]MessageHandler
	mu       sync.RWMutex
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers []string, groupID string, logger *zerolog.Logger) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	return &Consumer{
		consumer: consumer,
		logger:   logger,
		handlers: make(map[string]MessageHandler),
	}, nil
}

// RegisterHandler registers a handler for a specific topic
func (c *Consumer) RegisterHandler(topic string, handler MessageHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[topic] = handler
}

// Start starts consuming messages from the specified topics
func (c *Consumer) Start(ctx context.Context, topics []string) error {
	handler := &consumerGroupHandler{
		consumer: c,
		logger:   c.logger,
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-c.consumer.Errors():
				if err != nil {
					c.logger.Error().Err(err).Msg("Kafka consumer error")
				}
			}
		}
	}()

	for {
		if err := c.consumer.Consume(ctx, topics, handler); err != nil {
			c.logger.Error().Err(err).Msg("failed to consume from Kafka")
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

// Close closes the Kafka consumer
func (c *Consumer) Close() error {
	if c.consumer != nil {
		return c.consumer.Close()
	}
	return nil
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	consumer *Consumer
	logger   *zerolog.Logger
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		h.logger.Debug().
			Str("topic", message.Topic).
			Int32("partition", message.Partition).
			Int64("offset", message.Offset).
			Msg("received message from Kafka")

		// Get handler for this topic
		h.consumer.mu.RLock()
		handler, ok := h.consumer.handlers[message.Topic]
		h.consumer.mu.RUnlock()

		if !ok {
			h.logger.Warn().
				Str("topic", message.Topic).
				Msg("no handler registered for topic")
			session.MarkMessage(message, "")
			continue
		}

		// Process message
		if err := handler(session.Context(), message.Value); err != nil {
			h.logger.Error().
				Err(err).
				Str("topic", message.Topic).
				Msg("failed to process message")
			// Don't mark message if processing failed
			continue
		}

		// Mark message as processed
		session.MarkMessage(message, "")
	}

	return nil
}

// DecodeMessage decodes a JSON message
func DecodeMessage(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
