package cmd

import (
	"log"

	"github.com/connorkuljis/block-cli/blocker"
	"github.com/spf13/cobra"
)

var resetDNSCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset DNS cache.",
	Run: func(cmd *cobra.Command, args []string) {
		err := blocker.ResetDNS()
		if err != nil {
			log.Println(err)
		}
	},
}
