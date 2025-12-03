package containers

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

// KafkaContainer represents a Kafka test container
type KafkaContainer struct {
	Container *kafka.KafkaContainer
	Brokers   []string
}

// KafkaContainerConfig holds configuration for Kafka container
type KafkaContainerConfig struct {
	// Image is the Docker image to use
	Image string

	// ClusterID is the Kafka cluster ID (optional)
	ClusterID string
}

// DefaultKafkaConfig returns default configuration for Kafka container
func DefaultKafkaConfig() KafkaContainerConfig {
	return KafkaContainerConfig{
		Image:     "confluentinc/confluent-local:7.5.0",
		ClusterID: "test-cluster",
	}
}

// StartKafkaContainer starts a new Kafka container for testing
func StartKafkaContainer(ctx context.Context, config KafkaContainerConfig) (*KafkaContainer, error) {
	if config.Image == "" {
		config.Image = "confluentinc/confluent-local:7.5.0"
	}

	kafkaContainer, err := kafka.Run(ctx,
		config.Image,
		kafka.WithClusterID(config.ClusterID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start Kafka container: %w", err)
	}

	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		kafkaContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to get Kafka brokers: %w", err)
	}

	return &KafkaContainer{
		Container: kafkaContainer,
		Brokers:   brokers,
	}, nil
}

// Terminate stops and removes the container
func (k *KafkaContainer) Terminate(ctx context.Context) error {
	if k.Container != nil {
		return k.Container.Terminate(ctx)
	}
	return nil
}

// GetAdmin returns a Kafka admin client
func (k *KafkaContainer) GetAdmin() (sarama.ClusterAdmin, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V3_0_0_0

	admin, err := sarama.NewClusterAdmin(k.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka admin: %w", err)
	}

	return admin, nil
}

// CreateTopic creates a new topic in Kafka
func (k *KafkaContainer) CreateTopic(ctx context.Context, topic string, partitions int32, replicationFactor int16) error {
	admin, err := k.GetAdmin()
	if err != nil {
		return err
	}
	defer admin.Close()

	topicDetail := &sarama.TopicDetail{
		NumPartitions:     partitions,
		ReplicationFactor: replicationFactor,
	}

	err = admin.CreateTopic(topic, topicDetail, false)
	if err != nil {
		// Ignore "already exists" error
		if kafkaErr, ok := err.(*sarama.TopicError); ok && kafkaErr.Err == sarama.ErrTopicAlreadyExists {
			return nil
		}
		return fmt.Errorf("failed to create topic %s: %w", topic, err)
	}

	return nil
}

// DeleteTopic deletes a topic from Kafka
func (k *KafkaContainer) DeleteTopic(topic string) error {
	admin, err := k.GetAdmin()
	if err != nil {
		return err
	}
	defer admin.Close()

	err = admin.DeleteTopic(topic)
	if err != nil {
		return fmt.Errorf("failed to delete topic %s: %w", topic, err)
	}

	return nil
}

// GetProducer returns a Kafka sync producer
func (k *KafkaContainer) GetProducer() (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V3_0_0_0
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	producer, err := sarama.NewSyncProducer(k.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return producer, nil
}

// GetConsumer returns a Kafka consumer group
func (k *KafkaContainer) GetConsumer(groupID string) (sarama.ConsumerGroup, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V3_0_0_0
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumerGroup(k.Brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	return consumer, nil
}

// ProduceMessage sends a message to a topic
func (k *KafkaContainer) ProduceMessage(topic string, key, value []byte) error {
	producer, err := k.GetProducer()
	if err != nil {
		return err
	}
	defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	_, _, err = producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	return nil
}

// WaitForTopicReady waits for a topic to be ready for consumption
func (k *KafkaContainer) WaitForTopicReady(ctx context.Context, topic string, timeout time.Duration) error {
	admin, err := k.GetAdmin()
	if err != nil {
		return err
	}
	defer admin.Close()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		topics, err := admin.ListTopics()
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if _, ok := topics[topic]; ok {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for topic %s", topic)
}

