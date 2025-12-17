package com.microserviceboilerplate.infrastructure.job;

import com.microserviceboilerplate.domain.port.input.EventSource;
import com.microserviceboilerplate.adapters.input.KafkaSourceAdapter;
import com.microserviceboilerplate.adapters.input.NatsSource;
import com.microserviceboilerplate.adapters.input.SocketSource;
import com.microserviceboilerplate.adapters.output.PostgresSink;
import com.microserviceboilerplate.adapters.output.S3Sink;
import com.microserviceboilerplate.adapters.processor.JsonEventParser;
import com.microserviceboilerplate.domain.port.output.EventProcessor;
import com.microserviceboilerplate.domain.port.output.EventSink;
import com.microserviceboilerplate.domain.service.EventProcessingService;
import com.microserviceboilerplate.infrastructure.config.FlinkConfig;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.ArrayList;
import java.util.List;

/**
 * Main Flink job orchestration.
 * Sets up the execution environment and wires together all components.
 */
public class DataProcessorJob {
    private static final Logger logger = LoggerFactory.getLogger(DataProcessorJob.class);
    
    private final FlinkConfig config;
    private final StreamExecutionEnvironment env;

    public DataProcessorJob(FlinkConfig config) {
        this.config = config;
        this.env = StreamExecutionEnvironment.getExecutionEnvironment();
        setupEnvironment();
    }

    private void setupEnvironment() {
        // Enable checkpointing
        env.enableCheckpointing(config.getCheckpointInterval());
        
        logger.info("Flink environment configured with checkpoint interval: {} ms", 
                   config.getCheckpointInterval());
    }

    /**
     * Creates the event source based on configuration.
     */
    private EventSource createSource() {
        String sourceType = config.getSourceType().toLowerCase();
        
        switch (sourceType) {
            case "socket":
                logger.info("Using Socket source (host: localhost, port: 9999)");
                return new SocketSource("localhost", 9999);
            case "kafka":
                logger.info("Using Kafka source (servers: {}, topic: {})", 
                           config.getKafkaBootstrapServers(), config.getKafkaTopic());
                return new KafkaSourceAdapter(config.getKafkaBootstrapServers(), config.getKafkaTopic());
            case "nats":
                logger.info("Using NATS source (url: {}, subject: {})", 
                           config.getNatsUrl(), config.getNatsSubject());
                return new NatsSource(config.getNatsUrl(), config.getNatsSubject());
            default:
                logger.warn("Unknown source type: {}, defaulting to socket", sourceType);
                return new SocketSource("localhost", 9999);
        }
    }

    /**
     * Creates the event processor.
     */
    private EventProcessor createProcessor() {
        logger.info("Using JSON event parser");
        return new JsonEventParser();
    }

    /**
     * Creates all event sinks based on configuration.
     */
    private List<EventSink> createSinks() {
        List<EventSink> sinks = new ArrayList<>();

        // PostgreSQL sink
        logger.info("Adding PostgreSQL sink (url: {})", config.getPostgresUrl());
        sinks.add(new PostgresSink(
                config.getPostgresUrl(),
                config.getPostgresUser(),
                config.getPostgresPassword()
        ));

        // S3/MinIO sink
        logger.info("Adding S3/MinIO sink (endpoint: {}, bucket: {})", 
                   config.getS3Endpoint(), config.getS3Bucket());
        sinks.add(new S3Sink(
                config.getS3Endpoint(),
                config.getS3Bucket()
        ));

        return sinks;
    }

    /**
     * Builds and executes the Flink job.
     */
    public void execute() throws Exception {
        logger.info("Starting Data Processor Job");

        // Create components
        EventSource source = createSource();
        EventProcessor processor = createProcessor();
        List<EventSink> sinks = createSinks();

        // Create service and build pipeline
        EventProcessingService service = new EventProcessingService(source, processor, sinks);
        service.buildPipeline(env);

        // Execute job
        logger.info("Executing Flink job: Data Processor Job");
        env.execute("Data Processor Job");
        
        logger.info("Flink job completed");
    }

    public StreamExecutionEnvironment getEnvironment() {
        return env;
    }
}

