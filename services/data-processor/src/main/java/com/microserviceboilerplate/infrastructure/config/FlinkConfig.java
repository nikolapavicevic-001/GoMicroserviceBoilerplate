package com.microserviceboilerplate.infrastructure.config;

/**
 * Configuration for Flink data processor.
 * Loads values from environment variables with sensible defaults.
 */
public class FlinkConfig {
    private final String postgresUrl;
    private final String postgresUser;
    private final String postgresPassword;
    private final String s3Endpoint;
    private final String s3AccessKey;
    private final String s3SecretKey;
    private final String s3Bucket;
    private final String sourceType; // "socket", "kafka", "nats"
    private final String kafkaBootstrapServers;
    private final String kafkaTopic;
    private final String natsUrl;
    private final String natsSubject;
    private final long checkpointInterval;

    private FlinkConfig(Builder builder) {
        this.postgresUrl = builder.postgresUrl;
        this.postgresUser = builder.postgresUser;
        this.postgresPassword = builder.postgresPassword;
        this.s3Endpoint = builder.s3Endpoint;
        this.s3AccessKey = builder.s3AccessKey;
        this.s3SecretKey = builder.s3SecretKey;
        this.s3Bucket = builder.s3Bucket;
        this.sourceType = builder.sourceType;
        this.kafkaBootstrapServers = builder.kafkaBootstrapServers;
        this.kafkaTopic = builder.kafkaTopic;
        this.natsUrl = builder.natsUrl;
        this.natsSubject = builder.natsSubject;
        this.checkpointInterval = builder.checkpointInterval;
    }

    public static FlinkConfig fromEnvironment() {
        return new Builder()
                .postgresUrl(getEnv("POSTGRES_URL", "jdbc:postgresql://postgres:5432/postgres"))
                .postgresUser(getEnv("POSTGRES_USER", "postgres"))
                .postgresPassword(getEnv("POSTGRES_PASSWORD", "postgres"))
                .s3Endpoint(getEnv("S3_ENDPOINT", "http://minio:9000"))
                .s3AccessKey(getEnv("S3_ACCESS_KEY", "minioadmin"))
                .s3SecretKey(getEnv("S3_SECRET_KEY", "minioadmin"))
                .s3Bucket(getEnv("S3_BUCKET", "raw"))
                .sourceType(getEnv("SOURCE_TYPE", "socket"))
                .kafkaBootstrapServers(getEnv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"))
                .kafkaTopic(getEnv("KAFKA_TOPIC", "events"))
                .natsUrl(getEnv("NATS_URL", "nats://nats:4222"))
                .natsSubject(getEnv("NATS_SUBJECT", "events"))
                .checkpointInterval(Long.parseLong(getEnv("CHECKPOINT_INTERVAL_MS", "60000")))
                .build();
    }

    private static String getEnv(String key, String defaultValue) {
        String value = System.getenv(key);
        return value != null && !value.isEmpty() ? value : defaultValue;
    }

    // Getters
    public String getPostgresUrl() {
        return postgresUrl;
    }

    public String getPostgresUser() {
        return postgresUser;
    }

    public String getPostgresPassword() {
        return postgresPassword;
    }

    public String getS3Endpoint() {
        return s3Endpoint;
    }

    public String getS3AccessKey() {
        return s3AccessKey;
    }

    public String getS3SecretKey() {
        return s3SecretKey;
    }

    public String getS3Bucket() {
        return s3Bucket;
    }

    public String getSourceType() {
        return sourceType;
    }

    public String getKafkaBootstrapServers() {
        return kafkaBootstrapServers;
    }

    public String getKafkaTopic() {
        return kafkaTopic;
    }

    public String getNatsUrl() {
        return natsUrl;
    }

    public String getNatsSubject() {
        return natsSubject;
    }

    public long getCheckpointInterval() {
        return checkpointInterval;
    }

    public static class Builder {
        private String postgresUrl;
        private String postgresUser;
        private String postgresPassword;
        private String s3Endpoint;
        private String s3AccessKey;
        private String s3SecretKey;
        private String s3Bucket;
        private String sourceType;
        private String kafkaBootstrapServers;
        private String kafkaTopic;
        private String natsUrl;
        private String natsSubject;
        private long checkpointInterval;

        public Builder postgresUrl(String postgresUrl) {
            this.postgresUrl = postgresUrl;
            return this;
        }

        public Builder postgresUser(String postgresUser) {
            this.postgresUser = postgresUser;
            return this;
        }

        public Builder postgresPassword(String postgresPassword) {
            this.postgresPassword = postgresPassword;
            return this;
        }

        public Builder s3Endpoint(String s3Endpoint) {
            this.s3Endpoint = s3Endpoint;
            return this;
        }

        public Builder s3AccessKey(String s3AccessKey) {
            this.s3AccessKey = s3AccessKey;
            return this;
        }

        public Builder s3SecretKey(String s3SecretKey) {
            this.s3SecretKey = s3SecretKey;
            return this;
        }

        public Builder s3Bucket(String s3Bucket) {
            this.s3Bucket = s3Bucket;
            return this;
        }

        public Builder sourceType(String sourceType) {
            this.sourceType = sourceType;
            return this;
        }

        public Builder kafkaBootstrapServers(String kafkaBootstrapServers) {
            this.kafkaBootstrapServers = kafkaBootstrapServers;
            return this;
        }

        public Builder kafkaTopic(String kafkaTopic) {
            this.kafkaTopic = kafkaTopic;
            return this;
        }

        public Builder natsUrl(String natsUrl) {
            this.natsUrl = natsUrl;
            return this;
        }

        public Builder natsSubject(String natsSubject) {
            this.natsSubject = natsSubject;
            return this;
        }

        public Builder checkpointInterval(long checkpointInterval) {
            this.checkpointInterval = checkpointInterval;
            return this;
        }

        public FlinkConfig build() {
            return new FlinkConfig(this);
        }
    }
}

