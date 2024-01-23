package config

import "path/filepath"

type HiddenConfig struct {
	Path           string
	ConfigFilename string

	Config Config
}

// represents a config file the hidden config folder
type Config struct {
	FfmpegRecordingsPath string `yaml:"ffmpegRecordingsPath"`
	AvfoundationDevice   string `yaml:"avfoundationDevice"`
}

const (
	HiddenConfigDirName = ".config/block-cli"
	ConfigFileName      = "config.yaml"

	DefaultFfmpegRecordingsPath = "."
	DefaultAvfoundationDevice   = "1:0"
)

func NewHiddenConfig(homeDir string) HiddenConfig {
	config := Config{
		FfmpegRecordingsPath: DefaultFfmpegRecordingsPath,
		AvfoundationDevice:   DefaultAvfoundationDevice,
	}

	return HiddenConfig{
		Path:           filepath.Join(homeDir, HiddenConfigDirName),
		ConfigFilename: ConfigFileName,
		Config:         config,
	}
}

func GetFfmpegRecordingPath() string {
	return Cfg.HiddenConfig.Config.FfmpegRecordingsPath
}

func GetAvfoundationDevice() string {
	return Cfg.HiddenConfig.Config.AvfoundationDevice
}
