package containers

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBContainer represents a MongoDB test container
type MongoDBContainer struct {
	Container testcontainers.Container
	URI       string
	Host      string
	Port      string
}

// MongoDBContainerConfig holds configuration for MongoDB container
type MongoDBContainerConfig struct {
	// Image is the Docker image to use (default: mongo:7)
	Image string

	// Database is the database name to use
	Database string

	// Username is the MongoDB root username (optional)
	Username string

	// Password is the MongoDB root password (optional)
	Password string
}

// DefaultMongoDBConfig returns default configuration for MongoDB container
func DefaultMongoDBConfig() MongoDBContainerConfig {
	return MongoDBContainerConfig{
		Image:    "mongo:7",
		Database: "test_db",
	}
}

// StartMongoDBContainer starts a new MongoDB container for testing
func StartMongoDBContainer(ctx context.Context, config MongoDBContainerConfig) (*MongoDBContainer, error) {
	if config.Image == "" {
		config.Image = "mongo:7"
	}

	env := map[string]string{}
	if config.Username != "" && config.Password != "" {
		env["MONGO_INITDB_ROOT_USERNAME"] = config.Username
		env["MONGO_INITDB_ROOT_PASSWORD"] = config.Password
	}

	req := testcontainers.ContainerRequest{
		Image:        config.Image,
		ExposedPorts: []string{"27017/tcp"},
		Env:          env,
		WaitingFor: wait.ForAll(
			wait.ForLog("Waiting for connections").WithStartupTimeout(60*time.Second),
			wait.ForListeningPort("27017/tcp"),
		),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start MongoDB container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "27017")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	uri := fmt.Sprintf("mongodb://%s:%s", host, mappedPort.Port())
	if config.Username != "" && config.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%s",
			config.Username, config.Password, host, mappedPort.Port())
	}

	return &MongoDBContainer{
		Container: container,
		URI:       uri,
		Host:      host,
		Port:      mappedPort.Port(),
	}, nil
}

// Terminate stops and removes the container
func (m *MongoDBContainer) Terminate(ctx context.Context) error {
	if m.Container != nil {
		return m.Container.Terminate(ctx)
	}
	return nil
}

// GetClient returns a MongoDB client connected to the container
func (m *MongoDBContainer) GetClient(ctx context.Context) (*mongo.Client, error) {
	clientOpts := options.Client().ApplyURI(m.URI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return client, nil
}

// GetDatabase returns a MongoDB database from the container
func (m *MongoDBContainer) GetDatabase(ctx context.Context, dbName string) (*mongo.Database, error) {
	client, err := m.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.Database(dbName), nil
}

// Clean drops all collections in the specified database
func (m *MongoDBContainer) Clean(ctx context.Context, dbName string) error {
	db, err := m.GetDatabase(ctx, dbName)
	if err != nil {
		return err
	}

	collections, err := db.ListCollectionNames(ctx, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	for _, collection := range collections {
		if err := db.Collection(collection).Drop(ctx); err != nil {
			return fmt.Errorf("failed to drop collection %s: %w", collection, err)
		}
	}

	return nil
}

