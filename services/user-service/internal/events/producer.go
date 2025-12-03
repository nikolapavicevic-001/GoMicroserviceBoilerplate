package events

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/yourorg/boilerplate/shared/events"
	"github.com/yourorg/boilerplate/shared/kafka"
)

const (
	// UserEventsTopic is the Kafka topic for user events
	UserEventsTopic = "user.events"
)

// UserEventProducer handles publishing user-related events
type UserEventProducer struct {
	producer *kafka.Producer
	logger   *zerolog.Logger
}

// NewUserEventProducer creates a new user event producer
func NewUserEventProducer(producer *kafka.Producer, logger *zerolog.Logger) *UserEventProducer {
	return &UserEventProducer{
		producer: producer,
		logger:   logger,
	}
}

// PublishUserCreated publishes a user created event
func (p *UserEventProducer) PublishUserCreated(ctx context.Context, userID, email, name string) error {
	event, err := events.NewUserCreatedEvent(userID, email, name)
	if err != nil {
		p.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to create user created event")
		return err
	}

	return p.publishEvent(ctx, event)
}

// PublishUserUpdated publishes a user updated event
func (p *UserEventProducer) PublishUserUpdated(ctx context.Context, userID, email, name string) error {
	event, err := events.NewUserUpdatedEvent(userID, email, name)
	if err != nil {
		p.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to create user updated event")
		return err
	}

	return p.publishEvent(ctx, event)
}

// PublishUserDeleted publishes a user deleted event
func (p *UserEventProducer) PublishUserDeleted(ctx context.Context, userID string) error {
	event, err := events.NewUserDeletedEvent(userID)
	if err != nil {
		p.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to create user deleted event")
		return err
	}

	return p.publishEvent(ctx, event)
}

// publishEvent serializes and publishes an event to Kafka
func (p *UserEventProducer) publishEvent(ctx context.Context, event *events.Event) error {
	// Publish to Kafka using the key for partitioning
	err := p.producer.PublishWithKey(ctx, UserEventsTopic, event.AggregateID, event)
	if err != nil {
		p.logger.Error().Err(err).
			Str("event_type", event.Type).
			Str("event_id", event.ID).
			Str("topic", UserEventsTopic).
			Msg("Failed to publish event")
		return err
	}

	p.logger.Info().
		Str("event_type", event.Type).
		Str("event_id", event.ID).
		Str("aggregate_id", event.AggregateID).
		Str("topic", UserEventsTopic).
		Msg("Event published successfully")

	return nil
}

// Close closes the producer
func (p *UserEventProducer) Close() error {
	return p.producer.Close()
}

