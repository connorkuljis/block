package commands

import (
	"embed"

	"github.com/connorkuljis/block-cli/internal/server"
	"github.com/jmoiron/sqlx"
	"github.com/urfave/cli/v2"
)

var ServeCmd = &cli.Command{
	Name:  "serve",
	Usage: "Serves http server.",
	Action: func(ctx *cli.Context) error {
		db := ctx.Context.Value("db").(*sqlx.DB)
		www := ctx.Context.Value("www").(embed.FS)

		templatesPath := "www/templates"
		staticPath := "www/static"
		s, err := server.NewServer(www, db, "8080", templatesPath, staticPath)

		s.Routes()

		err = s.ListenAndServe()
		if err != nil {
			return err
		}

		return nil
	},
}
