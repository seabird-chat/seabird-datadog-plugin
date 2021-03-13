package client

import (
	"errors"
	"fmt"
	"log"
	"net/url"
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

	dogstatsdClient, err := statsd.New(dogstatsdEndpoint)
	if err != nil {
		log.Fatal(err)
	}

	return &SeabirdClient{
		Client:        seabirdClient,
		datadogClient: dogstatsdClient,
	}, nil
}

func countWords(s string) int64 {
	words := strings.Fields(s)
	var sum int64
	sum = 0
	for range words {
		sum++
	}
	return sum
}

func isFoul(s string) bool {
	for _, swear := range Swears() {
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

	// channel and user appear to be query escaped so, undo that
	cleanedChannel, _ := url.QueryUnescape(event.Source.ChannelId)
	channelTag := fmt.Sprintf("channel:%s", cleanedChannel)
	cleanedUser, _ := url.QueryUnescape(event.Source.User.Id)
	userTag := fmt.Sprintf("user:%s", cleanedUser)

	displayNameTag := fmt.Sprintf("display_name:%s", event.Source.User.DisplayName)
	tags := []string{
		channelTag,
		displayNameTag,
		userTag,
	}
	// TODO: This assumes max of 1 message per second per user.
	c.datadogClient.Count("seabird.message", 1, tags, 1)
	c.datadogClient.Count("seabird.message.words", countWords(event.Text), tags, 1)
	c.datadogClient.Count("seabird.message.characters", int64(len(event.Text)), tags, 1)
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

func parseFight(user, text string) (attacker, victim string) {
	attacker = user
	words := strings.Fields(text)
	victim = words[1]
	return attacker, victim
}

func isConsume(s string) bool {
	consumeWords := []string{
		"sips",
		"chugs",
		"gulps",
		"eats",
		"swallows",
		"devours",
	}
	for _, consume := range consumeWords {
		if strings.Contains(s, consume) {
			return true
		}
	}
	return false
}

func parseConsume(text string) (food string) {
	words := strings.Fields(text)
	food = strings.Join(words[:1], " ")
	return food
}

func (c *SeabirdClient) actionCallback(action *pb.ActionEvent) {
	log.Printf("Processing event: %s %s", action.Source, action.Text)

	// channel and user appear to be query escaped so, undo that
	cleanedChannel, _ := url.QueryUnescape(action.Source.ChannelId)
	channelTag := fmt.Sprintf("channel:%s", cleanedChannel)
	cleanedUser, _ := url.QueryUnescape(action.Source.User.Id)
	userTag := fmt.Sprintf("user:%s", cleanedUser)

	displayNameTag := fmt.Sprintf("display_name:%s", action.Source.User.DisplayName)
	tags := []string{
		channelTag,
		displayNameTag,
		userTag,
	}

	c.datadogClient.Count("seabird.message", 1, tags, 1)
	c.datadogClient.Count("seabird.message.words", countWords(action.Text), tags, 1)
	c.datadogClient.Count("seabird.message.characters", int64(len(action.Text)), tags, 1)
	c.datadogClient.Count("seabird.message.action", 1, tags, 1)
	if isAttack(action.Text) {
		attacker, victim := parseFight(action.Source.User.DisplayName, action.Text)
		attackerTag := fmt.Sprintf("attacker:%s", attacker)
		victimTag := fmt.Sprintf("victim:%s", victim)
		tags = append(tags, attackerTag, victimTag)
		c.datadogClient.Count("seabird.message.attack", 1, tags, 1)
	}
	if isConsume(action.Text) {
		c.datadogClient.Count("seabird.message.consume", 1, tags, 1)
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
		case *pb.Event_Action:
			go c.actionCallback(v.Action)
		}
	}
	return errors.New("event stream closed")
}
