package cmd

import (
	"github.com/connorkuljis/task-tracker-cli/server"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serves http server.",
	Run: func(cmd *cobra.Command, args []string) {
		server.Serve()
	},
}
