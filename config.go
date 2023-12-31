package main

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var cfg Config
var appDir = ".config/block-cli"
var configFile = "config.yaml"

type Config struct {
	Recordings string `yaml:"directory"`
	Database   string `yaml:"database"`
}

func FindFileInConfigDir(target string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(home, appDir, target)

}

func ReadConfig() {
	c := FindFileInConfigDir(configFile)
	bytes, err := os.ReadFile(c)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	err = yaml.Unmarshal([]byte(bytes), &cfg)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}

	log.Println(cfg)
}
