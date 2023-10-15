FROM alpine:3.17

COPY bin/swiss-knife /swiss-knife
COPY deploy/config.yaml /config.yaml

EXPOSE 8080

ENTRYPOINT ["/swiss-knife"]
