package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	TwitterKey      string `json:"twitter_key"`
	TwitterSecret   string `json:"twitter_secret"`
	TwitterCallback string `json:"twitter_callback"`

	MinterMnemonic   string `json:"minter_mnemonic"`
	MinterNodeUrl    string `json:"minter_node_url"`
	MinterPushBearer string `json:"minter_push_bearer"`

	Coin   string `json:"coin"`
	Amount string `json:"amount"`
	Sender string `json:"sender"`

	TweetID int64 `json:"tweet_id"`

	ListenAt string `json:"listen_at"`
}

func ParseConfig(path string) *Config {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	cfg := Config{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	return &cfg
}
