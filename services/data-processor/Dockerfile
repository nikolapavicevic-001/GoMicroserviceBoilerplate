# Stage 1: Build JAR with Maven
FROM maven:3.9-eclipse-temurin-11 AS builder

WORKDIR /app

# Copy pom.xml
COPY services/data-processor/pom.xml .

# Copy source code and build (Maven will download dependencies during build)
COPY services/data-processor/src ./src
RUN mvn clean package -DskipTests

# Stage 2: Serve JAR via HTTP
FROM eclipse-temurin:11-jre-alpine

WORKDIR /app

# Install Python for HTTP server (http.server is built-in, no pip needed)
RUN apk add --no-cache python3

# Copy JAR from builder
COPY --from=builder /app/target/data-processor-*.jar /app/data-processor.jar

# Expose port for serving JAR
EXPOSE 8080

# Serve JAR file via HTTP (http.server is a built-in Python module)
CMD ["python3", "-m", "http.server", "8080", "--directory", "/app"]

