package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Config represents MongoDB configuration
type Config struct {
	URI            string
	Database       string
	Timeout        time.Duration
	MaxPoolSize    uint64
	MinPoolSize    uint64
	MaxConnIdleTime time.Duration
}

// Connect establishes a connection to MongoDB
func Connect(ctx context.Context, cfg *Config) (*mongo.Client, error) {
	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetMaxConnIdleTime(cfg.MaxConnIdleTime).
		SetServerSelectionTimeout(cfg.Timeout)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return client, nil
}

// NewConfig creates a new MongoDB configuration with defaults
func NewConfig(uri, database string) *Config {
	return &Config{
		URI:             uri,
		Database:        database,
		Timeout:         10 * time.Second,
		MaxPoolSize:     100,
		MinPoolSize:     10,
		MaxConnIdleTime: 30 * time.Second,
	}
}

// Disconnect closes the MongoDB connection
func Disconnect(ctx context.Context, client *mongo.Client) error {
	if client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return client.Disconnect(ctx)
}

// Ping checks if MongoDB is reachable
func Ping(ctx context.Context, client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.Ping(ctx, readpref.Primary())
}
