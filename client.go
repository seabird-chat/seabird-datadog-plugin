package client

import (
	"errors"
	"log"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/seabird-chat/seabird-go"
	"github.com/seabird-chat/seabird-go/pb"
)

// SeabirdClient is a basic client for seabird
type SeabirdClient struct {
	*seabird.Client

	datadogClient *statsd.Client
}

// NewSeabirdClient returns a new seabird client
func NewSeabirdClient(seabirdCoreURL, seabirdCoreToken, dogstatsdEndpoint string) (*SeabirdClient, error) {
	seabirdClient, err := seabird.NewClient(seabirdCoreURL, seabirdCoreToken)
	if err != nil {
		return nil, err
	}

	dogstatsd_client, err := statsd.New(dogstatsdEndpoint)
	if err != nil {
		log.Fatal(err)
	}

	return &SeabirdClient{
		Client:        seabirdClient,
		datadogClient: dogstatsd_client,
	}, nil
}

func (c *SeabirdClient) close() error {
	return c.Client.Close()
}

func (c *SeabirdClient) messageCallback(event *pb.MessageEvent) {
	log.Printf("Processing event: %s %s", event.Source, event.Text)
}

// Run runs
func (c *SeabirdClient) Run() error {
	events, err := c.StreamEvents(map[string]*pb.CommandMetadata{})
	if err != nil {
		return err
	}

	defer events.Close()
	for event := range events.C {
		switch v := event.GetInner().(type) {
		case *pb.Event_Message:
			c.messageCallback(v.Message)
		}
	}
	return errors.New("event stream closed")
}
