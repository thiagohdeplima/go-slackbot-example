package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

var requiredVars = []string{
	"SLACK_AUTH_TOKEN",
	"SLACK_APP_TOKEN",
}

func main() {
	if err := validateEnvironment(); err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	client := getSlackClient(ctx)
	socket := getSlackSocket(client)

	runSocket(ctx, client, socket)
}

func validateEnvironment() error {
	var empty = []string{}

	for _, env := range requiredVars {
		if os.Getenv(env) == "" {
			empty = append(empty, env)
		}
	}

	if len(empty) > 0 {
		return errors.New(
			fmt.Sprintf(
				"empty environment variables: %s",
				strings.Join(empty[:], ", "),
			),
		)
	}

	return nil
}

func getSlackClient(ctx context.Context) *slack.Client {
	slackAuthToken := os.Getenv("SLACK_AUTH_TOKEN")
	slackAppToken := os.Getenv("SLACK_APP_TOKEN")

	return slack.New(
		slackAuthToken,
		slack.OptionDebug(true),
		slack.OptionAppLevelToken(slackAppToken),
	)
}

func getSlackSocket(client *slack.Client) *socketmode.Client {
	enableDebug := socketmode.OptionDebug(true)

	logLevel := socketmode.OptionLog(
		log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags),
	)

	return socketmode.New(client, enableDebug, logLevel)
}

func runSocket(ctx context.Context, client *slack.Client, socket *socketmode.Client) {
	go getAndHandleEvents(ctx, client, socket)

	socket.Run()
}

func getAndHandleEvents(ctx context.Context, client *slack.Client, socket *socketmode.Client) {
	for {
		select {
		case event := <-socket.Events:
			filterEvents(event)

		case <-ctx.Done():
			log.Println("shutdown the application...")
			return
		}
	}
}

func filterEvents(event socketmode.Event) {
	log.Println("----------------------> ", event.Type)

	switch event.Type {
	case socketmode.EventTypeEventsAPI:
		log.Printf("handling the event %+v", event)

	case socketmode.EventTypeSlashCommand:
		log.Printf("received a command %+v", event)

	default:
		log.Printf("unhandleable event: %+v", event)
		log.Printf("event data ---> %+v", event.Data)
	}
}
