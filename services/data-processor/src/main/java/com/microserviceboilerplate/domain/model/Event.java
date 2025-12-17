package com.microserviceboilerplate.domain.model;

import java.time.Instant;
import java.util.Objects;

/**
 * Domain model representing an event in the system.
 */
public class Event {
    private String id;
    private String userId;
    private String eventType;
    private String payload;
    private Instant timestamp;

    private Event() {
        // Private constructor for builder
    }

    public String getId() {
        return id;
    }

    public String getUserId() {
        return userId;
    }

    public String getEventType() {
        return eventType;
    }

    public String getPayload() {
        return payload;
    }

    public Instant getTimestamp() {
        return timestamp;
    }

    /**
     * Validates the event has all required fields.
     */
    public void validate() {
        if (userId == null || userId.isEmpty()) {
            throw new IllegalArgumentException("Event userId cannot be null or empty");
        }
        if (eventType == null || eventType.isEmpty()) {
            throw new IllegalArgumentException("Event eventType cannot be null or empty");
        }
        if (payload == null || payload.isEmpty()) {
            throw new IllegalArgumentException("Event payload cannot be null or empty");
        }
        if (timestamp == null) {
            throw new IllegalArgumentException("Event timestamp cannot be null");
        }
    }

    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private Event event = new Event();

        public Builder id(String id) {
            event.id = id;
            return this;
        }

        public Builder userId(String userId) {
            event.userId = userId;
            return this;
        }

        public Builder eventType(String eventType) {
            event.eventType = eventType;
            return this;
        }

        public Builder payload(String payload) {
            event.payload = payload;
            return this;
        }

        public Builder timestamp(Instant timestamp) {
            event.timestamp = timestamp;
            return this;
        }

        public Event build() {
            event.validate();
            return event;
        }
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        Event event = (Event) o;
        return Objects.equals(id, event.id) &&
                Objects.equals(userId, event.userId) &&
                Objects.equals(eventType, event.eventType) &&
                Objects.equals(payload, event.payload) &&
                Objects.equals(timestamp, event.timestamp);
    }

    @Override
    public int hashCode() {
        return Objects.hash(id, userId, eventType, payload, timestamp);
    }

    @Override
    public String toString() {
        return "Event{" +
                "id='" + id + '\'' +
                ", userId='" + userId + '\'' +
                ", eventType='" + eventType + '\'' +
                ", timestamp=" + timestamp +
                '}';
    }
}

