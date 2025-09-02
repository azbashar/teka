package main

import (
	"fmt"
	"os"

	"github.com/A-Bashar/Teka-Finance/cmd"
	"github.com/A-Bashar/Teka-Finance/internal/config"
)

func main() {
	// load the config
	if err := config.LoadConfig(); err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	cmd.Execute()
}