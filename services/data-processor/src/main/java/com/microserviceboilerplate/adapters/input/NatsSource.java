package com.microserviceboilerplate.adapters.input;

import com.microserviceboilerplate.domain.port.input.EventSource;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;

/**
 * NATS source adapter for reading events from NATS subjects.
 * TODO: Implement when NATS connector is available or create custom connector.
 */
public class NatsSource implements EventSource {
    private final String natsUrl;
    private final String subject;

    public NatsSource(String natsUrl, String subject) {
        this.natsUrl = natsUrl;
        this.subject = subject;
    }

    @Override
    public DataStream<String> createSource(StreamExecutionEnvironment env) {
        // TODO: Implement NATS source when NATS connector is available
        // This would require creating a custom Flink source function
        // or using a NATS connector library
        
        throw new UnsupportedOperationException("NATS source not yet implemented. " +
                "NATS connector needs to be created or a third-party library integrated.");
    }
}

