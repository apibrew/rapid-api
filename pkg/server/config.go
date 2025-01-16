package server

import (
	"encoding/json"
	"os"
)

type Config struct {
	ListenAddr string `json:"listenAddr"`
}

func LoadConfig(path string) (Config, error) {
	cfg := Config{}

	fp, err := os.Open(path)

	if err != nil {
		return cfg, err
	}

	err = json.NewDecoder(fp).Decode(&cfg)

	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
