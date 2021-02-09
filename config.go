package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/configor"
)

var config = loadConfig()

// Configuration ...
type Configuration struct {
	BaseURL    string `required:"true"`
	Username   string
	DataFile   string
	BackupPath string
}

var conf = Configuration{
	BaseURL:    "",
	Username:   "admin",
	DataFile:   "./data.json",
	BackupPath: "./backups",
}

func loadConfig() Configuration {
	if err := configor.Load(&conf, "./config.json"); err != nil {
		log.Fatal(err)
	}
	return conf
}
