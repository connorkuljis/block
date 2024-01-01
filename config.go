package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const APP_DIR = ".config/block-cli"
const CONFIG_FILE = "config.yaml"

type UserConfig struct {
	FfmpegRecordingsPath string  `yaml:"ffmpegRecordingsPath"`
	DefaultDuration      float64 `yaml:"defaultDuration"`
	AppInfo              AppInfo
}

type AppInfo struct {
	HomeDir       string
	AppDir        string
	ConfigFileDir string
}

func initConfig() {
	setAppInfo()

	checkHealth()

	bytes, err := os.ReadFile(config.AppInfo.ConfigFileDir)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	err = yaml.Unmarshal([]byte(bytes), &config)
	if err != nil {
		log.Fatalf("Error, initialising config file: %v", err)
	}

	if config.DefaultDuration <= 0.0 {
		log.Fatalf("Error, default duration must be greater than 0.0, given: %f", config.DefaultDuration)
	}

	_, err = os.Stat(config.FfmpegRecordingsPath)
	if err != nil {
		log.Fatalf("Error, %s does not exist: %v", config.FfmpegRecordingsPath, err)
	}

	if verbose {
		log.Printf("Config: %v", config)
	}
}

func setAppInfo() {
	var appInfo AppInfo

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error creating config dir: %v", err)
	}

	appInfo.HomeDir = homeDir
	appInfo.AppDir = filepath.Join(homeDir, APP_DIR)
	appInfo.ConfigFileDir = filepath.Join(homeDir, APP_DIR, CONFIG_FILE)

	config.AppInfo = appInfo
}

func checkHealth() {
	_, err := os.Stat(config.AppInfo.AppDir)
	if err != nil {
		createAppDirectory(config.AppInfo.AppDir)
	}

	_, err = os.Stat(config.AppInfo.ConfigFileDir)
	if err != nil {
		createConfig(config.AppInfo.ConfigFileDir)
	}
}

func createAppDirectory(path string) {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Fatalf("Error creating config dir: %v", err)
	}
	fmt.Printf("Created directory: %s\n", path)
}

func createConfig(path string) {
	file, err := os.Create(path)
	if err != nil {
		log.Fatalf("Error creating config file: %v", err)

	}
	defer file.Close()

	writeDefaultConfig(file)
	fmt.Printf("Written default config: %s\n", path)
}

func writeDefaultConfig(file *os.File) {
	var defaultConfig = `# Configuration file for block-cli app.
ffmpegRecordingsPath: ~/Downloads
defaultDuration: 10.0`

	_, err := file.WriteString(defaultConfig)
	if err != nil {
		log.Fatalf("Error creating config file: %v", err)

	}
}
