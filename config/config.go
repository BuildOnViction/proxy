package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Masternode   []string `json:"Masternode,omitempty"`
	Fullnode     []string `json:"Fullnode,omitempty"`
	Websocket    []string `json:"Websocket,omitempty"`
	WsServerName string   `json:"WsServerName,omitempty"`
	Certs        []Certs  `json:"Certs,omitempty"`
	*Headers     `json:"Headers,omitempty"`
}

type Headers map[string]string

type Certs struct {
	Crt string
	Key string
}

var config Config

func Init(configFile string) {
	jsonFile, _ := os.Open(configFile)
	defer jsonFile.Close()
	decoder := json.NewDecoder(jsonFile)
	_ = decoder.Decode(&config)
}

func GetConfig() Config {
	return config
}
