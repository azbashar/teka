package cmd

import (
	"fmt"

	"github.com/A-Bashar/Teka-Finance/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Prints out the path to the config file",
	Run: func(cmd *cobra.Command, args []string) {
		configPath, err := config.GetConfigPath()
		if err != nil {
			fmt.Println("Error getting config path:", err)
			return
		}
		fmt.Println(configPath)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}