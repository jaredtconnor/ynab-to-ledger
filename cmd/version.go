package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version holds the current version of the application
const version = "1.0.0"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version number of YNAB to Ledger converter`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("YNAB to Ledger Converter v%s\n", version)
	},
}
