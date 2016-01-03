# webhook-to-pubsub

Handles incoming requests and publishes their data up to Google
PubSub. The payload must be encrypted before sent to ensure the
data isn't spoofed.

Use AES128 encryption with a shared key and a random IV. When sent,
the IV should be prepended to the payload. The result should be
url-safe base64 (- instead of + and _ instead of /) and sent as
the `p` over the query string.

Config should be loaded via environment variables:

- HTTP_PORT - the port that the service is exposed on
- KEY - the encryption key
- PUBSUB_CREDS - json containing ProjectID and TopicName

## Example

```
GET /fulfillment/vanity-string?p=[iv+payload] HTTP/1.1
content-length: 0


```

## License

MIT
