package pubsub

type Message struct {
	// LoggableID will be set to an opaque message identifier for
	// received messages, useful for debug logging. No assumptions should
	// be made about the content.
	LoggableID string

	// Body contains the content of the message.
	Body []byte
}
