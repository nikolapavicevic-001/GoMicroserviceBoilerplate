package com.microserviceboilerplate.adapters.processor;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.microserviceboilerplate.domain.model.Event;
import com.microserviceboilerplate.domain.port.output.EventProcessor;
import org.apache.flink.api.common.functions.MapFunction;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.time.Instant;
import java.util.UUID;

/**
 * JSON event parser adapter.
 * Parses raw JSON strings into Event domain objects.
 */
public class JsonEventParser implements EventProcessor {
    private static final Logger logger = LoggerFactory.getLogger(JsonEventParser.class);
    private final ObjectMapper objectMapper;

    public JsonEventParser() {
        this.objectMapper = new ObjectMapper();
    }

    @Override
    public DataStream<Event> process(DataStream<String> rawStream) {
        return rawStream.map(new MapFunction<String, Event>() {
            @Override
            public Event map(String jsonString) throws Exception {
                try {
                    JsonNode jsonNode = objectMapper.readTree(jsonString);
                    
                    // Extract fields from JSON
                    String userId = getStringField(jsonNode, "user_id");
                    String eventType = getStringField(jsonNode, "event_type");
                    String payload = jsonString; // Store full JSON as payload
                    
                    // Extract timestamp or use current time
                    Instant timestamp = extractTimestamp(jsonNode);
                    
                    // Generate ID if not present
                    String id = getStringField(jsonNode, "id");
                    if (id == null || id.isEmpty()) {
                        id = UUID.randomUUID().toString();
                    }

                    return Event.builder()
                            .id(id)
                            .userId(userId)
                            .eventType(eventType)
                            .payload(payload)
                            .timestamp(timestamp)
                            .build();
                } catch (Exception e) {
                    logger.error("Failed to parse event JSON: {}", jsonString, e);
                    // Return a default event with error information
                    // In production, you might want to send to a dead letter queue
                    throw new RuntimeException("Failed to parse event: " + e.getMessage(), e);
                }
            }

            private String getStringField(JsonNode node, String fieldName) {
                JsonNode field = node.get(fieldName);
                return field != null && !field.isNull() ? field.asText() : null;
            }

            private Instant extractTimestamp(JsonNode node) {
                JsonNode timestampNode = node.get("timestamp");
                if (timestampNode != null && !timestampNode.isNull()) {
                    try {
                        // Try to parse as Unix timestamp (seconds or milliseconds)
                        long timestamp = timestampNode.asLong();
                        if (timestamp > 1e10) {
                            // Milliseconds
                            return Instant.ofEpochMilli(timestamp);
                        } else {
                            // Seconds
                            return Instant.ofEpochSecond(timestamp);
                        }
                    } catch (Exception e) {
                        // Try parsing as ISO-8601 string
                        try {
                            return Instant.parse(timestampNode.asText());
                        } catch (Exception e2) {
                            logger.warn("Could not parse timestamp, using current time", e2);
                        }
                    }
                }
                return Instant.now();
            }
        });
    }
}

