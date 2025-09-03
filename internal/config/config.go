package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	AmountColumn int `toml:"amount_column"`
	Accounts struct {
		ConversionAccount string `toml:"conversion"`
		FXGainAccount string `toml:"fx_gain"`
		FXLossAccount string `toml:"fx_loss"`
	}`toml:"accounts"`
	EfficientFileStructure struct {
		Enabled bool `toml:"enable"`
		FilesRoot string `toml:"files_root"`
	} `toml:"efficient_file_structure"`
}

var Cfg Config

// Load config into Cfg
func LoadConfig() error {
	// Look for config in user folder
	confPath, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}
	confPath = filepath.Join(confPath, "teka")
	_ = os.MkdirAll(confPath, 0700) // create if it doesnt exist
	configFile := filepath.Join(confPath, "tekaconf.toml")

	// Read config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// create default config if it doesnt exist
			Cfg = Config{
				AmountColumn: 40,
				Accounts: struct {
					ConversionAccount string `toml:"conversion"`
					FXGainAccount     string `toml:"fx_gain"`
					FXLossAccount     string `toml:"fx_loss"`
				}{
					ConversionAccount: "equity:conversion",
					FXGainAccount:     "income:fx gain",
					FXLossAccount:     "expenses:fx loss",
				},
				EfficientFileStructure: struct {
					Enabled bool `toml:"enable"`
					FilesRoot string `toml:"files_root"`
				} {
					Enabled: false,
					FilesRoot: "~/finance/",
				},
			}
			fmt.Println("No config file found.")
			return SaveConfig(configFile)
		}
		return err
	}

	if err := toml.Unmarshal(data, &Cfg); err != nil {
		return fmt.Errorf("invalid config file.\nfailed to parse file: %w", err)
	}

	return nil
}

// Save current Cfg into config file
func SaveConfig(configFile string) error {
	data, err := toml.Marshal(Cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	fmt.Println("Creating config file in: "+configFile)
	return os.WriteFile(configFile, data, 0644)
}