package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/A-Bashar/Teka-Finance/internal/prompt"
	"github.com/spf13/cobra"
)

type LineType int

const (
    LineTransaction LineType = iota
    LinePosting
    LineComment
)

type Line struct {
    Type    LineType
    Date    string // for LineTransaction
    Note    string // for LineTransaction
    Account string // for LinePosting
    Amount  string // for LinePosting
    Text    string // for LineComment
    Indent  bool   // true for ';', false for '#'
}

type Transaction struct {
    Lines []Line
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
		
		// Date
		for {
			date := prompt.Ask("Date?")
			if date == "" {
				fmt.Println("Abort.")
				return
			}

			if date == ";" || date == "#" {
				comment := prompt.Ask("Comment?")
				tx.Lines = append(tx.Lines, Line {
					Type:   LineComment,
					Text:   comment,
					Indent: false,
				})
				continue
			}

			// Parse date shortcuts
			date, err := prompt.ParseDate(date)
			if err != nil {
				fmt.Println(err)
				return
			}

			// Note
			note := prompt.Ask("Note?")
			tx.Lines = append(tx.Lines, Line {
				Type: LineTransaction,
				Date: date,
				Note: note,
			})
			break
		}

		// Postings
		for {
			account := prompt.Ask("Account?")
			if account == "" {
				break
			}

			// Account search
    		if strings.HasPrefix(account, ".") && len(account) > 1 {
				searchTerm := account[1:]
				selected, err := prompt.SearchAccounts(searchTerm, file)
				if err != nil {
					fmt.Println("Error searching accounts:", err)
					continue
				}
				if selected == "" {
					continue
				}
				account = selected
			}

			// Comment
			if account == ";" || account == "#" {
				comment := prompt.Ask("Comment?")
				tx.Lines = append(tx.Lines, Line {
					Type:   LineComment,
					Text:   comment,
					Indent: account == ";",
				})
				continue
			}

			amount := prompt.Ask("Amount?")
			tx.Lines = append(tx.Lines, Line {
				Type:    LinePosting,
				Account: account,
				Amount:  amount,
			})
		}

		// Format the transaction as text
		content := "\n"
		for _, line := range tx.Lines {
			switch line.Type {
				case LineComment:
					if line.Indent {
						content += "    " + line.Text + "\n"
					} else {
						content += line.Text + "\n"
					}
				case LineTransaction:
					content += fmt.Sprintf("%s %s\n", line.Date, line.Note)
				case LinePosting:
					content += fmt.Sprintf("    %-20s %s\n", line.Account, line.Amount)
			}
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