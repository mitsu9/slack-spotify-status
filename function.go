package function

import (
	"context"
	"fmt"
	"log"

	"github.com/mitsu9/slack-spotify-status/internal"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

func Subscription(ctx context.Context, m PubSubMessage) error {
	config, err := internal.GetConfigFromGCP()

	if err != nil {
		log.Println(err)
		return err
	}

	title, artist, err := internal.GetNowListening(&config.Spotify)

	text := "Now playing: " + title + " by " + artist
	if err != nil || title == "" || artist == "" {
		text = "Not Playing"
	}

	if err := internal.UpdateStatus(text, config.Slack); err != nil {
		log.Println(err)
		return err
	}
	fmt.Println("Updated status with text: " + text)

	if err := internal.SaveConfigToGCP(config); err != nil {
		log.Println(err)
		return err
	}
	fmt.Println("Saved config")

	return nil
}
