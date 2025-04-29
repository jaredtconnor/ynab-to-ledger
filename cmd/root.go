package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	outputFile  string
	mappingFile string
	rootCmd     = &cobra.Command{
		Use:   "ynab_to_ledger [file]",
		Short: "Convert YNAB export to Ledger format",
		Long: `Convert a YNAB (You Need a Budget) export file to a Ledger journal format.
This tool processes CSV files exported from YNAB and creates a journal file 
that can be used with Ledger or hledger accounting systems.

The input file should be the Register CSV export from YNAB, with dates 
in mm/dd/yyyy format and numbers using a period (.) as the decimal separator.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return convertFile(args[0], outputFile)
		},
	}

	genCoaCmd = &cobra.Command{
		Use:   "gen-coa [register.csv] [coa.yaml]",
		Short: "Generate a Chart of Accounts YAML from a YNAB Register CSV",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GenerateCOA(args[0], args[1])
		},
	}
)

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "ynab_ledger.dat", "output file path")
	rootCmd.Flags().StringVarP(&mappingFile, "mapping", "m", "coa.yaml", "chart of accounts mapping file")
	rootCmd.AddCommand(genCoaCmd)
}
