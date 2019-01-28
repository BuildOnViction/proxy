package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Masternode []string
	Fullnode   []string
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
