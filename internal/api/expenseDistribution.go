package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/A-Bashar/Teka-Finance/internal/fileselector"
)

func expenseDistribution(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	// Prepare hledger command
	cmdArgs := []string{
		"is",           // income statement
		"--value=then", // value at period end
		"-O", "json",
	}
	if startDate != "" {
		cmdArgs = append(cmdArgs, "--begin", startDate)
	}
	if endDate != "" {
		cmdArgs = append(cmdArgs, "--end", endDate)
	}

	// Add files from fileselector
	files, expr, err := fileselector.GetRequiredFiles(startDate, endDate, fileArg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, f := range files {
		cmdArgs = append(cmdArgs, "-f", f)
	}
	if expr != "" {
		cmdArgs = append(cmdArgs, expr)
	}

	cmd := exec.Command("hledger", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		http.Error(w, "hledger error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var isData map[string]interface{}
	if err := json.Unmarshal(output, &isData); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse hledger output: %v", err), http.StatusInternalServerError)
		return
	}

	cbrSubreports, ok := isData["cbrSubreports"].([]interface{})
	if !ok {
		http.Error(w, "invalid cbrSubreports in hledger output", http.StatusInternalServerError)
		return
	}

	// Response struct with expenseData key
	type ExpenseReportResponse struct {
		Total       struct {
			Amount   float64 `json:"amount"`
			Currency string  `json:"currency"`
		} `json:"total"`
		ExpenseData []IncomeReportItem `json:"expenseData"`
	}

	var expenseItems []IncomeReportItem
	totalAmount := 0.0
	totalCurrency := "USD"

	// Iterate over subreports to find Expenses
	for _, sub := range cbrSubreports {
		subArr, ok := sub.([]interface{})
		if !ok || len(subArr) < 2 {
			continue
		}

		subName, ok := subArr[0].(string)
		if !ok {
			continue
		}

		// Accept "Expenses" or anything starting with "expenses"
		if !strings.EqualFold(subName, "Expenses") && !strings.HasPrefix(strings.ToLower(subName), "expenses") {
			continue
		}

		data, ok := subArr[1].(map[string]interface{})
		if !ok {
			continue
		}

		prRows, ok := data["prRows"].([]interface{})
		if !ok || len(prRows) == 0 {
			continue
		}

		for _, row := range prRows {
			rowMap, ok := row.(map[string]interface{})
			if !ok {
				continue
			}

			accountName, _ := rowMap["prrName"].(string)
			prrTotal, ok := rowMap["prrTotal"].([]interface{})
			if !ok || len(prrTotal) == 0 {
				continue
			}

			amountData, ok := prrTotal[0].(map[string]interface{})
			if !ok {
				continue
			}

			amount := 0.0
			currency := "USD"
			if aq, ok := amountData["aquantity"].(map[string]interface{}); ok {
				amount, _ = aq["floatingPoint"].(float64)
			}
			if comm, ok := amountData["acommodity"].(string); ok {
				currency = comm
			}

			// Set totalCurrency from first valid currency
			if totalCurrency == "USD" {
				totalCurrency = currency
			}

			totalAmount += amount
			expenseItems = append(expenseItems, IncomeReportItem{
				Account:  accountName,
				Amount:   amount,
				Currency: currency,
			})
		}

		break // stop after first matching subreport
	}

	response := ExpenseReportResponse{
		ExpenseData: expenseItems,
	}
	response.Total.Amount = totalAmount
	response.Total.Currency = totalCurrency

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
