# slack-spotify-status

Spotifyで再生している音楽をSlackのStatusに設定します.

GCPで定期実行するためにGCP用のコードも入っています.

## Setup (ローカル・GCP共通)
config.toml.sampleを参考に必要情報を入力します.
Spotifyのaccess_token, refresh_tokenは初回実行時に取得することもできます.

## ローカルで動かす
```
$ go run cmd/main.go
```

## GCPで動かす
### Setup
Cloud StorageにCloud KMSで暗号化したconfig.tomlを保管します.

```
# 鍵の作成
$ gcloud kms keyrings create KEY_RING_NAME --location LOCATION
$ gcloud kms keys create KEY_NAME \
  --location LOCATION \
  --keyring KEY_RING_NAME \
  --purpose encryption
# 暗号化
$ gcloud kms encrypt \
  --location LOCATION \
  --keyring KEY_RING_NAME \
  --key KEY_NAME \
  --plaintext-file config.toml \
  --ciphertext-file config.toml.enc
# バケットへ保存
$ gsutil mb gs://BUCKET/
$ gsutil cp config.toml.enc gs://BUCKET/
```

### Deploy
以下のようなenv.yamlを作成します.

```
BUCKET: <bucket>
FILENAME: config.toml.enc
PROJECT_ID: <project_id>
LOCATION: <location>
KEY_RING_NAME: <key-ring-name>
KEY_NAME: <key-name>
```

Cloud Functionsのデプロイをします.
```
$ gcloud functions deploy Subscription \
  --runtime go111 \
  --trigger-topic <topic-name> \
  --env-vars-file env.yaml
```

以下のコマンドを実行してCloud Schedulerを使って定期実行するようにします.
以下の例では5分毎に実行しています.
```
$ gcloud beta scheduler jobs create pubsub <scheduler-name> \
  --schedule '*/5 * * * *' \
  --topic <topic-name> \
  --message-body 'slack-spotify-status' \
  --time-zone 'Asia/Tokyo'
```
