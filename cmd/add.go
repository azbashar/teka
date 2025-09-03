package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/A-Bashar/Teka-Finance/internal/config"
	"github.com/A-Bashar/Teka-Finance/internal/efficientfiles"
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
var noFileArg = false

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new transaction to your ledger",
	Run: func(cmd *cobra.Command, args []string) {
		// Get ledger file from ENV if not provided
		if file == "" {
			noFileArg = true
		}
		if noFileArg && !config.Cfg.EfficientFileStructure.Enabled {
			file = os.Getenv("LEDGER_FILE")
			if file == "" {
				fmt.Println("No ledger file specified. Use --file flag or set LEDGER_FILE environment variable.")
				return
			}
		}
		file = "-f "+ file
		// Collect transaction data
		tx := Transaction{}
		
		// Date
		for {
			date := Ask("Date?")
			if date == "" {
				fmt.Println("Abort.")
				return
			}

			// Comment
			if date == ";" || date == "#" {
				comment := Ask("Comment?")
				tx.Lines = append(tx.Lines, Line {
					Type:   LineComment,
					Text:   comment,
					Indent: false,
				})
				continue
			}

			// Parse date shortcuts
			date, err := ParseDate(date)
			if err != nil {
				fmt.Println(err)
				return
			}

			if noFileArg && config.Cfg.EfficientFileStructure.Enabled {
				file, err = efficientfiles.GetCurrentFile(date)
				if err != nil {
					fmt.Printf("Efficient file finder error: %v", err)
					return
				}
			}

			// Note
			AskNote:
				note := Ask("Note?") 

				// Note search
				if strings.HasPrefix(note, ".") {
					var searchTerm string
					if note == "." {
						searchTerm = ""
					} else {
						searchTerm = note[1:]
					}
					selected, err := SearchRecords("notes",searchTerm, file)
					if err != nil {
						fmt.Println("Error searching notes:", err)
						goto AskNote
					}
					if selected == "" {
						goto AskNote
					}
					note = selected
				}

			tx.Lines = append(tx.Lines, Line {
				Type: LineTransaction,
				Date: date,
				Note: note,
			})
			break
		}

		// Postings
		for {
			// Account
			account := Ask("Account?")
			if account == "" {
				break
			}

			// Account search
    		if strings.HasPrefix(account, ".") {
				var searchTerm string
				if account == "." {
					searchTerm = ""
				} else {
					searchTerm = account[1:]
				}
				selected, err := SearchRecords("accounts",searchTerm, file)
				if err != nil {
					fmt.Println("Error searching accounts:", err)
					continue
				}
				if selected == "" {
					continue
				}
				account = selected
			}

			// Check if converting currencies
			if strings.HasPrefix(account,"$") {
				err := convertCurrencies(&tx, account)
				if err != nil {
					fmt.Println("Can not calculate gain:", err, "\nTry adding the transacion manually.")
					return
				}
				break
			}

			// Comment
			if account == ";" || account == "#" {
				comment := Ask("Comment?")
				tx.Lines = append(tx.Lines, Line {
					Type:   LineComment,
					Text:   comment,
					Indent: account == ";",
				})
				continue
			}

			// Amount
			AskAmount:
				amount := Ask("Amount?")
				// Auto balance
				if amount == "." {
					bal, err := calculateBalanceAmount(&tx)
					if err != nil {
						fmt.Println("Error:", err)
						goto AskAmount
					}
					amount = bal
				}
			
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
					// negative amounts will align with the posetive amounts
					column := config.Cfg.AmountColumn
					if strings.HasPrefix(line.Amount, "-") {
						column = config.Cfg.AmountColumn - 1
					}
					content += fmt.Sprintf(
						"    %-*s %s\n",
						column,
						line.Account, 
						line.Amount,
					)
			}
		}

		// Display the collected transaction
		fmt.Printf("\nAdding this following transaction to %s:\n", file)
		fmt.Printf("%s\n", content)

		// Confirm before writing
		if !Confirm("Is this correct") {
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
		f, err := os.OpenFile(strings.TrimPrefix(file,"-f "), os.O_APPEND|os.O_WRONLY, 0644)
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
			cmdArgs = append(cmdArgs, file)
		}
		out, err := exec.Command("hledger", cmdArgs...).CombinedOutput()
		if err != nil {
			fmt.Println("Error validating ledger:")
			fmt.Println(string(out))

			if Confirm("Do you want to revert the changes?") {
				revertErr := os.Truncate(strings.TrimPrefix(file, "-f "), prevSize)
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

// Prompt for data
func Ask(question string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(question + " ")
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// Yes no confirm prompt
func Confirm(question string) bool {
	answer := Ask(question + " (Y/n)?")
	answer = strings.ToLower(strings.TrimSpace(answer))

	// Empty answer defaults to yes
	if answer == "" || answer == "y" || answer == "yes" {
		return true
	}
	return false
}

// Parse shortcut dates
func ParseDate(input string) (string, error) {
	today := time.Now()
	var d time.Time

	switch input {
	case ".":
		d = today
	case ".y":
		d = today.AddDate(0, 0, -1)
	case ".t":
		d = today.AddDate(0, 0, 1)
	default:
		parsed, err := time.Parse("2006-01-02", input)
		if err != nil {
			return "", fmt.Errorf("invalid date format, please use YYYY-MM-DD or . for today, .y for yesterday, .t for tomorrow")
		}
		d = parsed
	}
	return d.Format("2006-01-02"), nil
}

// SearchAccounts runs hledger accounts and lets user pick by index or type account
func SearchRecords(mode, searchTerm, file string) (string, error) {
	if noFileArg && config.Cfg.EfficientFileStructure.Enabled {
		file = efficientfiles.GetMainFile()
	}
    var cmdArgs []string
    if file != "" {
        cmdArgs = []string{mode, searchTerm, file}
    } else {
        cmdArgs = []string{mode, searchTerm}
    }

    out, err := exec.Command("hledger", cmdArgs...).CombinedOutput()
    if err != nil {
        fmt.Println("Error running hledger:", err)
        return "", err
    }

    results := strings.Split(strings.TrimSpace(string(out)), "\n")
    if len(results) == 0 || (len(results) == 1 && results[0] == "") {
        fmt.Println("No " + mode + " found.")
        return "", nil
    }

    // Print indexed list
    fmt.Println("Following " + mode + " found:")
    for i, r := range results {
        fmt.Printf("  %d) %s\n", i+1, r)
    }

    // Ask user to choose
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Select " + strings.TrimSuffix(mode, "s") + " (type index or full name): ")
    choice, _ := reader.ReadString('\n')
    choice = strings.TrimSpace(choice)

    // Try parsing as index
    num, err := strconv.Atoi(choice)
    if err == nil {
        if num >= 1 && num <= len(results) {
            return strings.TrimSpace(results[num-1]), nil
        }
    }

    // Otherwise use input as account name
    return choice, nil
}

// calculate balance amount
func calculateBalanceAmount(tx *Transaction) (string, error) {
	var total float64
	var currency string

	for _, l := range tx.Lines {
		if l.Type != LinePosting || l.Amount == "" {
			continue
		}
		parts := strings.Fields(l.Amount)
		if len(parts) < 2 {
			continue
		}
		val, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return "", fmt.Errorf("invalid number in %s", l.Amount)
		}
		if currency == "" {
			currency = parts[1]
		} else if parts[1] != currency {
			return "", fmt.Errorf("mixed currencies not supported for auto-balance")
		}
		total += val
	}
	if currency == "" {
		return "", fmt.Errorf("no amounts to balance")
	}
	return fmt.Sprintf("%g %s", -total, currency), nil
}

// creates postings for currency conversion transactions
func convertCurrencies(tx *Transaction, foreignAccount string) error {
	foreignAccount = strings.TrimPrefix(foreignAccount, "$")
	AskForeignAmount:
		foreignAmount := Ask("Amount?")
		foreignAmountValue,err := strconv.ParseFloat(strings.Split(foreignAmount, " ")[0], 64)
		if err != nil {
			fmt.Println("Invalid amount.")
			goto AskForeignAmount
		}
		foreignCurrency := strings.Split(foreignAmount, " ")[1]

	AskLocalAccount:
		localAccount := Ask("Account?")
		if localAccount == "" {
			fmt.Println("Local account must be specified.")
			goto AskLocalAccount
		}
		// Account search
		if strings.HasPrefix(localAccount, ".") {
			var searchTerm string
			if localAccount == "." {
				searchTerm = ""
			} else {
				searchTerm = localAccount[1:]
			}
			selected, err := SearchRecords("accounts",searchTerm, file)
			if err != nil {
				fmt.Println("Error searching accounts:", err)
				goto AskLocalAccount
			}
			if selected == "" {
				goto AskLocalAccount
			}
			localAccount = selected
		}
		// Comment
		if localAccount == ";" || localAccount == "#" {
			comment := Ask("Comment?")
			tx.Lines = append(tx.Lines, Line {
				Type:   LineComment,
				Text:   comment,
				Indent: localAccount == ";",
			})
			goto AskLocalAccount
		}

	AskLocalAmount:
		localAmount := Ask("Amount?")
		localAmountValue,err := strconv.ParseFloat(strings.Split(localAmount, " ")[0], 64)
		if err != nil {
			fmt.Println("Invalid amount.")
			goto AskLocalAmount
		}
		localCurrency := strings.Split(localAmount, " ")[1]
	
	// Local to foreign conversion
	if foreignAmountValue >= 0 {
		tx.Lines = append(tx.Lines, Line {
			Type:    LinePosting,
			Account: foreignAccount,
			Amount:  fmt.Sprintf("%s @@ %g %s", foreignAmount, localAmountValue*(-1), localCurrency),
		})
		tx.Lines = append(tx.Lines, Line {
			Type:    LinePosting,
			Account: localAccount,
			Amount:  localAmount,
		})
		tx.Lines = append(tx.Lines, Line {
			Type:    LinePosting,
			Account: config.Cfg.Accounts.ConversionAccount,
			Amount:  fmt.Sprintf("%g %s", foreignAmountValue*(-1), foreignCurrency),
		})
		tx.Lines = append(tx.Lines, Line {
			Type:    LinePosting,
			Account: config.Cfg.Accounts.ConversionAccount,
			Amount:  fmt.Sprintf("%g %s", localAmountValue*(-1), localCurrency),
		})
	} else {
		// Foreign to local conversion
		totalForeignBalance, totalForeignValue, err := getForeignBalance(foreignAccount, file)
		if err != nil {
			return err
		}

		var convertedForeignValue float64
		if -foreignAmountValue == totalForeignBalance {
			convertedForeignValue = totalForeignValue
		} else {
			wac := totalForeignValue / totalForeignBalance
			convertedForeignValue = (-foreignAmountValue) * wac
		}

		gainLoss := localAmountValue - convertedForeignValue
		gainLossAcc := config.Cfg.Accounts.FXLossAccount
		if gainLoss >= 0 {
			gainLossAcc = config.Cfg.Accounts.FXGainAccount
		}

		tx.Lines = append(tx.Lines, Line{
			Type:    LinePosting,
			Account: foreignAccount,
			Amount:  fmt.Sprintf("%s @@ %g %s", foreignAmount, convertedForeignValue, localCurrency),
		})
		tx.Lines = append(tx.Lines, Line{
			Type:    LinePosting,
			Account: localAccount,
			Amount:  localAmount,
		})
		tx.Lines = append(tx.Lines, Line{
			Type:    LinePosting,
			Account: gainLossAcc,
			Amount:  fmt.Sprintf("%g %s", -gainLoss, localCurrency),
		})
		tx.Lines = append(tx.Lines, Line{
			Type:    LinePosting,
			Account: config.Cfg.Accounts.ConversionAccount,
			Amount:  fmt.Sprintf("%g %s", -convertedForeignValue, localCurrency),
		})
		tx.Lines = append(tx.Lines, Line{
			Type:    LinePosting,
			Account: config.Cfg.Accounts.ConversionAccount,
			Amount:  fmt.Sprintf("%g %s", -foreignAmountValue, foreignCurrency),
		})
	}
	return nil
}

// getForeignBalance runs hledger and returns (balance, valueAtCost)
func getForeignBalance(account, file string) (float64, float64, error) {
	if noFileArg && config.Cfg.EfficientFileStructure.Enabled {
		file = efficientfiles.GetMainFile()
	}
	// hledger bal account --file file
	balCmd := exec.Command("hledger", "bal", account, "--file", file, "--no-total")
	balOut, err := balCmd.Output()
	if err != nil {
		fmt.Println("Error running hledger balance:", err)
		return 0, 0, err
	}
	balFields := strings.Fields(string(balOut))
	var balance float64
	if len(balFields) > 0 {
		balance, _ = strconv.ParseFloat(balFields[0], 64)
	} else {
		return 0,0,fmt.Errorf("can not calculate gain from zero balance. %s has no balance", account)
	}

	// hledger bal account --file file --value=then --cost
	valCmd := exec.Command("hledger", "bal", account, "--file", file, "--no-total", "--value=then", "--cost")
	valOut, err := valCmd.Output()
	if err != nil {
		fmt.Println("Error running hledger value:", err)
		return balance, 0, err
	}
	valFields := strings.Fields(string(valOut))
	var value float64
	if len(valFields) > 0 {
		value, _ = strconv.ParseFloat(valFields[0], 64)
	}

	return balance, value, nil
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&file, "file", "f", "", "Ledger file to write to")
}