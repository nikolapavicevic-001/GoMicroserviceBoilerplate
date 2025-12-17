package com.microserviceboilerplate.adapters.output;

import com.microserviceboilerplate.domain.model.Event;
import com.microserviceboilerplate.domain.port.output.EventSink;
import org.apache.flink.api.common.serialization.SimpleStringEncoder;
import org.apache.flink.core.fs.Path;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;
import org.apache.flink.streaming.api.functions.sink.filesystem.StreamingFileSink;
import org.apache.flink.streaming.api.functions.sink.filesystem.rollingpolicies.DefaultRollingPolicy;

import java.time.Duration;

/**
 * S3/MinIO sink adapter.
 * Writes events to S3-compatible storage (MinIO).
 */
public class S3Sink implements EventSink {
    private final String s3Endpoint;
    private final String bucket;
    private final String basePath;
    private final Duration rolloverInterval;
    private final Duration inactivityInterval;
    private final long maxPartSize;

    public S3Sink(String s3Endpoint, String bucket) {
        this(s3Endpoint, bucket, "events/", 
             Duration.ofMinutes(15), 
             Duration.ofMinutes(5), 
             128 * 1024 * 1024); // 128 MB
    }

    public S3Sink(String s3Endpoint, String bucket, String basePath,
                  Duration rolloverInterval, Duration inactivityInterval, long maxPartSize) {
        this.s3Endpoint = s3Endpoint;
        this.bucket = bucket;
        this.basePath = basePath;
        this.rolloverInterval = rolloverInterval;
        this.inactivityInterval = inactivityInterval;
        this.maxPartSize = maxPartSize;
    }

    @Override
    public void addSink(DataStream<Event> stream, StreamExecutionEnvironment env) {
        // Convert Event to JSON string for storage
        DataStream<String> jsonStream = stream.map(event -> {
            // Simple JSON serialization - in production use proper JSON library
            return String.format(
                "{\"id\":\"%s\",\"user_id\":\"%s\",\"event_type\":\"%s\",\"payload\":%s,\"timestamp\":%d}",
                event.getId(),
                event.getUserId(),
                event.getEventType(),
                event.getPayload(),
                event.getTimestamp().toEpochMilli()
            );
        });

        StreamingFileSink<String> s3Sink = StreamingFileSink
                .forRowFormat(
                    new Path("s3://" + bucket + "/" + basePath),
                    new SimpleStringEncoder<String>("UTF-8")
                )
                .withRollingPolicy(
                    DefaultRollingPolicy.builder()
                            .withRolloverInterval(rolloverInterval)
                            .withInactivityInterval(inactivityInterval)
                            .withMaxPartSize(maxPartSize)
                            .build()
                )
                .build();

        jsonStream.addSink(s3Sink).name("S3/MinIO Sink");
    }
}

