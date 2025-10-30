package main

import (
	"fmt"
	"os"

	"github.com/azbashar/teka/cmd"
	"github.com/azbashar/teka/internal/config"
)

func main() {
	// load the config
	if err := config.LoadConfig(); err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	cmd.Execute()
}
