package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Path                 string
	Yaml                 string
	FfmpegRecordingsPath string `yaml:"ffmpegRecordingsPath"`
	AvfoundationDevice   string `yaml:"avfoundationDevice"`
}

func initConfig() error {
	const (
		ConfigDir = ".config/block-cli"
		YamlFile  = "config.yaml"
	)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	path := filepath.Join(homeDir, ConfigDir)
	yaml := filepath.Join(path, YamlFile)

	cfg = Config{
		Path:                 path,
		Yaml:                 yaml,
		FfmpegRecordingsPath: ".",
		AvfoundationDevice:   "1:0",
	}

	if err := makeDirectoryIfNotExists(); err != nil {
		return err
	}

	if err := createOrLoadYamlFileIfNotExists(); err != nil {
		return err
	}

	return nil
}

func makeDirectoryIfNotExists() error {
	_, err := os.Stat(cfg.Path)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(cfg.Path, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func createOrLoadYamlFileIfNotExists() error {
	_, err := os.Stat(cfg.Yaml)
	if os.IsNotExist(err) {
		if err = writeDefaultConfigurationYaml(); err != nil {
			return err
		}
		return nil
	} else {
		loadConfig()
		sanitiseConfigValues()
	}
	return nil
}

// create config file and write default values
func writeDefaultConfigurationYaml() error {
	configFile, err := os.Create(cfg.Yaml)
	if err != nil {
		return err
	}
	defer configFile.Close()
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	_, err = configFile.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func loadConfig() error {
	contents, err := os.ReadFile(cfg.Yaml)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(contents, &cfg); err != nil {
		return err
	}
	return nil
}

func sanitiseConfigValues() error {
	if _, err := os.Stat(cfg.FfmpegRecordingsPath); err != nil {
		return err
	}
	return nil
}
