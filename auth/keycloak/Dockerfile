FROM quay.io/keycloak/keycloak:24.0

# Install curl and jq for realm import script
USER root
RUN microdnf install -y curl jq
USER keycloak

# Copy realm config
COPY realm-config.json /opt/keycloak/data/import/realm-config.json
COPY import-realm.sh /opt/keycloak/bin/import-realm.sh
RUN chmod +x /opt/keycloak/bin/import-realm.sh

# Start Keycloak and import realm
ENTRYPOINT ["/opt/keycloak/bin/kc.sh"]
CMD ["start-dev", "--import-realm"]

