package main

import (
	"fmt"

	"github.com/mitsu9/slack-spotify-status/internal"
)

func main() {
	config, err := internal.GetConfigFromToml()
	//config, err := GetConfigFromGCP()

	if err != nil {
		fmt.Println(err)
		return
	}

	title, artist, err := internal.GetNowListening(&config.Spotify)

	text := "Now playing: " + title + " by " + artist
	if err != nil || title == "" || artist == "" {
		text = "Not Playing"
	}

	if err := internal.UpdateStatus(text, config.Slack); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Updated status with text: " + text)

	if err := internal.SaveConfigToToml(config); err != nil {
		//if err := SaveConfigToGCP(config); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Saved config")
}
