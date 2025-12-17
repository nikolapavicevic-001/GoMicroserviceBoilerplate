package com.microserviceboilerplate.domain.port.input;

import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;

/**
 * Input port for event sources.
 * Implementations can read from NATS, Kafka, Socket, etc.
 */
public interface EventSource {
    /**
     * Creates a data stream from the source.
     *
     * @param env Flink execution environment
     * @return DataStream of raw event strings
     */
    DataStream<String> createSource(StreamExecutionEnvironment env);
}

