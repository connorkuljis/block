package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var resetDNSCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset DNS cache.",
	Run: func(cmd *cobra.Command, args []string) {
		flags.Verbose = true
		err := ResetDNS()
		if err != nil {
			log.Println(err)
		}
	},
}
