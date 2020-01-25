package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type AccessTokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// generated by https://mholt.github.io/json-to-go/
type AutoGenerated struct {
	Timestamp            int64   `json:"timestamp"`
	Context              Context `json:"context"`
	ProgressMs           int     `json:"progress_ms"`
	Item                 Item    `json:"item"`
	CurrentlyPlayingType string  `json:"currently_playing_type"`
	Actions              Actions `json:"actions"`
	IsPlaying            bool    `json:"is_playing"`
}
type ExternalUrls struct {
	Spotify string `json:"spotify"`
}
type Context struct {
	ExternalUrls ExternalUrls `json:"external_urls"`
	Href         string       `json:"href"`
	Type         string       `json:"type"`
	URI          string       `json:"uri"`
}
type Artists struct {
	ExternalUrls ExternalUrls `json:"external_urls"`
	Href         string       `json:"href"`
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	URI          string       `json:"uri"`
}
type Images struct {
	Height int    `json:"height"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
}
type Album struct {
	AlbumType            string       `json:"album_type"`
	Artists              []Artists    `json:"artists"`
	AvailableMarkets     []string     `json:"available_markets"`
	ExternalUrls         ExternalUrls `json:"external_urls"`
	Href                 string       `json:"href"`
	ID                   string       `json:"id"`
	Images               []Images     `json:"images"`
	Name                 string       `json:"name"`
	ReleaseDate          string       `json:"release_date"`
	ReleaseDatePrecision string       `json:"release_date_precision"`
	TotalTracks          int          `json:"total_tracks"`
	Type                 string       `json:"type"`
	URI                  string       `json:"uri"`
}
type ExternalIds struct {
	Isrc string `json:"isrc"`
}
type Item struct {
	Album            Album        `json:"album"`
	Artists          []Artists    `json:"artists"`
	AvailableMarkets []string     `json:"available_markets"`
	DiscNumber       int          `json:"disc_number"`
	DurationMs       int          `json:"duration_ms"`
	Explicit         bool         `json:"explicit"`
	ExternalIds      ExternalIds  `json:"external_ids"`
	ExternalUrls     ExternalUrls `json:"external_urls"`
	Href             string       `json:"href"`
	ID               string       `json:"id"`
	IsLocal          bool         `json:"is_local"`
	Name             string       `json:"name"`
	Popularity       int          `json:"popularity"`
	PreviewURL       string       `json:"preview_url"`
	TrackNumber      int          `json:"track_number"`
	Type             string       `json:"type"`
	URI              string       `json:"uri"`
}
type Disallows struct {
	Resuming bool `json:"resuming"`
}
type Actions struct {
	Disallows Disallows `json:"disallows"`
}

var redirectUrl = "https://example.com/callback"
var spotifyApiUrl = "https://api.spotify.com/v1/me/player/currently-playing"

func GetAccessToken(config *SpotifyConfig) error {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Add("code", config.AuthorizationCode)
	data.Add("redirect_uri", redirectUrl)

	client := &http.Client{}
	r, _ := http.NewRequest("POST", "https://accounts.spotify.com/api/token", bytes.NewBufferString(data.Encode()))
	r.SetBasicAuth(config.ClientId, config.ClientSecret)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(r)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	jsonBytes := ([]byte)(byteArray)
	jsonData := new(AccessTokenResp)

	if err := json.Unmarshal(jsonBytes, jsonData); err != nil {
		fmt.Println("JSON Unmarshal error:", err)
		return err
	}

	config.AccessToken = jsonData.AccessToken
	config.RefreshToken = jsonData.RefreshToken

	return nil
}

func RefreshAccessToken(config *SpotifyConfig) (string, string, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Add("refresh_token", config.RefreshToken)

	client := &http.Client{}
	r, _ := http.NewRequest("POST", "https://accounts.spotify.com/api/token", bytes.NewBufferString(data.Encode()))
	r.SetBasicAuth(config.ClientId, config.ClientSecret)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(r)

	if err != nil {
		return "", "", err
	}

	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	jsonBytes := ([]byte)(byteArray)
	jsonData := new(AccessTokenResp)

	if err := json.Unmarshal(jsonBytes, jsonData); err != nil {
		fmt.Println("JSON Unmarshal error:", err)
		return "", "", err
	}

	return jsonData.AccessToken, jsonData.RefreshToken, nil
}

func RefreshAndRetry(config *SpotifyConfig) (string, string, error) {
	newAccessToken, newRefreshToken, err := RefreshAccessToken(config)

	if err != nil {
		return "", "", err
	}

	fmt.Println("Refreshed token")

	config.AccessToken = newAccessToken
	if newRefreshToken != "" {
		config.RefreshToken = newRefreshToken
		fmt.Println("Updated newRefreshToken")
	}

	client := &http.Client{}
	r, _ := http.NewRequest("GET", spotifyApiUrl, nil)
	var bearer = "Bearer " + newAccessToken
	r.Header.Add("Authorization", bearer)

	resp, err := client.Do(r)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return ParseResp(resp)
	case 204:
		return "", "", nil
	default:
		return "", "", nil
	}
}

func ParseResp(resp *http.Response) (string, string, error) {
	byteArray, _ := ioutil.ReadAll(resp.Body)
	jsonBytes := ([]byte)(byteArray)
	data := new(AutoGenerated)

	if err := json.Unmarshal(jsonBytes, data); err != nil {
		fmt.Println("JSON Unmarshal error:", err)
		return "", "", err
	}

	title := data.Item.Name
	artist_name := data.Item.Artists[0].Name

	return title, artist_name, nil
}

func GetNowListening(config *SpotifyConfig) (string, string, error) {
	if config.AccessToken == "" {
		if err := GetAccessToken(config); err != nil {
			return "", "", err
		}
	}

	client := &http.Client{}
	r, _ := http.NewRequest("GET", spotifyApiUrl, nil)
	var bearer = "Bearer " + config.AccessToken
	r.Header.Add("Authorization", bearer)

	resp, err := client.Do(r)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return ParseResp(resp) // playing
	case 204:
		return "", "", nil // not playing
	case 401:
		return RefreshAndRetry(config) // unauthorized
	default:
		return "", "", nil
	}
}
