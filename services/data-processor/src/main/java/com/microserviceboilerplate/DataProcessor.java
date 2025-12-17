package com.microserviceboilerplate;

import com.microserviceboilerplate.infrastructure.config.FlinkConfig;
import com.microserviceboilerplate.infrastructure.job.DataProcessorJob;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Main entry point for the Data Processor Flink job.
 * Loads configuration and executes the job.
 */
public class DataProcessor {
    private static final Logger logger = LoggerFactory.getLogger(DataProcessor.class);

    public static void main(String[] args) throws Exception {
        logger.info("Initializing Data Processor");

        // Load configuration from environment
        FlinkConfig config = FlinkConfig.fromEnvironment();
        logger.info("Configuration loaded: source={}, postgres={}, s3={}", 
                   config.getSourceType(), config.getPostgresUrl(), config.getS3Bucket());

        // Create and execute job
        DataProcessorJob job = new DataProcessorJob(config);
        job.execute();
    }
}
