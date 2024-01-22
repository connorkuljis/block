package main

import (
	"log"
	"path/filepath"

	"github.com/connorkuljis/task-tracker-cli/cmd"
	"github.com/jmoiron/sqlx"
)

const DBName = "app_data.db"

var (
	flags Flags
	cfg   Config
	db    *sqlx.DB
)

type Flags struct {
	DisableBlocker bool
	ScreenRecorder bool
	Verbose        bool
}

func main() {
	var err error

	if err := initConfig(); err != nil {
		log.Fatal(err)
	}

	if err = initDB(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func initDB() error {
	var err error

	db, err = sqlx.Connect("sqlite3", filepath.Join(cfg.Path, DBName))
	if err != nil {
		return err
	}

	_, err = db.Exec(schema)
	if err != nil {
		return err
	}

	return nil
}
