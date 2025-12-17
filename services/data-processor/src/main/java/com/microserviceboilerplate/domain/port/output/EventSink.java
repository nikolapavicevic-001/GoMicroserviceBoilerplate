package com.microserviceboilerplate.domain.port.output;

import com.microserviceboilerplate.domain.model.Event;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;

/**
 * Output port for event sinks.
 * Implementations can write to PostgreSQL, S3, ClickHouse, etc.
 */
public interface EventSink {
    /**
     * Adds a sink to the data stream.
     *
     * @param stream DataStream of events to sink
     * @param env Flink execution environment
     */
    void addSink(DataStream<Event> stream, StreamExecutionEnvironment env);
}

