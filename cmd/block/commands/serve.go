package commands

import (
	"log"
	"net/http"

	"github.com/connorkuljis/block-cli/pkg/server"
	"github.com/urfave/cli/v2"
)

var ServeCmd = &cli.Command{
	Name:  "serve",
	Usage: "Serves http server.",
	Action: func(ctx *cli.Context) error {
		s := server.NewServer("8080")

		s.AppData = server.AppData{
			Title:   "Block CLI",
			DevMode: false, // load from env
		}

		err := s.Routes()
		if err != nil {
			return err
		}

		log.Println("[ 💿 Spinning up server on http://localhost:" + s.Port + " ]")
		if err := http.ListenAndServe(":"+s.Port, s.Router); err != nil {
			return err
		}

		return nil
	},
}
