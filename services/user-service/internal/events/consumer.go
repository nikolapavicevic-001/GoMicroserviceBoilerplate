package events

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/rs/zerolog"
	"github.com/yourorg/boilerplate/shared/events"
	"github.com/yourorg/boilerplate/shared/kafka"
)

// UserEventHandler defines the interface for handling user events
type UserEventHandler interface {
	HandleUserCreated(ctx context.Context, event *events.Event, data *events.UserCreatedData) error
	HandleUserUpdated(ctx context.Context, event *events.Event, data *events.UserUpdatedData) error
	HandleUserDeleted(ctx context.Context, event *events.Event, data *events.UserDeletedData) error
}

// UserEventConsumer handles consuming user-related events
type UserEventConsumer struct {
	consumer *kafka.Consumer
	handler  UserEventHandler
	logger   *zerolog.Logger
	wg       sync.WaitGroup
}

// NewUserEventConsumer creates a new user event consumer
func NewUserEventConsumer(consumer *kafka.Consumer, handler UserEventHandler, logger *zerolog.Logger) *UserEventConsumer {
	return &UserEventConsumer{
		consumer: consumer,
		handler:  handler,
		logger:   logger,
	}
}

// Start begins consuming events from the user events topic
func (c *UserEventConsumer) Start(ctx context.Context) error {
	c.logger.Info().Str("topic", UserEventsTopic).Msg("Starting user event consumer")

	// Register the message handler
	c.consumer.RegisterHandler(UserEventsTopic, c.handleMessage)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		if err := c.consumer.Start(ctx, []string{UserEventsTopic}); err != nil {
			c.logger.Error().Err(err).Msg("Error consuming user events")
		}
	}()

	return nil
}

// handleMessage processes a single message from Kafka
func (c *UserEventConsumer) handleMessage(ctx context.Context, message []byte) error {
	// Deserialize the event
	var event events.Event
	if err := kafka.DecodeMessage(message, &event); err != nil {
		c.logger.Error().Err(err).Msg("Failed to deserialize event")
		return err
	}

	c.logger.Info().
		Str("event_id", event.ID).
		Str("event_type", event.Type).
		Str("aggregate_id", event.AggregateID).
		Msg("Processing event")

	// Route to appropriate handler based on event type
	switch event.Type {
	case events.UserCreated:
		return c.handleUserCreated(ctx, &event)
	case events.UserUpdated:
		return c.handleUserUpdated(ctx, &event)
	case events.UserDeleted:
		return c.handleUserDeleted(ctx, &event)
	default:
		c.logger.Warn().Str("event_type", event.Type).Msg("Unknown event type")
		return nil
	}
}

// Stop stops the consumer and waits for it to finish
func (c *UserEventConsumer) Stop() error {
	c.logger.Info().Msg("Stopping user event consumer")
	c.wg.Wait()
	return c.consumer.Close()
}

func (c *UserEventConsumer) handleUserCreated(ctx context.Context, event *events.Event) error {
	var data events.UserCreatedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}
	return c.handler.HandleUserCreated(ctx, event, &data)
}

func (c *UserEventConsumer) handleUserUpdated(ctx context.Context, event *events.Event) error {
	var data events.UserUpdatedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}
	return c.handler.HandleUserUpdated(ctx, event, &data)
}

func (c *UserEventConsumer) handleUserDeleted(ctx context.Context, event *events.Event) error {
	var data events.UserDeletedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}
	return c.handler.HandleUserDeleted(ctx, event, &data)
}

// NoOpEventHandler is a default handler that does nothing (for services that only produce events)
type NoOpEventHandler struct {
	logger *zerolog.Logger
}

// NewNoOpEventHandler creates a new no-op event handler
func NewNoOpEventHandler(logger *zerolog.Logger) *NoOpEventHandler {
	return &NoOpEventHandler{logger: logger}
}

// HandleUserCreated logs the event but takes no action
func (h *NoOpEventHandler) HandleUserCreated(ctx context.Context, event *events.Event, data *events.UserCreatedData) error {
	h.logger.Info().
		Str("event_id", event.ID).
		Str("user_id", data.UserID).
		Msg("User created event received (no action)")
	return nil
}

// HandleUserUpdated logs the event but takes no action
func (h *NoOpEventHandler) HandleUserUpdated(ctx context.Context, event *events.Event, data *events.UserUpdatedData) error {
	h.logger.Info().
		Str("event_id", event.ID).
		Str("user_id", data.UserID).
		Msg("User updated event received (no action)")
	return nil
}

// HandleUserDeleted logs the event but takes no action
func (h *NoOpEventHandler) HandleUserDeleted(ctx context.Context, event *events.Event, data *events.UserDeletedData) error {
	h.logger.Info().
		Str("event_id", event.ID).
		Str("user_id", data.UserID).
		Msg("User deleted event received (no action)")
	return nil
}

