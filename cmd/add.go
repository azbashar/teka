package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/azbashar/teka/internal/config"
	"github.com/azbashar/teka/internal/fileselector"
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

var currentFile string

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new transaction to your ledger",
	Run: func(cmd *cobra.Command, args []string) {
		fileArg = rootCmd.Flag("file").Value.String()
		mainFileArg = rootCmd.Flag("mainfile").Value.String()
		// Collect transaction data
		tx := Transaction{}

		// Date
	AskDate:
		date := Ask("Date?")
		if date == "" {
			fmt.Println("Abort.")
			return
		}

		// Comment
		if date == ";" || date == "#" {
			comment := Ask("Comment?")
			tx.Lines = append(tx.Lines, Line{
				Type:   LineComment,
				Text:   comment,
				Indent: false,
			})
			goto AskDate
		}

		// Parse date shortcuts
		date, err := ParseDate(date)
		if err != nil {
			fmt.Println(err)
			return
		}

		currentFile, err = fileselector.GetCurrentFile(date, fileArg)
		if err != nil {
			fmt.Println(err)
			return
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
			selected, err := SearchRecords("notes", searchTerm)
			if err != nil {
				fmt.Println("Error searching notes:", err)
				goto AskNote
			}
			if selected == "" {
				goto AskNote
			}
			note = selected
		}

		tx.Lines = append(tx.Lines, Line{
			Type: LineTransaction,
			Date: date,
			Note: note,
		})

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
				selected, err := SearchRecords("accounts", searchTerm)
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
			if strings.HasPrefix(account, "$") {
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
				tx.Lines = append(tx.Lines, Line{
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
					fmt.Println("Error balancing transaction:", err)
					goto AskAmount
				}
				amount = bal
			}

			tx.Lines = append(tx.Lines, Line{
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
		fmt.Printf("\nAdding this following transaction to %s:\n", currentFile)
		fmt.Printf("%s\n", content)

		// Confirm before writing
		if !Confirm("Is this correct") {
			fmt.Println("Transaction discarded.")
			return
		}

		// Store previous state for potential revert
		info, err := os.Stat(currentFile)
		prevSize := int64(0)
		if err == nil {
			prevSize = info.Size()
		}

		// Save transaction to file
		f, err := os.OpenFile(currentFile, os.O_APPEND|os.O_WRONLY, 0644)
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
		cmdArgs := []string{"check", "-f", currentFile}
		out, err := exec.Command("hledger", cmdArgs...).CombinedOutput()
		if err != nil {
			fmt.Println("Error validating ledger:")
			fmt.Println(string(out))

			if Confirm("Do you want to revert the changes?") {
				revertErr := os.Truncate(currentFile, prevSize)
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

func SearchRecords(mode, searchTerm string) (string, error) {
	mainFile, err := fileselector.GetMainFile(fileArg, mainFileArg)
	if err != nil {
		return "", err
	}

	cmdArgs := []string{mode, searchTerm, "-f", mainFile}

	out, err := exec.Command("hledger", cmdArgs...).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
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

// creates postings for currency conversion transactions
func convertCurrencies(tx *Transaction, foreignAccount string) error {
	foreignAccount = strings.TrimPrefix(foreignAccount, "$")
AskForeignAmount:
	foreignAmount := Ask("Amount?")
	foreignAmountValue, err := strconv.ParseFloat(strings.Split(foreignAmount, " ")[0], 64)
	if err != nil {
		fmt.Println("Invalid amount: ", err)
		goto AskForeignAmount
	}
	foreignCurrency := strings.Split(foreignAmount, " ")[1]

AskLocalAccount:
	localAccount := Ask("Account?")
	if localAccount == "" {
		fmt.Println("Local account must be specified when converting currencies.")
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
		selected, err := SearchRecords("accounts", searchTerm)
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
		tx.Lines = append(tx.Lines, Line{
			Type:   LineComment,
			Text:   comment,
			Indent: localAccount == ";",
		})
		goto AskLocalAccount
	}

AskLocalAmount:
	localAmount := Ask("Amount?")
	localAmountValue, err := strconv.ParseFloat(strings.Split(localAmount, " ")[0], 64)
	if err != nil {
		fmt.Println("Invalid amount:", err)
		goto AskLocalAmount
	}
	localCurrency := strings.Split(localAmount, " ")[1]

	// Local to foreign conversion
	if foreignAmountValue >= 0 {
		tx.Lines = append(tx.Lines, Line{
			Type:    LinePosting,
			Account: foreignAccount,
			Amount:  fmt.Sprintf("%s @@ %g %s", foreignAmount, localAmountValue*(-1), localCurrency),
		})
		tx.Lines = append(tx.Lines, Line{
			Type:    LinePosting,
			Account: localAccount,
			Amount:  localAmount,
		})
		tx.Lines = append(tx.Lines, Line{
			Type:    LinePosting,
			Account: config.Cfg.Accounts.ConversionAccount,
			Amount:  fmt.Sprintf("%g %s", foreignAmountValue*(-1), foreignCurrency),
		})
		tx.Lines = append(tx.Lines, Line{
			Type:    LinePosting,
			Account: config.Cfg.Accounts.ConversionAccount,
			Amount:  fmt.Sprintf("%g %s", localAmountValue*(-1), localCurrency),
		})
	} else {
		// Foreign to local conversion
		totalForeignBalance, totalForeignValue, err := getForeignBalance(foreignAccount)
		if err != nil {
			return err
		}

		var convertedForeignValue float64
		// convert full balance without rounding error
		if -foreignAmountValue == totalForeignBalance {
			convertedForeignValue = totalForeignValue
		} else { // partial amount conversion
			weightedAverageCost := totalForeignValue / totalForeignBalance
			convertedForeignValue = (-foreignAmountValue) * weightedAverageCost
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

func getForeignBalance(account string) (float64, float64, error) {
	mainFile, err := fileselector.GetMainFile(fileArg, mainFileArg)
	if err != nil {
		return 0, 0, err
	}

	// get balance in foreign currency
	// hledger bal account --file file
	balCmd := exec.Command("hledger", "bal", account, "-f", currentFile, "--no-total")
	balOut, err := balCmd.Output()
	if err != nil {
		return 0, 0, err
	}
	balFields := strings.Fields(string(balOut))
	var balance float64
	if len(balFields) > 0 {
		balance, _ = strconv.ParseFloat(balFields[0], 64)
	} else {
		fmt.Println(account, "has no balance.")
		return 0, 0, errors.New("can not calculate gain from zero balance")
	}

	// get value of foreign balance in local currency
	// hledger bal account --file file --value=then --cost
	valCmd := exec.Command("hledger", "bal", account, "-f", mainFile, "--no-total", "--value=then,"+config.Cfg.BaseCurrency, "--cost")
	valOut, err := valCmd.Output()
	if err != nil {
		return balance, 0, err
	}
	valFields := strings.Fields(string(valOut))
	var value float64
	if len(valFields) > 0 {
		value, _ = strconv.ParseFloat(valFields[0], 64)
	}

	return balance, value, nil
}

func calculateBalanceAmount(tx *Transaction) (string, error) {
	var total float64
	var currency string

	for _, l := range tx.Lines {
		if l.Type != LinePosting {
			continue
		}
		if l.Amount == "" {
			return "", errors.New("can not balance if postings are missing amount")
		}
		parts := strings.Fields(l.Amount)
		if len(parts) < 2 {
			return "", errors.New("invalid amount format. Only acceptable format is 1000.00 CUR")
		}
		val, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return "", errors.New("invalid number in " + l.Amount)
		}
		if currency == "" {
			currency = parts[1]
		}
		if parts[1] != currency {
			return "", errors.New("mixed currencies not supported for auto-balance")
		}
		total += val
	}
	if currency == "" {
		return "", errors.New("no amounts to balance")
	}
	return fmt.Sprintf("%g %s", -total, currency), nil
}

func init() {
	rootCmd.AddCommand(addCmd)
}
