package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/A-Bashar/Teka-Finance/internal/prompt"
	"github.com/spf13/cobra"
)

type Posting struct {
	Account string
	Amount  string
}

type Transaction struct {
	Date        string
	Note string
	Postings    []Posting
}

var file string

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new transaction to your ledger",
	Run: func(cmd *cobra.Command, args []string) {
		// Get ledger file from ENV if not provided
		if file == "" {
			file = os.Getenv("LEDGER_FILE")
			if file == "" {
				fmt.Println("No ledger file specified. Use --file flag or set LEDGER_FILE environment variable.")
				return
			}
		}
		// Collect transaction data
		tx := Transaction{}
		tx.Date = prompt.Ask("Date?")
		tx.Note = prompt.Ask("Note?")

		for {
			acc := prompt.Ask("Account?")
			// End transaction on empty account
			if acc == "" {
				break
			}
			amt := prompt.Ask("Amount?")
			tx.Postings = append(tx.Postings, Posting {
				Account: acc,
				Amount:  amt,
			})
		}

		// Format the transaction as text
		content := fmt.Sprintf("\n%s %s\n", tx.Date, tx.Note)
		for _, p := range tx.Postings {
			content += fmt.Sprintf("    %-40s %s\n", p.Account, p.Amount)
		}

		// Display the collected transaction
		fmt.Printf("\nAdding this following transaction to %s:\n", file)
		fmt.Printf("%s\n", content)

		// Confirm before writing
		if !prompt.Confirm("Is this correct") {
			fmt.Println("Transaction discarded.")
			return
		}

		// Store previous state for potential revert
		info, err := os.Stat(file)
		prevSize := int64(0)
		if err == nil {
			prevSize = info.Size()
		}


		// Save transaction to file
		f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			return
		}
		defer f.Close()
		// Write the content
		if _, err := f.WriteString(content); err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
			return
		}
		fmt.Println("Validating transaction...")
		f.Close()
		// Validate changes
		cmdArgs := []string{"check"}
		if file != "" {
			cmdArgs = append(cmdArgs, "-f", file)
		}
		out, err := exec.Command("hledger", cmdArgs...).CombinedOutput()
		if err != nil {
			fmt.Println("Error validating ledger:")
			fmt.Println(string(out))

			if prompt.Confirm("Do you want to revert the changes?") {
				revertErr := os.Truncate(file, prevSize)
				if revertErr != nil {
					fmt.Printf("Error reverting changes: %v\n", revertErr)
				} else {
					fmt.Println("Changes reverted.")
				}
				return
			} else {
				fmt.Println("Changes kept despite validation errors.")
			}
		} else {
			fmt.Println("Transaction added successfully.")
		} 
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&file, "file", "f", "", "Ledger file to write to")
}