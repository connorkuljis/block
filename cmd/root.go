package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "block",
	Short: "Block removes distractions when you work on tasks.",
	Long: `
Block saves you time by blocking websites at IP level.
Progress bar is displayed directly in the terminal. 
Automatically unblock sites when the task is complete.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

func init() {
	rootCmd.AddCommand(
		startCmd,
		historyCmd,
		deleteTaskCmd,
		resetDNSCmd,
		generateCmd,
		serveCmd,
	)

	rootCmd.PersistentFlags().BoolP("noBlock", "f", false, "Do not block hosts file.")
	rootCmd.PersistentFlags().BoolP("screenRecorder", "x", false, "Enable screen recorder.")
	rootCmd.PersistentFlags().Bool("debug", false, "Logs additional details.")

}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
