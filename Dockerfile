FROM alpine
ADD webhook-to-pubsub /webhook-to-pubsub
ENTRYPOINT ["/webhook-to-pubsub"]
