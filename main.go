package main

import (
	"log"
	"path/filepath"

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

func init() {
	rootCmd.AddCommand(
		startCmd,
		historyCmd,
		deleteTaskCmd,
		resetDNSCmd,
		generateCmd,
	)

	rootCmd.PersistentFlags().BoolVarP(&flags.DisableBlocker, "no-block", "d", false, "Do not block hosts file.")
	rootCmd.PersistentFlags().BoolVarP(&flags.ScreenRecorder, "screen-recorder", "x", false, "Enable screen recorder.")
	rootCmd.PersistentFlags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Logs additional details.")
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

	if err = rootCmd.Execute(); err != nil {
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
