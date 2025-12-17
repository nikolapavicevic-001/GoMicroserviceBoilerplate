package com.microserviceboilerplate.domain.port.output;

import com.microserviceboilerplate.domain.model.Event;
import org.apache.flink.streaming.api.datastream.DataStream;

/**
 * Output port for event processing/transformation.
 * Implementations can parse JSON, route events, transform data, etc.
 */
public interface EventProcessor {
    /**
     * Processes raw event strings into Event objects.
     *
     * @param rawStream DataStream of raw event strings
     * @return DataStream of parsed Event objects
     */
    DataStream<Event> process(DataStream<String> rawStream);
}

