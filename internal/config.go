package internal

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"

	cloudkms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/storage"
	"github.com/BurntSushi/toml"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

type Config struct {
	Slack   SlackConfig
	Spotify SpotifyConfig
}

type SlackConfig struct {
	Emoji       string
	AccessToken string `toml:"access_token"`
}

type SpotifyConfig struct {
	ClientId          string `toml:"client_id"`
	ClientSecret      string `toml:"client_secret"`
	AuthorizationCode string `toml:"authorization_code"`
	AccessToken       string `toml:"access_token"`
	RefreshToken      string `toml:"refresh_token"`
}

var configFile = "config.toml"

func GetConfigFromToml() (Config, error) {
	var config Config
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return config, err
	}
	return config, nil
}

func SaveConfigToToml(config Config) error {
	var buffer bytes.Buffer
	encoder := toml.NewEncoder(&buffer)
	if err := encoder.Encode(config); err != nil {
		return err
	}

	file, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer file.Close()

	file.Write(([]byte)(buffer.String()))
	return nil
}

func read(client *storage.Client, bucket, object string) ([]byte, error) {
	ctx := context.Background()
	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func write(client *storage.Client, bucket, object string, text []byte, keyName string) error {
	ctx := context.Background()

	obj := client.Bucket(bucket).Object(object)

	wc := obj.NewWriter(ctx)
	wc.KMSKeyName = keyName
	if _, err := wc.Write(text); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}

func getKeyName() string {
	projectID := os.Getenv("PROJECT_ID")
	location := os.Getenv("LOCATION")
	keyRingName := os.Getenv("KEY_RING_NAME")
	keyName := os.Getenv("KEY_NAME")

	return fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", projectID, location, keyRingName, keyName)
}

func decrypt(name string, text []byte) ([]byte, error) {
	ctx := context.Background()
	client, err := cloudkms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, err
	}

	req := &kmspb.DecryptRequest{
		Name:       name,
		Ciphertext: text,
	}
	resp, err := client.Decrypt(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Plaintext, nil
}

func encrypt(name string, text []byte) ([]byte, error) {
	ctx := context.Background()
	client, err := cloudkms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, err
	}

	req := &kmspb.EncryptRequest{
		Name:      name,
		Plaintext: text,
	}
	resp, err := client.Encrypt(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Ciphertext, nil
}

func GetConfigFromGCP() (Config, error) {
	var config Config

	fullKeyName := getKeyName()
	bucket := os.Getenv("BUCKET")
	filename := os.Getenv("FILENAME")

	// cloud strageから設定ファイルを取得
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return config, err
	}

	data, err := read(client, bucket, filename)
	if err != nil {
		return config, err
	}

	// kmsで復号化
	decrypted_data, err := decrypt(fullKeyName, data)
	if err != nil {
		return config, err
	}

	// Structに変換
	if _, err := toml.Decode(string(decrypted_data), &config); err != nil {
		return config, err
	}

	return config, nil
}

func SaveConfigToGCP(config Config) error {
	fullKeyName := getKeyName()
	bucket := os.Getenv("BUCKET")
	filename := os.Getenv("FILENAME")

	// config to buffer
	var buffer bytes.Buffer
	encoder := toml.NewEncoder(&buffer)
	if err := encoder.Encode(config); err != nil {
		return err
	}

	// encrypt
	data := ([]byte)(buffer.String())
	encrypted_data, err := encrypt(fullKeyName, data)
	if err != nil {
		return err
	}

	// cloud strageに設定ファイルを保存
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	if err := write(client, bucket, filename, encrypted_data, fullKeyName); err != nil {
		return err
	}

	return nil
}
