package main

import (
	"context"
	"embed"
	"log"
	"log/slog"
	"os"

	"github.com/connorkuljis/block-cli/internal/commands"
	"github.com/connorkuljis/block-cli/internal/config"
	"github.com/connorkuljis/block-cli/internal/db"

	"github.com/urfave/cli/v2"
)

//go:embed www
var www embed.FS

func main() {
	// TODO: Only load config if command requires it.

	err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Loaded config.")

	db, err := db.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Loaded db.")

	app := &cli.App{
		Name:  "block",
		Usage: "block-cli blocks distractions from the command line. track tasks and capture your screen.",
		Before: func(c *cli.Context) error {
			c.Context = context.WithValue(c.Context, "db", db)
			c.Context = context.WithValue(c.Context, "www", www)
			return nil
		},
		// TODO: Refactor out cli commands to a seperate module, with one command per file.
		Commands: []*cli.Command{
			commands.StartCmd,
			commands.HistoryCmd,
			commands.DeleteTaskCmd,
			commands.ServeCmd,
			commands.GenerateCmd,
			commands.ResetDNSCmd,
			commands.UpCmd,
			commands.DownCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
