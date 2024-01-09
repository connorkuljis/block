package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	FfmpegRecordingsPath string `yaml:"ffmpegRecordingsPath"`
	AvfoundationDevice   string `yaml:"avfoundationDevice"`
}

const (
	ConfigDir = ".config/block-cli"
	YamlFile  = "config.yaml"
)

func InitConfig() (Config, error) {
	var config Config
	config.FfmpegRecordingsPath = "."
	config.AvfoundationDevice = "1:0"

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config, err
	}

	configPath := filepath.Join(homeDir, ConfigDir)
	configYaml := filepath.Join(configPath, YamlFile)

	// check config path exists, if not create it
	_, err = os.Stat(configPath)
	if err != nil {
		err := os.MkdirAll(configPath, os.ModePerm)
		if err != nil {
			return config, err
		}
	}

	_, err = os.Stat(configYaml)
	if err != nil {
		err = writeDefaults(&config, configYaml)
		if err != nil {
			return config, err
		}
	} else {
		err = loadConfig(&config, configYaml)
		if err != nil {
			return config, err
		}
	}

	err = sanitiseConfigValues(config)
	if err != nil {
		return config, err
	}

	log.Println(config)
	return config, nil
}

func sanitiseConfigValues(config Config) error {
	// check if the ffmpeg recording paths exists
	_, err := os.Stat(config.FfmpegRecordingsPath)
	if err != nil {
		return fmt.Errorf("Error sanitising configuration file values -> %v", err)
	}

	validInput := "1:0"
	if config.AvfoundationDevice != validInput {
		return fmt.Errorf("Error, invalid avfoundation input device. have '%s',  expected '%s...'", config.AvfoundationDevice, validInput)
	}

	return nil
}

// create config file and write default values
func writeDefaults(config *Config, configYaml string) error {
	configFile, err := os.Create(configYaml)
	if err != nil {
		return err
	}
	defer configFile.Close()

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	_, err = configFile.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func loadConfig(config *Config, configYaml string) error {
	contents, err := os.ReadFile(configYaml)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(contents, config)
	if err != nil {
		return err
	}

	return nil
}
