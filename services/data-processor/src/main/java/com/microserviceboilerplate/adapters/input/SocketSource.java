package com.microserviceboilerplate.adapters.input;

import com.microserviceboilerplate.domain.port.input.EventSource;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;

/**
 * Socket source adapter for reading events from a socket.
 * Useful for testing and development.
 */
public class SocketSource implements EventSource {
    private final String host;
    private final int port;

    public SocketSource(String host, int port) {
        this.host = host;
        this.port = port;
    }

    @Override
    public DataStream<String> createSource(StreamExecutionEnvironment env) {
        return env.socketTextStream(host, port)
                .filter(line -> line != null && !line.isEmpty());
    }
}

