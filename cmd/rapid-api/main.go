package main

import (
	"flag"
	"github.com/apibrew/rapid-api/pkg/data"
	"github.com/apibrew/rapid-api/pkg/server"
	log "github.com/sirupsen/logrus"
)

func main() {
	var configPath = flag.String("config", "", "Config file")

	flag.Parse()

	if *configPath == "" {
		configPath = new(string)
		*configPath = "config.json"
	}

	serverCfg, err := server.LoadConfig(*configPath)

	if err != nil {
		log.Fatal("Could not read config file", err)
	}

	dynamodbCfg, err := data.LoadDynamodbConfig(*configPath)

	if err != nil {
		log.Fatal("Could not read config file", err)
	}

	di := data.NewDynamoDBInterface(dynamodbCfg)

	if err := di.Init(); err != nil {
		log.Fatal("Could not initialize data interface", err)
	}

	srv := &server.Server{
		DataInterface: di,
		Config:        serverCfg,
	}

	if err := srv.Run(); err != nil {
		log.Fatal("Could not start server", err)
	}
}
