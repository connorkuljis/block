package main

import (
	"context"
	"embed"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/connorkuljis/block-cli/internal/commands"
	"github.com/connorkuljis/block-cli/internal/config"
	"github.com/connorkuljis/block-cli/internal/db"
	"github.com/connorkuljis/block-cli/internal/ffmpeg"

	"github.com/urfave/cli/v2"
)

//go:embed www
var www embed.FS

func main() {
	stop := make(chan int, 1)

	// Start the screen recording in a goroutine
	go func() {
		err := ffmpeg.RecordScreen("0", "output.mkv", stop)
		if err != nil {
			log.Println("Error in RecordScreen:", err)
		}
	}()

	// Wait for 5 seconds
	time.Sleep(10 * time.Second)

	// Send stop signal
	log.Println("Sending stop signal after 5 seconds")
	stop <- 1

	// Wait a bit to allow for cleanup
	time.Sleep(1 * time.Second)
	log.Println("Main function exiting")
}

func start() {
	err := config.InitConfig() // TODO: Only load config if command requires it.
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
