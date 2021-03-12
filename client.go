package client

import (
	"errors"
	"fmt"
	"log"
	"strings"

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

func countWords(s string) int64 {
	words := strings.Fields(s)
	var sum int64
	sum = 0
	for range words {
		sum += 1
	}
	return sum
}

func isAttack(s string) bool {
	violentWords := []string{
		"slaps",
		"beats",
		"smacks",
		"hits",
	}
	for _, attack := range violentWords {
		if strings.Contains(s, attack) {
			return true
		}
	}
	return false
}

func parseFight(s string) (attacker, victim string) {
	return "", ""
}

func isFoul(s string) bool {
	foulWords := []string{
		"ass",
		"fuck",
		"bitch",
		"shit",
		"scheisse",
		"scheiße",
		"kacke",
		"arsch",
		"ficker",
		"ficken",
		"schlampe",
	}

	for _, swear := range foulWords {
		if strings.Contains(s, swear) {
			return true
		}
	}
	return false
}

func isShout(s string) bool {
	if s[len(s):] == "!" {
		return true
	}
	return false
}

func isQuestion(s string) bool {
	if strings.Contains(s, "?") {
		return true
	}
	return false
}

func isHappy(s string) bool {
	happiness := []string{
		":)",
		"(:",
	}
	for _, happy := range happiness {
		if strings.Contains(s, happy) {
			return true
		}
	}
	return false
}

func isSad(s string) bool {
	sadness := []string{
		":(",
		"):",
	}
	for _, sad := range sadness {
		if strings.Contains(s, sad) {
			return true
		}
	}
	return false
}

func (c *SeabirdClient) close() error {
	return c.Client.Close()
}

func (c *SeabirdClient) messageCallback(event *pb.MessageEvent) {
	log.Printf("Processing event: %s %s", event.Source, event.Text)
	channelTag := fmt.Sprintf("channel:%s", event.Source.ChannelId)
	displayNameTag := fmt.Sprintf("display_name:%s", event.Source.User.DisplayName)
	userTag := fmt.Sprintf("user:%s", event.Source.User.Id)
	tags := []string{
		channelTag,
		displayNameTag,
		userTag,
	}
	// TODO: This assumes max of 1 message per second per user.
	c.datadogClient.Count("seabird.message", 1, tags, 1)
	c.datadogClient.Count("seabird.message.words", countWords(event.Text), tags, 1)
	c.datadogClient.Count("seabird.message.characters", int64(len(event.Text)), tags, 1)
	if isAttack(event.Text) {
		c.datadogClient.Count("seabird.message.attack", 1, tags, 1)
	}
	if isFoul(event.Text) {
		c.datadogClient.Count("seabird.message.swear", 1, tags, 1)
	}
	if isShout(event.Text) {
		c.datadogClient.Count("seabird.message.shout", 1, tags, 1)
	}
	if isQuestion(event.Text) {
		c.datadogClient.Count("seabird.message.question", 1, tags, 1)
	}
	if isHappy(event.Text) {
		c.datadogClient.Count("seabird.message.happy", 1, tags, 1)
	}
	if isSad(event.Text) {
		c.datadogClient.Count("seabird.message.sad", 1, tags, 1)
	}

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
			go c.messageCallback(v.Message)
		}
	}
	return errors.New("event stream closed")
}
