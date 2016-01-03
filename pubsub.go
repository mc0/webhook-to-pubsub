package main

import (
	"cloud.google.com/go/pubsub"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"log"
)

// TopicMessage is used for passing a message to the pubsub topic
type TopicMessage struct {
	Msg       string
	ReplyChan chan *PublishReply
}

// PublishReply is a local struct to wrap pubsub.PublishResult
type PublishReply struct {
	*pubsub.PublishResult
}

// PubSubSpec is what we use for grabbing the google pubsub
// projectID and topicName for publishing our messages.
type PubSubSpec struct {
	ProjectID string
	TopicName string
}

func startPubSubChannel(creds *PubSubSpec) chan TopicMessage {
	c := make(chan TopicMessage, 5000)

	// Runs to consume the channel asynchronously to allow the request to
	// be non-blocking. This may not be desirable for all webhooks,
	// if it is not then this channel can be made blocking by
	// removing the size above.
	go (func() {
		ctx := context.Background()

		projectID := creds.ProjectID

		client, err := pubsub.NewClient(ctx, projectID)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}

		// Only fail if the topic is nil since "already exists" is fine
		topic, err := client.CreateTopic(ctx, creds.TopicName)
		if topic == nil {
			log.Fatalf("Failed to create topic: %v", err)
		}

		for job := range c {
			publishResult := topic.Publish(ctx, &pubsub.Message{Data: []byte(job.Msg)})
			if nil != err {
				log.Printf("failed to send message: %s", err)
				metrics.PubSubFailures.With(prometheus.Labels{"topic": creds.TopicName}).Inc()
			} else {
				metrics.PubSubMessages.With(prometheus.Labels{"topic": creds.TopicName}).Inc()
			}
			if job.ReplyChan != nil {
				job.ReplyChan <- &PublishReply{publishResult}
			}
		}
	})()

	return c
}
