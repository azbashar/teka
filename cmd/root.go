package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var fileArg, mainFileArg string

var rootCmd = &cobra.Command{
	Use:   "teka",
	Short: "Teka is your HLedger helper",
	Long: `Teka helps you add transactions and manage your ledger with ease.`,
}

func Execute() {
	rootCmd.PersistentFlags().StringP("file", "f", "", "Ledger file to write to")
	rootCmd.PersistentFlags().StringP("mainfile", "m", "", "Main file to write to")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fileArg = rootCmd.Flag("file").Value.String()
	mainFileArg = rootCmd.Flag("mainfile").Value.String()
}