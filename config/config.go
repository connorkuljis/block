package config

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

var Cfg Config

func InitConfig() error {
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

	cfg := Config{
		Path:                 path,
		Yaml:                 yaml,
		FfmpegRecordingsPath: ".",
		AvfoundationDevice:   "1:0",
	}

	if err := makeDirectoryIfNotExists(cfg); err != nil {
		return err
	}

	if err := createOrLoadYamlFileIfNotExists(cfg); err != nil {
		return err
	}

	// package global
	Cfg = cfg

	return nil
}

func makeDirectoryIfNotExists(cfg Config) error {
	_, err := os.Stat(cfg.Path)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(cfg.Path, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func createOrLoadYamlFileIfNotExists(cfg Config) error {
	_, err := os.Stat(cfg.Yaml)
	if os.IsNotExist(err) {
		if err = writeDefaultConfigurationYaml(cfg); err != nil {
			return err
		}
		return nil
	} else {
		loadConfig(cfg)
		sanitiseConfigValues(cfg)
	}
	return nil
}

// create config file and write default values
func writeDefaultConfigurationYaml(cfg Config) error {
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

func loadConfig(cfg Config) error {
	contents, err := os.ReadFile(cfg.Yaml)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(contents, &cfg); err != nil {
		return err
	}
	return nil
}

func sanitiseConfigValues(cfg Config) error {
	if _, err := os.Stat(cfg.FfmpegRecordingsPath); err != nil {
		return err
	}
	return nil
}
