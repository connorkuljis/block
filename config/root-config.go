package config

import (
	"path/filepath"
)

type RootConfig struct {
	Path       string
	DbFileName string
}

const (
	RootConfigDirName = ".block-cli"
	DbName            = "app_data.db"
)

func NewRootConfig(homeDir string) RootConfig {
	return RootConfig{
		Path:       filepath.Join(homeDir, RootConfigDirName),
		DbFileName: DbName,
	}
}

func GetDBPath() string {
	return filepath.Join(Cfg.RootConfig.Path, Cfg.RootConfig.DbFileName)
}
