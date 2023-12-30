package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var cfg Config

type Config struct {
	Directory string `yaml:"directory"`
}

func ReadConfig() {
	var config Config

	homeDir, err := os.UserHomeDir()

	configFilePath := ".config/block-cli/config.yaml"

	filepath := filepath.Join(homeDir, configFilePath)

	// Read the contents of the file
	data, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}

	fmt.Println(config.Directory)

	cfg = config
}
