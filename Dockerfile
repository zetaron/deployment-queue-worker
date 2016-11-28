FROM alpine:3.4

CMD ["/usr/bin/secret-wrapper", "/usr/bin/deployment-queue-worker"]
LABEL org.label-schema.schema-version="1.0" \
      org.label-schema.url="https://github.com/zetaron/deployment-queue-worker" \
      org.label-schema.vcs-url="https://github.com/zetaron/deployment-queue-worker" \
      org.label-schema.name="deployment-queue-worker" \
      org.label-schema.docker.cmd="docker run -d -p 80 -v ${SECRETS_VOLUME_NAME:-deployment-queue-worker-secrets}:/var/secrets:ro --name deployment-queue-worker zetaron/deployment-queue-worker:1.0.0"

COPY secret-wrapper /usr/bin/secret-wrapper
COPY deployment-queue-worker /usr/bin/deployment-queue-worker
