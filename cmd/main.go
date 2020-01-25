package main

import (
	"fmt"

	"github.com/mitsu9/slack-spotify-status"
)

func main() {
	config, err := GetConfigFromToml()
	//config, err := GetConfigFromGCP()

	if err != nil {
		fmt.Println(err)
		return
	}

	title, artist, err := GetNowListening(&config.Spotify)

	text := "Now playing: " + title + " by " + artist
	if err != nil || title == "" || artist == "" {
		text = "Not Playing"
	}

	if err := UpdateStatus(text, config.Slack); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Updated status with text: " + text)

	if err := SaveConfigToToml(config); err != nil {
	//if err := SaveConfigToGCP(config); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Saved config")
}
