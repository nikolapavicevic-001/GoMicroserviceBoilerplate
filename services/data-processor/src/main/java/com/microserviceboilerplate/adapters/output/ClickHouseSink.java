package com.microserviceboilerplate.adapters.output;

import com.microserviceboilerplate.domain.model.Event;
import com.microserviceboilerplate.domain.port.output.EventSink;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;

/**
 * ClickHouse sink adapter.
 * TODO: Implement when ClickHouse JDBC connector is added.
 */
public class ClickHouseSink implements EventSink {
    private final String jdbcUrl;
    private final String username;
    private final String password;

    public ClickHouseSink(String jdbcUrl, String username, String password) {
        this.jdbcUrl = jdbcUrl;
        this.username = username;
        this.password = password;
    }

    @Override
    public void addSink(DataStream<Event> stream, StreamExecutionEnvironment env) {
        // TODO: Implement ClickHouse sink when ClickHouse JDBC dependency is added
        // Similar to PostgresSink but with ClickHouse-specific SQL and connection options
        
        throw new UnsupportedOperationException("ClickHouse sink not yet implemented. " +
                "Add ClickHouse JDBC dependency and implement ClickHouseSink.");
    }
}

