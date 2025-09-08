package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Accounts struct {
	ConversionAccount string `yaml:"conversion"`
	FXGainAccount     string `yaml:"fx_gain"`
	FXLossAccount     string `yaml:"fx_loss"`
}

type EfficientFileStructure struct {
	Enabled   bool   `yaml:"enable"`
	FilesRoot string `yaml:"files_root"`
}

type StarredAccount struct {
	DisplayName string `yaml:"display_name"`
	Account     string `yaml:"account"`
}

type Config struct {
	AmountColumn           int                    `yaml:"amount_column"`
	Accounts               Accounts               `yaml:"accounts"`
	EfficientFileStructure EfficientFileStructure `yaml:"efficient_file_structure"`
	StarredAccounts        []StarredAccount       `yaml:"starred_accounts"`
}

var Cfg Config

func LoadConfig() error {
	confPath, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}
	confPath = filepath.Join(confPath, "teka")
	_ = os.MkdirAll(confPath, 0700)
	configFile := filepath.Join(confPath, "tekaconf.yaml")

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// create default config if it doesn't exist
			Cfg = Config{
				AmountColumn: 40,
				Accounts: Accounts{
					ConversionAccount: "equity:conversion",
					FXGainAccount:     "income:fx gain",
					FXLossAccount:     "expenses:fx loss",
				},
				EfficientFileStructure: EfficientFileStructure{
					Enabled:   false,
					FilesRoot: "~/finance/",
				},
				StarredAccounts: []StarredAccount{
					{DisplayName: "Cash Wallet", Account: "assets:cash"},
					{DisplayName: "Bank", Account: "assets:bank"},
				},
			}
			fmt.Println("No config file found.")
			return SaveConfig(configFile)
		}
		return err
	}

	if err := yaml.Unmarshal(data, &Cfg); err != nil {
		return fmt.Errorf("invalid config file.\nfailed to parse file: %w", err)
	}

	return nil
}

func SaveConfig(configFile string) error {
	data, err := yaml.Marshal(Cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	fmt.Println("Creating config file in: " + configFile)
	return os.WriteFile(configFile, data, 0644)
}
