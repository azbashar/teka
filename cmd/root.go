package cmd

import (
	"os"
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "teka",
	Short: "Teka is your HLedger helper",
	Long: `Teka helps you add transactions and manage your ledger with ease.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}