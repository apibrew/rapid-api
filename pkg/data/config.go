package data

import (
	"encoding/json"
	"os"
)

type DynamodbConfig struct {
	Region          string `json:"region"`
	AccessKeyID     string `json:"access-key-id"`
	SecretAccessKey string `json:"secret-access-key"`
	TableName       string `json:"table-name"`
}

func LoadDynamodbConfig(path string) (DynamodbConfig, error) {
	cfg := DynamodbConfig{}

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
