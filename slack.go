package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

var apiUrl = "https://slack.com/api/users.profile.set"

func UpdateStatus(text string, config SlackConfig) error {
	m := map[string]string{"status_text": text, "status_emoji": config.Emoji}
	json, _ := json.Marshal(m)

	data := url.Values{}
	data.Set("token", config.AccessToken)
	data.Add("profile", string(json))

	client := &http.Client{}
	r, _ := http.NewRequest("POST", fmt.Sprintf("%s", apiUrl), bytes.NewBufferString(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	_, err := client.Do(r)

	if err != nil {
		return err
	}

	return nil
}
