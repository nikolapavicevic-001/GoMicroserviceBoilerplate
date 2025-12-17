package com.microserviceboilerplate.domain.service;

import com.microserviceboilerplate.domain.model.Event;
import com.microserviceboilerplate.domain.port.input.EventSource;
import com.microserviceboilerplate.domain.port.output.EventProcessor;
import com.microserviceboilerplate.domain.port.output.EventSink;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;

import java.util.List;

/**
 * Service that orchestrates the event processing pipeline.
 * Composes source → processor → sinks.
 */
public class EventProcessingService {
    private final EventSource source;
    private final EventProcessor processor;
    private final List<EventSink> sinks;

    public EventProcessingService(
            EventSource source,
            EventProcessor processor,
            List<EventSink> sinks) {
        this.source = source;
        this.processor = processor;
        this.sinks = sinks;
    }

    /**
     * Builds and executes the event processing pipeline.
     *
     * @param env Flink execution environment
     * @return The processed event stream (for further composition if needed)
     */
    public DataStream<Event> buildPipeline(StreamExecutionEnvironment env) {
        // Create source stream
        DataStream<String> rawStream = source.createSource(env);

        // Process events
        DataStream<Event> eventStream = processor.process(rawStream);

        // Add all sinks
        for (EventSink sink : sinks) {
            sink.addSink(eventStream, env);
        }

        return eventStream;
    }
}

