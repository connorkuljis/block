package config

import (
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	HiddenConfig HiddenConfig
	RootConfig   RootConfig
}

var Cfg AppConfig

func InitConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	hiddenConfig := NewHiddenConfig(homeDir)
	rootConfig := NewRootConfig(homeDir)

	slog.Debug("Checking: " + hiddenConfig.Path)
	if err := makeDirIfNotExists(hiddenConfig.Path); err != nil {
		return err
	}

	slog.Debug("Checking: " + rootConfig.Path)
	if err := makeDirIfNotExists(rootConfig.Path); err != nil {
		return err
	}

	slog.Debug("Checking: " + hiddenConfig.ConfigFilename)
	if err := loadOrMakeConfigFileIfNotExists(hiddenConfig); err != nil {
		return err
	}

	Cfg = AppConfig{
		HiddenConfig: hiddenConfig,
		RootConfig:   rootConfig,
	}

	return nil
}

func makeDirIfNotExists(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func loadOrMakeConfigFileIfNotExists(h HiddenConfig) error {
	_, err := os.Stat(filepath.Join(h.Path, h.ConfigFilename))
	if os.IsNotExist(err) {
		if err = makeDefaultConfigFile(h); err != nil {
			return err
		}
		return nil
	} else {
		loadConfig(h)
		sanitiseConfigValues(h)
	}

	return nil
}

// create config file and write default values
func makeDefaultConfigFile(h HiddenConfig) error {
	configFile, err := os.Create(filepath.Join(h.Path, h.ConfigFilename))
	if err != nil {
		return err
	}
	defer configFile.Close()

	data, err := yaml.Marshal(&h.Config)
	if err != nil {
		return err
	}

	_, err = configFile.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func loadConfig(h HiddenConfig) error {
	contents, err := os.ReadFile(h.Path)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(contents, &h.Config); err != nil {
		return err
	}
	return nil
}

func sanitiseConfigValues(h HiddenConfig) error {
	if _, err := os.Stat(h.Config.FfmpegRecordingsPath); err != nil {
		return err
	}
	return nil
}
